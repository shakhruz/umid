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

package schema

import (
	"os"
	"umid/storage/postgres/schema/objtypes"
	"umid/storage/postgres/schema/routines"
	"umid/storage/postgres/schema/sequences"
	"umid/storage/postgres/schema/tables"
	"umid/storage/postgres/schema/views"
)

// Migrations ...
func Migrations() [][]string {
	v := make([][]string, 2)

	v = append(v, v2())
	v = append(v, v3())
	v = append(v, v4())

	return v
}

func v2() []string {
	return []string{
		objtypes.AddressType,

		tables.AddressBalanceConfirmed, tables.AddressBalanceConfirmedIdx,
		tables.AddressBalanceConfirmedLog, tables.AddressBalanceConfirmedLogIdx,
		tables.Block, tables.BlockHeightUidx, tables.BlockPrevUidx, tables.BlockConfIdx,
		tables.Level, tables.LevelData,
		tables.Mempool, tables.MempoolIdx,
		tables.StructureAddress,
		tables.StructureBalance,
		tables.StructureBalanceLog,
		tables.StructurePercent,
		tables.StructurePercentLog, tables.StructurePercentLogIdx,
		tables.StructureSettings, tables.StructureSettingsPrefixUidx,
		tables.StructureSettingsLog, tables.StructureSettingsLogIdx,
		tables.StructureStats,
		tables.Transaction,

		sequences.TxHeight,
	}
}

func v3() []string {
	return []string{
		routines.AddBlock,
		routines.AddGenesis,
		routines.AddTransaction,
		routines.ConfirmNextBlock,
		routines.ConfirmTxAddStructure,
		routines.ConfirmTxAddTransitAddress,
		routines.ConfirmTxBasic,
		routines.ConfirmTxDelTransitAddress,
		routines.ConfirmTxGenesis,
		routines.ConfirmTxUpdFeeAddress,
		routines.ConfirmTxUpdProfitAddress,
		routines.ConfirmTxUpdStructure,
		routines.ConvertPrefixToVersion,
		routines.ConvertVersionToPrefix,
		routines.GetAddressBalance,
		routines.GetAddressTransactions,
		func() string {
			if val, ok := os.LookupEnv("NETWORK"); ok && val == "testnet" {
				return routines.GetDevAddressTestnet
			}

			return routines.GetDevAddressMainnet
		}(),
		routines.GetStructureBalance,
		routines.ParseAddress,
		routines.ParseBlockHeader,
		routines.ParseTransaction,
		routines.TruncateBlockchain,
		routines.UpdAddressBalance,
		routines.UpdStructureBalance,
		routines.UpdStructureLevel,

		views.AddressBalanceConfirmedView,
		views.BlockView,
		views.StructureAddressView,
		views.StructureSettingsView,
		views.StructureView,
		views.TransactionView,

		func() string {
			if val, ok := os.LookupEnv("NETWORK"); ok && val == "testnet" {
				return `select add_genesis(true)`
			}

			return `select add_genesis(false)`
		}(),
	}
}

func v4() []string {
	return []string{
		tables.TransactionIdxSender,
		tables.TransactionIdxRecipient,
		tables.TransactionIdxFeeAddress,

		routines.GetAddressTransactions,
		routines.GetStructures,
		routines.GetStructureByPrefix,
	}
}
