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
	"umid/umid"

	"github.com/jackc/pgx/v4"
)

func (s *postgres) Structures() ([]*umid.Structure2, error) {
	rows, err := s.conn.Query(context.Background(), `select * from get_structures()`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	sts := make([]*umid.Structure2, 0)

	for rows.Next() {
		st := &umid.Structure2{}

		err := rows.Scan(
			&st.Prefix, &st.Name, &st.FeePercent, &st.ProfitPercent, &st.DepositPercent, &st.FeeAddress,
			&st.ProfitAddress, &st.MasterAddress, &st.TransitAddresses, &st.Balance, &st.AddressCount,
		)
		if err != nil {
			return nil, err
		}

		sts = append(sts, st)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return sts, nil
}

func (s *postgres) StructureByPrefix(p string) (*umid.Structure2, error) {
	row := s.conn.QueryRow(context.Background(), `select * from get_structures_by_prefix($1)`, p)

	st := &umid.Structure2{}

	err := row.Scan(
		&st.Prefix, &st.Name, &st.FeePercent, &st.ProfitPercent, &st.DepositPercent, &st.FeeAddress,
		&st.ProfitAddress, &st.MasterAddress, &st.TransitAddresses, &st.Balance, &st.AddressCount,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			err = ErrNotFound
		}

		return nil, err
	}

	return st, nil
}
