package cmd

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/go-downloader/internal/app"
	"github.com/go-downloader/internal/server"
	"github.com/go-downloader/internal/tray"
	"github.com/go-downloader/internal/utils"
	"github.com/spf13/cobra"
)

var Version = "0.0.1"

var (
	globalToken string
)

var rootCmd = &cobra.Command{
	Use:   "downloader [url]...",
	Short: "Downloader is a powerful and easy-to-use download manager",
	Long: `Downloader supports multi-threaded downloads, pausing/resuming,
and categorization. It provides a modern Web interface for easy management.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		utils.Debug("rootCmd.RunE invoked with args: %v", args)
		// 1. Acquire lock
		utils.Debug("Attempting to acquire instance lock")
		releaseLock, err := app.AcquireLock()
		if err != nil {
			utils.Debug("Lock acquisition failed: %v", err)
			return nil
		}
		defer func() {
			if releaseLock {
				utils.Debug("Releasing instance lock")
				app.ReleaseLock()
			}
		}()

		// 2. Initialize app
		utils.Debug("Initializing application state")
		app.Initialize()
		utils.Debug("Application state initialized")

		// 3. Start server
		portFlag, _ := cmd.Flags().GetInt("port")
		outputDir, _ := cmd.Flags().GetString("output")
		batchFile, _ := cmd.Flags().GetString("batch")

		utils.Debug("Binding server listener on port flag: %d", portFlag)
		port, ln, err := server.BindServerListener(portFlag)
		if err != nil {
			utils.Debug("Failed to bind server listener: %v", err)
			utils.Notify("启动失败", fmt.Sprintf("无法绑定端口: %v", err))
			return err
		}
		defer ln.Close()

		utils.Debug("Server bound to port %d", port)
		server.SaveActivePort(port)
		defer server.RemoveActivePort()

		utils.Debug("Starting HTTP server")
		go server.StartHTTPServer(ln, port, outputDir, app.GlobalService, globalToken)

		// 4. Initial downloads
		var urls []string
		if batchFile != "" {
			utils.Debug("Reading URLs from batch file: %s", batchFile)
			data, err := os.ReadFile(batchFile)
			if err == nil {
				for _, line := range strings.Split(string(data), "\n") {
					if trimmed := strings.TrimSpace(line); trimmed != "" {
						urls = append(urls, trimmed)
					}
				}
			} else {
				utils.Debug("Failed to read batch file: %v", err)
			}
		}
		urls = append(urls, args...)
		if len(urls) > 0 {
			utils.Debug("Processing %d initial download(s)", len(urls))
			app.ProcessDownloads(urls, outputDir)
		}

		// 5. Start tray (blocks until quit)
		url := fmt.Sprintf("http://127.0.0.1:%d", port)

		utils.Debug("Starting system tray loop")
		utils.Notify("Downloader 已启动", fmt.Sprintf("访问地址: %s", url))
		tray.Run(url)
		utils.Debug("System tray loop exited")
		return nil
	},
}

func Execute() error {
	utils.Debug("cmd.Execute() invoked")
	// Always discard standard outputs for Cobra's internal help/usage messages
	// in case of errors during double-click or non-terminal start.
	rootCmd.SetOut(io.Discard)
	rootCmd.SetErr(io.Discard)

	utils.Debug("Running rootCmd.Execute()")
	return rootCmd.Execute()
}

// RunGUI directly invokes the RunE logic without Cobra parsing,
// useful for Windows GUI mode where Cobra's internal stdout handling might crash.
func RunGUI() error {
	utils.Debug("cmd.RunGUI() invoked")
	return rootCmd.RunE(rootCmd, nil)
}

func init() {
	rootCmd.PersistentFlags().StringVar(&globalToken, "token", "", "Set custom API auth token (or use SURGE_TOKEN environment variable)")
	rootCmd.Flags().StringP("batch", "b", "", "File containing URLs to download (one per line)")
	rootCmd.Flags().IntP("port", "p", 0, "Port to listen on (default: 8080 or first available)")
	rootCmd.Flags().StringP("output", "o", "", "Output directory (defaults to current working directory)")
	rootCmd.Flags().Bool("no-resume", false, "Do not auto-resume paused downloads on startup")
	rootCmd.Flags().Bool("exit-when-done", false, "Exit when all downloads complete")
	rootCmd.SetVersionTemplate("Downloader v{{.Version}}\n")
	rootCmd.Version = Version
}
