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
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

const wsQueueLen = 16

// WebSocket ...
func (rpc *RPC) WebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := rpc.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	res := make(chan []byte, wsQueueLen)

	go rpc.reader(ctx, conn, res, cancel)
	go rpc.writer(ctx, conn, res)
}

func (rpc *RPC) reader(ctx context.Context, conn *websocket.Conn, res chan<- []byte, done func()) {
	rpc.wg.Add(1)
	defer rpc.wg.Done()

	for {
		t, req, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}

			if err = conn.Close(); err != nil && !strings.Contains(err.Error(), "use of closed network connection") {
				log.Println("reader conn.Close()", err.Error())
			}

			done()

			break
		}

		if t == websocket.TextMessage {
			rpc.queue <- rawRequest{ctx, req, res}
		}
	}
}

func (rpc *RPC) writer(ctx context.Context, conn *websocket.Conn, res <-chan []byte) {
	rpc.wg.Add(1)
	defer rpc.wg.Done()

	for {
		select {
		case data := <-res:
			if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
				log.Println(err.Error())
			}
		case <-ctx.Done():
			log.Println("local context done")

			return
		case <-rpc.ctx.Done():
			log.Println("ws writer shutdown")

			data := websocket.FormatCloseMessage(websocket.CloseServiceRestart, "")
			if err := conn.WriteControl(websocket.CloseMessage, data, time.Now().Add(time.Second)); err != nil {
				log.Println("ccccc", err.Error())
			}

			if err := conn.Close(); err != nil && !strings.Contains(err.Error(), "use of closed network connection") {
				log.Println("writer conn.Close()", err.Error())
			}

			return
		}
	}
}
