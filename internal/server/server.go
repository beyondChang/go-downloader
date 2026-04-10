package server

import (
	"crypto/subtle"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-downloader/internal/config"
	"github.com/go-downloader/internal/core"
	"github.com/go-downloader/internal/utils"
)

const ServerBindHost = "0.0.0.0"

// FindAvailablePort tries ports starting from 'start' until one is available
func FindAvailablePort(start int) (int, net.Listener) {
	bindHost := ServerBindHost
	for port := start; port < start+100; port++ {
		ln, err := net.Listen("tcp", fmt.Sprintf("%s:%d", bindHost, port))
		if err == nil {
			return port, ln
		}
	}
	return 0, nil
}

func BindServerListener(portFlag int) (int, net.Listener, error) {
	bindHost := ServerBindHost
	if portFlag > 0 {
		ln, err := net.Listen("tcp", fmt.Sprintf("%s:%d", bindHost, portFlag))
		if err != nil {
			return 0, nil, fmt.Errorf("could not bind to port %d: %w", portFlag, err)
		}
		return portFlag, ln, nil
	}
	port, ln := FindAvailablePort(1700)
	if ln == nil {
		return 0, nil, fmt.Errorf("could not find available port")
	}
	return port, ln, nil
}

// SaveActivePort writes the active port for local CLI and extension discovery.
func SaveActivePort(port int) {
	if err := os.MkdirAll(config.GetRuntimeDir(), 0o755); err != nil {
		utils.Debug("Error creating runtime directory for port file: %v", err)
		return
	}

	portFile := filepath.Join(config.GetRuntimeDir(), "port")
	if err := os.WriteFile(portFile, []byte(fmt.Sprintf("%d", port)), 0o644); err != nil {
		utils.Debug("Error writing port file: %v", err)
	}
	utils.Debug("HTTP server listening on port %d", port)
}

// RemoveActivePort cleans up the port file on exit
func RemoveActivePort() {
	portFile := filepath.Join(config.GetRuntimeDir(), "port")
	if err := os.Remove(portFile); err != nil && !os.IsNotExist(err) {
		utils.Debug("Error removing port file: %v", err)
	}
}

// GetActivePort reads the active port from the port file.
func GetActivePort() (int, error) {
	portFile := filepath.Join(config.GetRuntimeDir(), "port")
	data, err := os.ReadFile(portFile)
	if err != nil {
		return 0, err
	}
	var port int
	_, err = fmt.Sscanf(string(data), "%d", &port)
	if err != nil {
		return 0, err
	}
	return port, nil
}

// StartHTTPServer starts the HTTP server using an existing listener
func StartHTTPServer(ln net.Listener, port int, defaultOutputDir string, service core.DownloadService, tokenOverride string) {
	authToken := strings.TrimSpace(tokenOverride)
	if authToken == "" {
		authToken = ensureAuthToken()
	} else {
		persistAuthToken(authToken)
	}

	mux := http.NewServeMux()
	RegisterHTTPRoutes(mux, port, defaultOutputDir, service, authToken)

	// Wrap mux with Auth and CORS
	handler := corsMiddleware(authMiddleware(authToken, mux))

	server := &http.Server{Handler: handler}
	if err := server.Serve(ln); err != nil && err != http.ErrServerClosed {
		utils.Debug("HTTP server error: %v", err)
	}
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS, PUT, PATCH")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, Access-Control-Allow-Private-Network")
		w.Header().Set("Access-Control-Allow-Private-Network", "true")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func authMiddleware(token string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if token == "" {
			next.ServeHTTP(w, r)
			return
		}

		if r.URL.Path == "/" || r.URL.Path == "/index.html" || strings.HasPrefix(r.URL.Path, "/assets/") || r.URL.Path == "/health" {
			next.ServeHTTP(w, r)
			return
		}

		authHeader := r.Header.Get("Authorization")
		queryToken := r.URL.Query().Get("token")

		providedToken := ""
		if strings.HasPrefix(authHeader, "Bearer ") {
			providedToken = strings.TrimPrefix(authHeader, "Bearer ")
		} else if queryToken != "" {
			providedToken = queryToken
		}

		if subtle.ConstantTimeCompare([]byte(providedToken), []byte(token)) != 1 {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func ensureAuthToken() string {
	tokenFile := filepath.Join(config.GetStateDir(), "auth_token")
	data, err := os.ReadFile(tokenFile)
	if err == nil && len(data) > 0 {
		return strings.TrimSpace(string(data))
	}

	// 生成32字节的随机token
	token := utils.GenerateRandomToken(32)
	persistAuthToken(token)
	return token
}

func persistAuthToken(token string) {
	tokenFile := filepath.Join(config.GetStateDir(), "auth_token")
	if err := os.WriteFile(tokenFile, []byte(token), 0o600); err != nil {
		utils.Debug("Error persisting auth token: %v", err)
	}
}
