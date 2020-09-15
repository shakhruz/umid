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

	"github.com/jackc/pgx/v4"
)

// Mempool ...
func (s *Postgres) Mempool(ctx context.Context) (<-chan []byte, error) {
	ctz := context.Background()

	tx, err := s.conn.Begin(ctz)
	if err != nil {
		return nil, err
	}

	_, err = tx.Exec(ctz, `declare cur no scroll cursor for select raw from mempool order by priority`)
	if err != nil {
		_ = tx.Rollback(ctz)

		return nil, err
	}

	c := make(chan []byte)

	go fetcher(ctx, tx, c)

	return c, nil
}

func fetcher(ctx context.Context, tx pgx.Tx, c chan<- []byte) {
	ctz := context.Background()

L:
	for {
		raw := make([]byte, 150)
		row := tx.QueryRow(ctz, `fetch next from cur`)

		if err := row.Scan(&raw); err != nil {
			break L
		}

		select {
		case c <- raw:
			continue
		case <-ctx.Done():
			break L
		}
	}

	close(c)

	_ = tx.Rollback(ctz)
}
