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

// UpdStructureBalance ...
const UpdStructureBalance = `
create or replace function upd_structure_balance(version integer,
                                                 delta_value bigint,
                                                 epoch timestamptz,
                                                 tx_height integer,
                                                 comment text default null)
    returns void
    language plpgsql
as
$$
declare
    str_prefix  char(3);
    cur_value   bigint;
    cur_percent smallint;
    new_value   bigint;
    new_percent smallint;
begin
    select value, percent
    into cur_value, cur_percent
    from get_structure_balance(upd_structure_balance.version, upd_structure_balance.epoch);
    --
    new_value := cur_value + upd_structure_balance.delta_value;
    new_percent := cur_percent;
    --
    insert into structure_balance(version, prefix, value, percent, tx_height, updated_at)
    values (upd_structure_balance.version,
            convert_version_to_prefix(upd_structure_balance.version),
            new_value,
            new_percent,
            upd_structure_balance.tx_height,
            upd_structure_balance.epoch)
    on conflict on constraint structure_balance_pk
        do update set
            value      = new_value,
            percent    = new_percent,
            tx_height  = upd_structure_balance.tx_height,
            updated_at = upd_structure_balance.epoch
    returning prefix into str_prefix;
    --
    insert into structure_balance_log (version, prefix, value, percent, tx_height, updated_at, comment)
    values (upd_structure_balance.version, str_prefix, new_value, new_percent, upd_structure_balance.tx_height,
            upd_structure_balance.epoch, upd_structure_balance.comment);
end
$$;
`
