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
	"log"
	"sync"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
)

// BlockConfirmer ...
func BlockConfirmer(ctx context.Context, wg *sync.WaitGroup, conn *pgxpool.Pool) {
	wg.Add(1)
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Second):
			confirm(ctx, conn)
		}
	}
}

func confirm(ctx context.Context, conn *pgxpool.Pool) {
	var blkHeight int

	select {
	case <-ctx.Done():
		return
	default:
		break
	}

	err := conn.QueryRow(context.Background(), `select coalesce(confirm_next_block(), 0)`).Scan(&blkHeight)
	if err != nil {
		log.Println(err.Error())
	}

	if blkHeight != 0 {
		confirm(ctx, conn)
	}
}
