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

// AddGenesis ...
const AddGenesis = `
create or replace function add_genesis(testnet boolean default false)
    returns void
    language plpgsql
as
$$
begin
    if testnet 
    then
		-- umi1tjsevd884gzt78cn26yjmke3tyer4e0tfxl74yq8unrd9s0wts4q5w7xrc
		perform add_block(
			decode('AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAbqGM9xN+MMHByv1kFt69emP4R6CqvwQg0tKe82wMvaleq2aAAAFGUtYJfyRDT6y8GCiU+vpPxa3njbPgFwWmMk012NNKaQLfhLb1H+/6Yga/aDh/NND2D2FDiFV25kKoU22jx1PGhBtDIyEDTOkC9MQWyDJoniSNd3BoNMT3SC0wg0LKMggAAABGUtYJfyRDT6y8GCiU+vpPxa3njbPgFwWmMk012NNKaVWpXKGWNOeqBL8fE1aJLdsxWTI65etJv+qQB+TG0sHuXCoAAAAAa0nSAAAAAAAAAAAASlANBycRcGMYcezGYPGz9V0uNncrUyyyZA4QWaOZb4dxi5dWOHIM4nU4qBuwT38ckuQ9fyVHSjmyThmcyClTDwA=', 'base64')
		);
	else
		-- umi1funq869de06x8tdfth0hs5z8v37ral3yyalmuvj7j0z94rtxcaxq289cf9
		perform add_block(
			decode('AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA8SFc9pyaSvD83oKoXLIKMiPE4Zj+3yVYLER3MDSFNRde1EUAAAFFiFyWh9eZpNH014bYY5J00pPtAkrX96Q2cV1iF8n3Kxzo1AHGwhBhyRRflMUsmqkaZF2nLVHAKxHxLlyNpv3mdispBu93l6lFWjuqoP+GmcZkP5j708c4HCxe+mQYlAMAAABFiFyWh9eZpNH014bYY5J00pPtAkrX96Q2cV1iF8n3K1WpTyYD6K3L9GOtqV3feFBHZHw+/iQnf74yXpPEWo1mx0wAAAAAa0nSAAAAAAAAAAAA8kVS09bWGbNyDAlWMafmxeiv9I/V3aAjeeKgL1x7AiNcfAnxd9KHCZfhvUsOltaIFnjqGFjDb2xtNWv2e6ldAgA=', 'base64')
		);
	end if;
    perform confirm_next_block();
end
$$;
`
