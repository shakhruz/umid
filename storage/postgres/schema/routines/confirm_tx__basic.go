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

// ConfirmTxBasic ...
const ConfirmTxBasic = `
create or replace function confirm_tx__basic(bytes bytea,
                                             tx_height integer,
                                             blk_height integer,
                                             blk_tx_idx integer,
                                             blk_timestamp timestamptz)
    returns void
    language plpgsql
as
$$
declare
    umi_basic constant integer := x'55A9'::integer;
    --
    tx_hash            bytea;
    tx_ver             smallint;
    tx_sender          bytea;
    tx_recipient       bytea;
    tx_value           bigint;
    tx_fee_val         bigint;
    tx_fee_adr         bytea;
    --
    sender_ver         integer;
    recipient_ver      integer;
    --
    dev_address_       bytea;
    profit_address_    bytea;
    fee_address_       bytea;
    fee_percent_       smallint;
    --
begin
    select hash, version, sender, recipient, value
    into tx_hash, tx_ver, tx_sender, tx_recipient, tx_value
    from parse_transaction(bytes);

    --
    -- ОТПРАВИТЕЛЬ
    -- 

    -- списываем баланс отправителя
    perform upd_address_balance(tx_sender, -tx_value, blk_timestamp, tx_height, 'списание [sender]');

    sender_ver := (get_byte(tx_sender, 0) << 8) + get_byte(tx_sender, 1);
    if sender_ver <> umi_basic
    then
        select dev_address, profit_address, fee_address, fee_percent
        into dev_address_, profit_address_, fee_address_, fee_percent_
        from structure_settings
        where version = sender_ver
        limit 1;

        if tx_sender not in (dev_address_, profit_address_, fee_address_)
        then
            perform upd_structure_balance(sender_ver, -tx_value, blk_timestamp, tx_height, 'списание [структура]');
            perform upd_address_balance(dev_address_, -tx_value, blk_timestamp, tx_height, 'списание [dev]');
            perform upd_address_balance(profit_address_, -tx_value, blk_timestamp, tx_height, 'списание [profit]');
		elseif tx_sender = profit_address_
		then
            perform upd_address_balance(dev_address_, -tx_value, blk_timestamp, tx_height, 'списание [dev из-за profit]');
        end if;
    end if;

    --
    -- ПОЛУЧАТЕЛЬ
    --

    recipient_ver := (get_byte(tx_recipient, 0) << 8) + get_byte(tx_recipient, 1);
    if recipient_ver <> umi_basic then
        select dev_address, profit_address, fee_address, fee_percent
        into dev_address_, profit_address_, fee_address_, fee_percent_
        from structure_settings
        where version = recipient_ver
        limit 1;

        -- проверяем, нужно ли списывать комиссию
        if not exists(
                select 1
                from structure_address
                where (address = tx_recipient or address = tx_sender) 
                  and created_tx_height < tx_height
                  and (deleted_tx_height is null or deleted_tx_height > tx_height)
            )
        then
            if fee_percent_ > 0
            then
                tx_fee_val := ceil((tx_value::bigint * fee_percent_::bigint)::double precision / 10000)::bigint;
                tx_value := tx_value - tx_fee_val;
                tx_fee_adr := fee_address_;

                perform upd_address_balance(fee_address_, tx_fee_val, blk_timestamp, tx_height, 'перевод комиссии [fee]');
            end if;
        end if;

        -- баланс структуры
        if tx_recipient not in (dev_address_, profit_address_, fee_address_)
        then
            perform upd_structure_balance(recipient_ver, tx_value, blk_timestamp, tx_height, 'пополнение [структура]');
            perform upd_address_balance(dev_address_, tx_value, blk_timestamp, tx_height, 'пополнение [dev]');
            perform upd_address_balance(profit_address_, tx_value, blk_timestamp, tx_height, 'пополнение [profit]');
        elseif tx_recipient = profit_address_
		then -- дев включает в себя баланс профита
            perform upd_address_balance(dev_address_, tx_value, blk_timestamp, tx_height, 'пополнение [dev из-за profit]');
        end if;
    end if;

    -- обновляем баланс получателя
    perform upd_address_balance(tx_recipient, tx_value, blk_timestamp, tx_height, 'пополнение [recipient]');

    -- добавляем транзакцию
    insert into transaction (hash, height, confirmed_at, block_height, block_tx_idx, version, sender, recipient, value,
                             fee_address, fee_value)
    values (tx_hash, tx_height, blk_timestamp, blk_height, blk_tx_idx, tx_ver, tx_sender, tx_recipient, tx_value,
            tx_fee_adr, tx_fee_val);
end
$$;
`
