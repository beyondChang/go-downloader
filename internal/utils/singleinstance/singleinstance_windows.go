//go:build windows

package singleinstance

import (
	"fmt"

	"golang.org/x/sys/windows"
)

var handle windows.Handle

// Acquire 尝试获取命名互斥体，确保只有一个实例运行
// 如果已有实例在运行则返回 false
func Acquire(name string) bool {
	mutexName, _ := windows.UTF16PtrFromString(fmt.Sprintf("Global\\%s", name))
	handle, err := windows.CreateMutex(nil, false, mutexName)
	if err != nil {
		return false
	}
	if windows.GetLastError() == windows.ERROR_ALREADY_EXISTS {
		windows.CloseHandle(handle)
		return false
	}
	return true
}

// Release 释放互斥体
func Release() {
	if handle != 0 {
		windows.CloseHandle(handle)
	}
}
