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

// ConfirmTxUpdProfitAddress ...
const ConfirmTxUpdProfitAddress = `
create or replace function confirm_tx__upd_profit_address(bytes bytea,
                                              tx_height integer,
                                              blk_height integer,
                                              blk_tx_idx integer,
                                              blk_time timestamptz)
    returns void
    language plpgsql
as
$$
declare
    tx_hash      bytea;
    tx_version   smallint;
    tx_sender    bytea;
    tx_recipient bytea;
    tx_struct    jsonb := '{}'::jsonb;
    st_version   integer;
    st_prefix    text;
	--
    dev_adr      bytea;
    mst_adr       bytea;
    profit_adr   bytea;
    fee_adr      bytea;
    old_balance  bigint;
    new_balance  bigint;
    str_balance  bigint;
    --
    st_name       text;
    st_profit_prc smallint;
    st_fee_prc    smallint;
begin
    select hash, version, sender, recipient, prefix
    into tx_hash, tx_version, tx_sender, tx_recipient, st_prefix
    from parse_transaction(bytes);

    st_version := convert_prefix_to_version(st_prefix);
    tx_struct := jsonb_set(tx_struct, '{prefix}', to_jsonb(st_prefix));

    --
    select value into str_balance from get_structure_balance(st_version, blk_time); -- баланс структуры
	--

	select dev_address, profit_address, fee_address into dev_adr, profit_adr, fee_adr
    from structure_settings where version = st_version limit 1;
    
	--
    if profit_adr = fee_adr -- первое обновление fee, в остальных случаях эти адреса не могут совпадать
    then
		-- меняем тип старого адреса, т.к. изначально он был profit
		update structure_address
		set type = 'fee',
		    created_at = blk_time,
		    created_tx_height = confirm_tx__upd_profit_address.tx_height
		where version = st_version
		  and type = 'profit'
		  and deleted_at is null;
		--
		select confirmed_value into old_balance from get_address_balance(profit_adr, blk_time);
		-- больше не учитывается структура, т.к. это теперь fee-адресс
		perform upd_address_balance(profit_adr, -str_balance, blk_time, tx_height, 'update profit [1]', 'fee'::address_type);
		-- активный баланс старого profit становтися fee и не считается в dev
		perform upd_address_balance(dev_adr, -old_balance, blk_time, tx_height, 'update profit [1]', 'dev'::address_type);

        select confirmed_value into new_balance from get_address_balance(tx_recipient, blk_time);
		-- новый адрес не учитывается в балансе структуры
		perform upd_structure_balance(st_version, -new_balance, blk_time, tx_height, 'update profit [1]');
		-- включаем в баланс нового profit баланс структуры
		str_balance := str_balance - new_balance; 
		perform upd_address_balance(tx_recipient, str_balance, blk_time, tx_height, 'update profit [1]', 'profit'::address_type);
	else
		-- деактивируем старый адрес
		update structure_address
		set deleted_at = blk_time,
		    deleted_tx_height = confirm_tx__upd_profit_address.tx_height
		where version = st_version
		  and type = 'profit'
		  and deleted_at is null;
        --

		select confirmed_value into old_balance from get_address_balance(profit_adr, blk_time);
		select confirmed_value into new_balance from get_address_balance(tx_recipient, blk_time);

		-- выпиливаем баланс структуры из старого профита - остается только активный баланс
		perform upd_address_balance(profit_adr, str_balance, blk_time, tx_height, 'update profit', 'deposit'::address_type);

		-- выпиливаем баланс нового адреса из баланса структуры
		perform upd_structure_balance(st_version, -new_balance, blk_time, tx_height, 'update profit');
		perform upd_address_balance(dev_adr, -new_balance, blk_time, tx_height, 'update profit', 'dev'::address_type);

		-- добавляем активный баланс старого профит-адреса в структуру
		perform upd_structure_balance(st_version, old_balance, blk_time, tx_height, 'update profit*');

		-- увеличиваем баланс нового профит-адреса на размер структуры
		str_balance := str_balance + old_balance;
		perform upd_address_balance(tx_recipient, str_balance, blk_time, tx_height, 'update profit', 'profit'::address_type);
	end if;

	-- активируем новый адрес
	insert into structure_address (version, prefix, address, type, created_at, created_tx_height)
	values (st_version, st_prefix, tx_recipient, 'profit', blk_time, tx_height);

	-- обновляем настройки
	update structure_settings
       set profit_address = tx_recipient,
           updated_at = blk_time,
           tx_height = confirm_tx__upd_profit_address.tx_height
    where version = st_version
    returning dev_address, master_address, fee_address, name, profit_percent, fee_percent
         into dev_adr, mst_adr, fee_adr, st_name, st_profit_prc, st_fee_prc;
    --
    insert into structure_settings_log (version, prefix, name, profit_percent, fee_percent, dev_address, master_address, profit_address, fee_address, created_at, tx_height, comment)
    values (st_version, st_prefix, st_name, st_profit_prc, st_fee_prc, dev_adr, mst_adr, tx_recipient, fee_adr, blk_time, confirm_tx__upd_profit_address.tx_height, 'update profit');

    --
	insert into transaction (hash, height, confirmed_at, block_height, block_tx_idx, version, sender, recipient, struct)
    values (tx_hash, tx_height, blk_time, blk_height, blk_tx_idx, tx_version, tx_sender, tx_recipient, tx_struct);
end
$$;
`
