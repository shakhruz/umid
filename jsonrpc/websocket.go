// Copyright (c) 2020 UMI
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package jsonrpc

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const wsQueueLen = 16

// Client ...
type Client struct {
	ctx    context.Context
	cancel func()
	conn   *websocket.Conn
	res    chan []byte
	req    chan<- rawRequest
}

// NewClient ...
func NewClient(conn *websocket.Conn, req chan<- rawRequest) *Client {
	ctx, cancel := context.WithCancel(context.Background())

	return &Client{
		ctx:    ctx,
		cancel: cancel,
		conn:   conn,
		res:    make(chan []byte, wsQueueLen),
		req:    req,
	}
}

// WebSocket ...
func (rpc *RPC) WebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := rpc.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	cl := NewClient(conn, rpc.queue)
	go cl.reader()
	go cl.writer()
}

func (c *Client) reader() {
	for {
		msgType, msg, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}

			_ = c.conn.Close()
			c.cancel()

			break
		}

		if msgType == websocket.TextMessage {
			c.req <- rawRequest{c.ctx, msg, c.res}
		}
	}
}

func (c *Client) writer() {
	for {
		select {
		case data := <-c.res:
			err := c.conn.WriteMessage(websocket.TextMessage, data)
			if err != nil {
				log.Println(err.Error())

				_ = c.conn.Close()
				c.cancel()

				return
			}
		case <-c.ctx.Done():
			data := websocket.FormatCloseMessage(websocket.CloseServiceRestart, "")
			_ = c.conn.WriteControl(websocket.CloseMessage, data, time.Now().Add(time.Second))
			_ = c.conn.Close()

			return
		}
	}
}
