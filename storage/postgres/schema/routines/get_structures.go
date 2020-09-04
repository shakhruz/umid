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

package routines

// GetStructures ...
const GetStructures = `
create or replace function get_structures()
    returns table
            (
                prefix            char(3),
                name              text,
                fee_percent       smallint,
                profit_percent    smallint,
                deposit_percent   smallint,
                fee_address       bytea,
                profit_address    bytea,
                master_address    bytea,
                transit_addresses bytea[],
                balance           bigint,
                address_count     int
            )
    language sql
as
$$

select t.prefix,
       t.name,
       t.fee_percent,
       t.profit_percent,
       (t.balance).percent,
       t.fee_address,
       t.profit_address,
       t.master_address,
       t.transit_addresses,
       (t.balance).value,
       t.address_count
from (
         select s.prefix,
                s.name,
                s.profit_percent,
                s.fee_percent,
                s.profit_address,
                s.fee_address,
                s.master_address,
                get_structure_balance(s.version) as balance,
                (select array_agg(address)
                 from structure_address
                 where version = s.version
                   and type = 'transit'
                   and deleted_at is null)       as transit_addresses,
                ss.address_count                 as address_count
         from structure_settings s
                  left join structure_stats ss on s.version = ss.version
     ) as t

$$;
`
