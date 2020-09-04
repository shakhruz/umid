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

// GetDevAddressMainnet ...
const GetDevAddressMainnet string = `
create or replace function get_dev_address()
    returns bytea
    language sql
    immutable
as
$$
	-- umi16uz7khspwq0patw777wgn7hgk6pvds2sxqgwt546z5n489mwmj2szdn2h5
    select decode('55a9d705eb5e01701e1eaddef79c89fae8b682c6c1503010e5d2ba152753976edc95', 'hex');
$$;
`

// GetDevAddressTestnet ...
const GetDevAddressTestnet string = `
create or replace function get_dev_address()
    returns bytea
    language sql
    immutable
as
$$
	-- umi16dhtrj348vaa63lp46u24hs5mjjjxzwqn75qwvnzke6uyr5txukqgckvra
    select decode('55a9d36eb1ca353b3bdd47e1aeb8aade14dca52309c09fa8073262b675c20e8b372c', 'hex');
$$;
`
