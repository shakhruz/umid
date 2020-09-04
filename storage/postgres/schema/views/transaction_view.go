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

package views

// TransactionView ...
const TransactionView = `
create or replace view transaction_view as
select encode(hash, 'hex')::text                                      as hash,
       height,
       confirmed_at,
       block_height,
       block_tx_idx,
       version,
       encode(sender, 'hex')::text                                    as sender,
       encode(recipient, 'hex')::text                                 as recipient,
       (value::numeric / 100)::money                                  as value,
       encode(fee_address, 'hex')::text                               as fee_address,
       (fee_value::numeric / 100)::money                              as fee_value,
       struct ->> 'prefix'                                            as struct_prefix,
       struct ->> 'name'                                              as struct_name,
       ((struct ->> 'profit_percent')::numeric / 100)::numeric(10, 2) as struct_profit_percent,
       ((struct ->> 'fee_percent')::numeric / 100) ::numeric(10, 2)   as struct_fee_percent
from transaction;
`
