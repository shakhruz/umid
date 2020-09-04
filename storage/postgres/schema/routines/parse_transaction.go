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

// ParseTransaction ...
const ParseTransaction = `
create or replace function parse_transaction(bytes bytea,
                                             out hash bytea,
                                             out version smallint,
                                             out sender bytea,
                                             out recipient bytea,
                                             out value bigint,
                                             out prefix text,
                                             out name text,
                                             out profit_percent integer,
                                             out fee_percent integer)
    language plpgsql
as
$$
declare
    trx_length          constant smallint := 150;
    ver_genesis         constant smallint := 0;
    ver_basic           constant smallint := 1;
    ver_add_struct      constant smallint := 2;
    ver_upd_struct      constant smallint := 3;
    ver_upd_profit_adr  constant smallint := 4;
    ver_upd_fee_adr     constant smallint := 5;
    ver_add_transit_adr constant smallint := 6;
    ver_del_transit_adr constant smallint := 7;

begin
    if length(bytes) != trx_length then
        raise exception 'transaction size should be % bytes, got %', trx_length, length(bytes);
    end if;

	hash := sha256(bytes);
    version := get_byte(bytes, 0);
    sender := substr(bytes, 2, 34);

    if version in (ver_genesis, ver_basic) then 
        value := (get_byte(bytes, 69)::bigint << 56) + (get_byte(bytes, 70)::bigint << 48) +
                 (get_byte(bytes, 71)::bigint << 40) + (get_byte(bytes, 72)::bigint << 32) +
                 (get_byte(bytes, 73)::bigint << 24) + (get_byte(bytes, 74)::bigint << 16) +
                 (get_byte(bytes, 75)::bigint << 8) + (get_byte(bytes, 76)::bigint);
    end if;

    if version in (ver_add_struct, ver_upd_struct) then
        prefix := convert_version_to_prefix((get_byte(bytes, 35) << 8) + get_byte(bytes, 36));
        name := convert_from(substr(bytes, 43, get_byte(bytes, 41)), 'UTF-8');
        profit_percent := (get_byte(bytes, 37) << 8) + get_byte(bytes, 38);
        fee_percent := (get_byte(bytes, 39) << 8) + get_byte(bytes, 40);
    else
        recipient := substr(bytes, 36, 34);
    end if;

    if version in (ver_upd_profit_adr, ver_upd_fee_adr, ver_add_transit_adr, ver_del_transit_adr) then
        prefix := convert_version_to_prefix((get_byte(bytes, 35) << 8) + get_byte(bytes, 36));
    end if;
end
$$;
`
