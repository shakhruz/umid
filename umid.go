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

package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"
	"umid/blockchain"
	"umid/jsonrpc"
	"umid/network"
	"umid/storage"
)

const (
	httpWriteTimeoutSec = 15
	httpIdleTimeoutSec  = 60
)

type server struct {
	srv *http.Server
}

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.Println("start")

	ctx, cancel := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}

	db := storage.NewStorage(ctx, wg)
	bc := blockchain.NewBlockchain(ctx, wg, db)
	rpc := jsonrpc.NewRPC(ctx, wg, bc)
	net := network.NewNetwork(ctx, wg, bc)

	go db.Worker()
	go bc.Worker()
	go rpc.Worker()
	go net.Worker()

	http.HandleFunc("/json-rpc", rpc.HTTP)
	http.HandleFunc("/json-rpc-ws", rpc.WebSocket)
	http.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	})

	srv := newServer()
	go srv.serve()

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt)
	<-sigint

	log.Println("prepare to shutdown")
	time.Sleep(time.Second)
	log.Println("shutting down")

	cancel()

	srv.shutdown()

	log.Println("wait group")
	wg.Wait()

	log.Println("stopped")
}

func newServer() server {
	addr := "127.0.0.1:8080"
	if val, ok := os.LookupEnv("BIND"); ok {
		addr = val
	}

	return server{
		srv: &http.Server{
			Addr: addr,

			ReadTimeout:       time.Second,
			WriteTimeout:      time.Second * httpWriteTimeoutSec,
			IdleTimeout:       time.Second * httpIdleTimeoutSec,
			ReadHeaderTimeout: time.Second,
		},
	}
}

func (s server) serve() {
	if err := s.srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Println(err.Error())
	}
}

func (s server) shutdown() {
	if err := s.srv.Shutdown(context.Background()); err != nil {
		log.Println(err.Error())
	}
}
