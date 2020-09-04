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

package tables

// StructureSettingsLog ...
const StructureSettingsLog = `
create table if not exists structure_settings_log
(
    version        integer                  not null,
    prefix         char(3)                  not null,
    name           text                     not null,
    profit_percent smallint                 not null,
    fee_percent    smallint                 not null,
    dev_address    bytea                    not null,
    master_address bytea                    not null,
    profit_address bytea                    not null,
    fee_address    bytea                    not null,
    created_at     timestamp with time zone not null,
    tx_height      integer                  not null,
    comment        text,
    check (profit_percent between 0 and 500 and fee_percent between 0 and 2000)
);
`

// StructureSettingsLogIdx ...
const StructureSettingsLogIdx = `
create index if not exists structure_settings_log_idx
    on structure_settings_log (version);
`
