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

// TruncateBlockchain ...
const TruncateBlockchain = `
create or replace function truncate_blockchain()
    returns void
    language plpgsql
as
$$
declare
    _sql text;
begin
    select into _sql string_agg(format('truncate table %s cascade;', tablename), E'\n')
    from pg_catalog.pg_tables
    where schemaname = 'public' and tablename not in ('level');
    if _sql is not null then
        execute _sql;
    end if;
    --
	perform setval('tx_height', 0, false);
    perform lo_unlink(lo.loid) from (select distinct loid from pg_largeobject) lo;
end
$$;
`
