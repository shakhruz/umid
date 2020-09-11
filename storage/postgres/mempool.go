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
	"umid/umid"

	"github.com/jackc/pgx/v4"
)

type mempool struct {
	tx  pgx.Tx
	val []byte
}

// Mempool ...
func (s *postgres) Mempool() (mem umid.IMempool, err error) {
	ctx := context.Background()

	tx, err := s.conn.Begin(ctx)
	if err != nil {
		return nil, err
	}

	_, err = tx.Exec(ctx, `declare cur no scroll cursor for select raw from mempool order by priority for update`)
	if err != nil {
		_ = tx.Rollback(ctx)

		return nil, err
	}

	mem = &mempool{tx, []byte{}}

	return mem, nil
}

func (m *mempool) Next() bool {
	ctx := context.Background()
	row := m.tx.QueryRow(ctx, `fetch next from cur`)

	if err := row.Scan(&m.val); err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			_ = m.tx.Rollback(ctx)
		}

		return false
	}

	return true
}

func (m *mempool) Value() []byte {
	return m.val
}

func (m *mempool) Close() {
	if err := m.tx.Commit(context.Background()); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
		log.Println(err.Error())
	}
}
