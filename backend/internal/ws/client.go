package ws

import "github.com/gorilla/websocket"


// Client: 一个在线用户的连接
type Client struct {
	UserID uint64
	Conn   *websocket.Conn
	Send   chan []byte
}