package handler

import (
	"gochat/internal/ws"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func WSHandler(hub *ws.Hub, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		userId := c.GetUint64("user_id") // JWT 中间件解析

		// 验证用户ID有效性
		if userId == 0 {
			logger.Warn("Invalid user ID from WebSocket connection")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user"})
			return
		}

		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			logger.Error("WebSocket upgrade failed", zap.Error(err))
			return
		}

		// 设置连接参数
		conn.SetReadLimit(512 * 1024) // 512KB
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		conn.SetPongHandler(func(string) error {
			conn.SetReadDeadline(time.Now().Add(60 * time.Second))
			return nil
		})

		client := &ws.Client{
			UserID: userId,
			Conn:   conn,
			Send:   make(chan []byte, 256), // 缓冲通道防止阻塞
		}

		logger.Info("WebSocket client connected", zap.Uint64("user_id", userId))

		hub.Register <- client

		go readLoop(hub, client, logger)
		go writeLoop(client, logger)
	}
}

func readLoop(hub *ws.Hub, client *ws.Client, logger *zap.Logger) {
	defer func() {
		hub.Unregister <- client
		client.Conn.Close()
		logger.Info("WebSocket read loop closed", zap.Uint64("user_id", client.UserID))
	}()

	for {
		var msg ws.WSMessage
		if err := client.Conn.ReadJSON(&msg); err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logger.Warn("WebSocket read error", 
					zap.Uint64("user_id", client.UserID), 
					zap.Error(err))
			}
			break
		}
		
		msg.FromID = client.UserID
		if msg.Type == "heartbeat" {
			client.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
			continue
		}

		ws.RouteMessage(hub, &msg)
	}
}

func writeLoop(client *ws.Client, logger *zap.Logger) {
	ticker := time.NewTicker(30 * time.Second) // 心跳定时器
	defer func() {
		ticker.Stop()
		client.Conn.Close()
		logger.Info("WebSocket write loop closed", zap.Uint64("user_id", client.UserID))
	}()

	for {
		select {
		case msg, ok := <-client.Send:
			if !ok {
				// 通道已关闭
				client.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			
			client.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := client.Conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				logger.Warn("WebSocket write error", 
					zap.Uint64("user_id", client.UserID), 
					zap.Error(err))
				return
			}
			
		case <-ticker.C:
			// 发送心跳
			client.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := client.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
