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

package network

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

const (
	httpWriteTimeoutSec = 15
	httpIdleTimeoutSec  = 60
)

// Server ...
type Server struct {
	Readiness bool
	http      *http.Server
}

// NewServer ...
func NewServer() *Server {
	addr := "127.0.0.1:8080"
	if val, ok := os.LookupEnv("LISTEN"); ok {
		addr = val
	}

	srv := &Server{
		Readiness: true,
		http: &http.Server{
			Addr:              addr,
			ReadTimeout:       time.Second,
			WriteTimeout:      time.Second * httpWriteTimeoutSec,
			IdleTimeout:       time.Second * httpIdleTimeoutSec,
			ReadHeaderTimeout: time.Second,
		},
	}

	http.Handle("/healthz", srv)

	return srv
}

// Serve ...
func (s *Server) Serve() {
	if err := s.http.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Println(err.Error())
	}
}

// Shutdown ...
func (s *Server) Shutdown() {
	log.Println("shut down")

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	if err := s.http.Shutdown(ctx); err != nil {
		log.Println(err.Error())
	}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	if !s.Readiness {
		w.WriteHeader(http.StatusServiceUnavailable)
	}
}

// DrainConnections ...
func (s *Server) DrainConnections() {
	s.Readiness = false
	timeout := 1

	if val, ok := os.LookupEnv("DRAINING_TIMEOUT"); ok {
		if i, err := strconv.Atoi(val); err == nil {
			timeout = i
		}
	}

	log.Printf("drain connections (%d sec)\n", timeout)
	time.Sleep(time.Duration(timeout) * time.Second)
}
