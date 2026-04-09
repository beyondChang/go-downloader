package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/go-downloader/internal/app"
	"github.com/go-downloader/internal/server"
	"github.com/go-downloader/internal/tray"
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
		// 1. Acquire lock
		releaseLock, err := app.AcquireLock()
		if err != nil {
			return err
		}
		defer func() {
			if releaseLock {
				app.ReleaseLock()
			}
		}()

		// 2. Initialize app
		app.Initialize()

		// 3. Start server
		portFlag, _ := cmd.Flags().GetInt("port")
		outputDir, _ := cmd.Flags().GetString("output")
		batchFile, _ := cmd.Flags().GetString("batch")

		port, ln, err := server.BindServerListener(portFlag)
		if err != nil {
			return err
		}
		defer ln.Close()

		server.SaveActivePort(port)
		defer server.RemoveActivePort()

		go server.StartHTTPServer(ln, port, outputDir, app.GlobalService, globalToken)

		// 4. Initial downloads
		var urls []string
		if batchFile != "" {
			data, err := os.ReadFile(batchFile)
			if err == nil {
				for _, line := range strings.Split(string(data), "\n") {
					if trimmed := strings.TrimSpace(line); trimmed != "" {
						urls = append(urls, trimmed)
					}
				}
			}
		}
		urls = append(urls, args...)
		if len(urls) > 0 {
			app.ProcessDownloads(urls, outputDir)
		}

		// 5. Start tray (blocks until quit)
		url := fmt.Sprintf("http://127.0.0.1:%d", port)
		fmt.Printf("\n  🚀 Downloader Web 已启动!\n")
		fmt.Printf("  🔗 访问地址: %s\n", url)
		fmt.Printf("  💡 已最小化到系统托盘\n\n")

		tray.Run(url)
		return nil
	},
}

func Execute() error {
	return rootCmd.Execute()
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
