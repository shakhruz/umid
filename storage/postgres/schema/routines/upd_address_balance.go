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

// UpdAddressBalance ...
const UpdAddressBalance = `
create or replace function upd_address_balance(address bytea,
                                               delta_value bigint,
                                               tx_time timestamptz,
                                               tx_height integer,
                                               comment text default null,
                                               type address_type default null)
    returns void
    language plpgsql
as
$$
declare
    adr_version integer := (get_byte(upd_address_balance.address, 0) << 8) + get_byte(upd_address_balance.address, 1);
    --
    cur_type    address_type;
    cur_value   bigint;
    cur_percent smallint;
    --
    new_value   bigint;
    new_type    address_type;
    new_percent smallint;
    --
    crt_tx      integer;
begin
	select b.confirmed_value, b.confirmed_percent, b.type
	into cur_value, cur_percent, cur_type
	from get_address_balance(upd_address_balance.address, upd_address_balance.tx_time, false) as b;    

	new_type := coalesce(upd_address_balance.type, cur_type); 
	new_value := cur_value + upd_address_balance.delta_value;
	new_percent := cur_percent;

	insert into address_balance_confirmed (address, version, value, percent, type, tx_height,
	                                       updated_at, created_at, created_tx_height)
	values (upd_address_balance.address, adr_version, new_value, cur_percent, new_type, upd_address_balance.tx_height,
	        upd_address_balance.tx_time, upd_address_balance.tx_time, upd_address_balance.tx_height)
	on conflict on constraint address_balance_confirmed_pkey do update set
		value = new_value,
		percent = cur_percent,
		type = new_type,
		tx_height = upd_address_balance.tx_height,
		updated_at = upd_address_balance.tx_time
	returning created_tx_height into crt_tx;

	-- пишем лог
    insert into address_balance_confirmed_log (address, version, value, percent, type, tx_height,
                                               updated_at, delta_value, comment)
    values (upd_address_balance.address, adr_version, new_value, new_percent, new_type, upd_address_balance.tx_height,
            upd_address_balance.tx_time, upd_address_balance.delta_value, upd_address_balance.comment);

	if crt_tx = upd_address_balance.tx_height
	then
		-- статистика
		insert into structure_stats (version, prefix, address_count, updates_at, tx_height)
		values (adr_version, convert_version_to_prefix(adr_version), 1, upd_address_balance.tx_time,
		        upd_address_balance.tx_height)
		on conflict on constraint structure_stats_pk do update
			set address_count = structure_stats.address_count + 1;
	end if;
end
$$;
`
