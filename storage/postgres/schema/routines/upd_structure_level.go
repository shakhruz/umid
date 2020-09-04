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

// UpdStructureLevel ...
const UpdStructureLevel = `
create or replace function upd_structure_level(block_height integer,
                                               block_time timestamptz,
                                               comment text default null)
    returns void
    language plpgsql
as
$$
    # variable_conflict use_column
declare
    rec         record;
    balance     bigint;
    cur_level   bigint;
    cur_percent smallint;
    upd_ver     integer;
begin
    --
    for rec in
        select version, prefix, profit_percent from structure_settings
        loop
            select value into balance from get_structure_balance(rec.version, block_time);
            --
            if balance is null
			then
				raise '% %', rec.version, rec.prefix;
			end if;
            --
            select id, percent
            into cur_level, cur_percent
            from level
            where balance between min_value and max_value
            limit 1;
            -- 
            update structure_percent
            set level           = cur_level,
                percent         = cur_percent,
                dev_percent     = case cur_level when 0 then 0 else cur_percent + 200 end,  -- 2%
                profit_percent  = cur_percent,
                deposit_percent = case cur_level when 0 then 0 else cur_percent - rec.profit_percent end,
                block_height    = upd_structure_level.block_height,
                updated_at      = upd_structure_level.block_time
            where version = rec.version
              and level <> cur_level
            returning version into upd_ver;
            --
            if upd_ver is not null
            then
                insert into structure_percent_log (version, prefix, level, percent, dev_percent, profit_percent, deposit_percent, block_height, updated_at, comment)
                values (
                        rec.version,  -- version
                        rec.prefix,   -- prefix
                        cur_level,    -- level
                        cur_percent,  -- percent
                        case cur_level when 0 then 0 else cur_percent + 200 end, -- dev_percent
                        cur_percent,  -- profit_percent
                        case cur_level when 0 then 0 else cur_percent - rec.profit_percent end, -- deposit_percent
                        upd_structure_level.block_height, -- block_height
                        upd_structure_level.block_time, -- updated_at
                        'level update' -- comment
                        );
            end if;
        end loop;
end
$$;
`
