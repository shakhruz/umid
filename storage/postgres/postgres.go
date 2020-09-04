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

package postgres

import (
	"context"
	"errors"
	"log"
	"os"
	"sync"
	"umid/umid"

	"github.com/jackc/pgx/v4/pgxpool"
)

// ErrNotFound ...
var ErrNotFound = errors.New("not found")

type postgres struct {
	ctx  context.Context
	wg   *sync.WaitGroup
	conn *pgxpool.Pool
}

// NewStorage ...
func NewStorage(ctx context.Context, wg *sync.WaitGroup) umid.IStorage {
	cfg, err := pgxpool.ParseConfig(os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err.Error())
	}

	cfg.LazyConnect = true

	conn, err := pgxpool.ConnectConfig(ctx, cfg)
	if err != nil {
		log.Fatal(err.Error())
	}

	return &postgres{ctx, wg, conn}
}
