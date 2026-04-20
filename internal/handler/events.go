// Package handler 实时事件处理器
// 提供 Server-Sent Events (SSE) 用于实时推送下载进度
package handler

import (
	"fmt"
	"net/http"
	"time"
)

// Events SSE 事件流处理器
// 返回实时下载进度更新
func (h *Handler) Events(w http.ResponseWriter, r *http.Request) {
	// 设置 SSE 响应头
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// 获取响应写入器的刷新器
	flusher, ok := w.(http.Flusher)
	if !ok {
		writeError(w, http.StatusInternalServerError, "不支持 SSE")
		return
	}

	// 发送初始连接消息
	fmt.Fprintf(w, "data: {\"type\":\"connected\"}\n\n")
	flusher.Flush()

	// 创建定时器，定期发送心跳
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	// 监听客户端断开连接
	ctx := r.Context()

	for {
		select {
		case <-ctx.Done():
			// 客户端断开连接
			return
			
		case <-ticker.C:
			// 发送心跳包
			fmt.Fprintf(w, "data: {\"type\":\"heartbeat\"}\n\n")
			flusher.Flush()
			
		// TODO: 从下载服务接收事件并推送
		// case event := <-h.eventChan:
		//     fmt.Fprintf(w, "data: %s\n\n", event)
		//     flusher.Flush()
		}
	}
}
