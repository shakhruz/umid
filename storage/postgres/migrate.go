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
	"log"
	"time"
	"umid/storage/postgres/schema"
	"umid/storage/postgres/schema/sequences"

	"github.com/jackc/pgx/v4"
)

func (s *postgres) Migrate() {
	s.wg.Add(1)
	defer s.wg.Done()

	// retry if database is not available
	for {
		select {
		case <-s.ctx.Done():
			return
		default:
			if s.doMigrate() {
				return
			}

			time.Sleep(time.Second)
		}
	}
}

func (s *postgres) doMigrate() bool {
	tx, err := s.conn.Begin(s.ctx)
	if err != nil {
		log.Println(err.Error())

		return false
	}

	defer func() { _ = tx.Rollback(s.ctx) }()

	cur, _ := s.currentVersion(tx)

	var ok bool

	for v, m := range schema.Migrations() {
		if v > cur {
			for n, sql := range m {
				log.Printf("db migration: %d.%d\n", v, n)

				_, err := tx.Exec(s.ctx, sql)
				if err != nil {
					log.Fatal(err.Error())
				}
			}

			if _, err = tx.Exec(s.ctx, `select setval('db_version', $1, false)`, v); err != nil {
				log.Fatalln(err.Error())
			}

			ok = true
		}
	}

	if ok {
		if err := tx.Commit(s.ctx); err != nil {
			log.Println(err.Error())

			return false
		}

		log.Println("db migration: done")
	}

	return true
}

func (s *postgres) currentVersion(tx pgx.Tx) (v int, err error) {
	if _, err = tx.Exec(s.ctx, sequences.DBVersion); err == nil {
		err = tx.QueryRow(s.ctx, `select setval('db_version', nextval('db_version'), false)`).Scan(&v)
	}

	return v, err
}
