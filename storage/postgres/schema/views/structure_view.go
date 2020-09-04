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

// StructureView ...
const StructureView = `
create or replace view structure_view as
select prefix,
       name,
       dev_address,
       master_address,
       profit_address,
       fee_address,
       (d).confirmed_value as dev_confirmed,
       (d).composite_value as dev_composite,
       (p).confirmed_value as profit_confirmed,
       (p).composite_value as profit_composite,
       (f).confirmed_value as fee_confirmed,
       (b).value           as balance
from (
         select prefix,
                name,
                encode(dev_address, 'hex')::text    as dev_address,
                encode(master_address, 'hex')::text as master_address,
                encode(profit_address, 'hex')::text as profit_address,
                encode(fee_address, 'hex')::text    as fee_address,
                get_address_balance(dev_address)    as d,
                get_address_balance(profit_address) as p,
                get_address_balance(fee_address)    as f,
                get_structure_balance(version)      as b
         from structure_settings
     ) as s;
`
