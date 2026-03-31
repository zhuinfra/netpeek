package websocket

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

// TestWebSocket 测试WebSocket连接
func TestWebSocket(wsURL string) {
	// 验证URL格式
	if !strings.HasPrefix(wsURL, "ws://") && !strings.HasPrefix(wsURL, "wss://") {
		fmt.Printf("错误: 无效的WebSocket URL格式。必须以 ws:// 或 wss:// 开头\n")
		os.Exit(1)
	}
	
	fmt.Printf("正在测试 WebSocket 连接: %s\n", wsURL)
	fmt.Println("----------------------------------")
	
	// 设置连接参数
	dialer := &websocket.Dialer{
		HandshakeTimeout: 5 * time.Second,
	}
	
	// 记录开始时间
	startTime := time.Now()
	
	// 建立WebSocket连接
	conn, resp, err := dialer.Dial(wsURL, nil)
	if err != nil {
		fmt.Printf("连接失败: %v\n", err)
		if resp != nil {
			fmt.Printf("HTTP状态码: %d %s\n", resp.StatusCode, resp.Status)
		}
		os.Exit(1)
	}
	defer conn.Close()
	
	// 计算连接时间
	connectTime := time.Since(startTime)
	fmt.Printf("连接成功！连接耗时: %v\n", connectTime)
	
	// 发送测试消息
	testMessage := "NetPeek WebSocket Test"
	startSendTime := time.Now()
	
	err = conn.WriteMessage(websocket.TextMessage, []byte(testMessage))
	if err != nil {
		fmt.Printf("发送消息失败: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Printf("发送消息: %s\n", testMessage)
	
	// 设置读取超时
	err = conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	if err != nil {
		fmt.Printf("设置读取超时失败: %v\n", err)
		os.Exit(1)
	}
	
	// 接收响应
	messageType, response, err := conn.ReadMessage()
	if err != nil {
		fmt.Printf("接收消息失败: %v\n", err)
		if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
			fmt.Printf("意外关闭: %v\n", err)
		}
		// 不退出，继续显示其他信息
	} else {
		// 计算响应时间
		responseTime := time.Since(startSendTime)
		fmt.Printf("接收响应: %s\n", string(response))
		fmt.Printf("响应类型: %s\n", getWebSocketMessageType(messageType))
		fmt.Printf("响应耗时: %v\n", responseTime)
	}
	
	// 发送关闭消息
	err = conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		fmt.Printf("发送关闭消息失败: %v\n", err)
	}
	
	// 等待关闭确认
	time.Sleep(100 * time.Millisecond)
	
	fmt.Println("----------------------------------")
	fmt.Println("WebSocket测试完成")
}

// getWebSocketMessageType 获取WebSocket消息类型的字符串表示
func getWebSocketMessageType(messageType int) string {
	switch messageType {
	case websocket.TextMessage:
		return "文本消息"
	case websocket.BinaryMessage:
		return "二进制消息"
	case websocket.CloseMessage:
		return "关闭消息"
	case websocket.PingMessage:
		return "Ping消息"
	case websocket.PongMessage:
		return "Pong消息"
	default:
		return fmt.Sprintf("未知类型(%d)", messageType)
	}
}
