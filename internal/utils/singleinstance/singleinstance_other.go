//go:build !windows

package singleinstance

import (
	"os"
)

// Acquire 非 Windows 平台暂不实现单实例检测，始终返回 true
func Acquire(name string) bool {
	return true
}

// Release 非 Windows 平台无需释放
func Release() {
}
