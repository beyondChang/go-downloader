package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var (
	infoLog  *log.Logger // 信息日志实例
	errorLog *log.Logger // 错误日志实例
	file     *os.File    // 当前日志文件句柄
	mu       sync.Mutex  // 文件写入锁
	current  string      // 当前日志文件日期，格式 "2006-01-02"
)

// Init 初始化日志系统，同时输出到控制台和按天分割的日志文件
func Init(dataDir string) {
	logDir := filepath.Join(dataDir, "logs")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		log.Fatalf("创建日志目录失败: %v", err)
	}

	// 打开当日日志文件
	openFile := func() {
		today := time.Now().Format("2006-01-02")
		path := filepath.Join(logDir, today+".log")
		f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatalf("打开日志文件失败: %v", err)
		}
		file = f
		current = today
	}

	openFile()

	// 创建同时写入控制台和文件的日志实例
	infoLog = log.New(io.MultiWriter(os.Stdout, &rotatingWriter{logDir: logDir}), "[INFO] ", log.Ldate|log.Ltime|log.Lshortfile)
	errorLog = log.New(io.MultiWriter(os.Stderr, &rotatingWriter{logDir: logDir}), "[ERROR] ", log.Ldate|log.Ltime|log.Lshortfile)
}

// Close 关闭日志文件
func Close() {
	if file != nil {
		file.Close()
	}
}

// rotatingWriter 带日期检查的文件写入器，跨天自动切换日志文件
type rotatingWriter struct {
	logDir string
}

// Write 实现 io.Writer 接口，每次写入时检查日期是否变更，跨天自动切换文件
func (w *rotatingWriter) Write(p []byte) (n int, err error) {
	mu.Lock()
	defer mu.Unlock()

	today := time.Now().Format("2006-01-02")
	if today != current {
		// 日期变更，关闭旧文件，打开新文件
		if file != nil {
			file.Close()
		}
		path := filepath.Join(w.logDir, today+".log")
		file, err = os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return 0, err
		}
		current = today
	}

	return file.Write(p)
}

// Info 输出信息级别日志
func Info(format string, v ...interface{}) {
	if infoLog != nil {
		infoLog.Output(2, fmt.Sprintf(format, v...))
	}
}

// Error 输出错误级别日志
func Error(format string, v ...interface{}) {
	if errorLog != nil {
		errorLog.Output(2, fmt.Sprintf(format, v...))
	}
}

// Fatal 输出错误级别日志并退出程序
func Fatal(format string, v ...interface{}) {
	if errorLog != nil {
		errorLog.Output(2, fmt.Sprintf(format, v...))
	}
	os.Exit(1)
}
