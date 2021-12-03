// Copyright (c) 2021 UMI
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

package ledger

import (
	"gitlab.com/umitop/umid/pkg/umi"
)

type level struct {
	balance      uint64
	interestRate uint16
}

func (confirmer *Confirmer) stopStaking() {
	for pfx, structure := range confirmer.ledger.structures {
		if pfx == umi.PfxVerUmi {
			continue
		}

		structure.Balance = structure.BalanceAt(confirmer.BlockTimestamp)
		structure.UpdatedAt = confirmer.BlockTimestamp
		structure.Level = 0
		structure.LevelInterestRate = 0

		confirmer.updateLevelAddresses(pfx)
	}
}

func (confirmer *Confirmer) checkGlize() {
	if _, ok := confirmer.ledger.structures[umi.PfxVerGls]; !ok {
		return
	}

	if _, ok := confirmer.ledger.structures[umi.PfxVerGlz]; !ok {
		return
	}

	totalGls := confirmer.ledger.structures[umi.PfxVerGls].BalanceAt(confirmer.BlockTimestamp)
	totalGlz := confirmer.ledger.structures[umi.PfxVerGlz].BalanceAt(confirmer.BlockTimestamp)
	totalSupply := totalGls + totalGlz
	levels := []level{{0, 0}}

	switch {
	case totalSupply < 250_000_000_00: // Агрессивный этап эмиссии
		levels = append(levels, level{5_000_000_00, 8_00},
			level{10_000_000_00, 10_00},
			level{20_000_000_00, 12_00},
			level{30_000_000_00, 15_00},
			level{50_000_000_00, 18_00},
			level{75_000_000_00, 21_00},
			level{100_000_000_00, 25_00},
			level{150_000_000_00, 30_00})
	case totalSupply < 500_000_000_00: // Нейтральный этап эмиссии
		levels = append(levels, level{50_000_000_00, 8_00},
			level{100_000_000_00, 10_00},
			level{150_000_000_00, 12_00},
			level{200_000_000_00, 14_00},
			level{300_000_000_00, 16_00},
			level{400_000_000_00, 20_00})
	case totalSupply < 1_000_000_000_00: // Пассивный этап эмиссии
		levels = append(levels, level{50_000_000_00, 6_00},
			level{100_000_000_00, 7_00},
			level{200_000_000_00, 8_00},
			level{300_000_000_00, 9_00},
			level{450_000_000_00, 10_00},
			level{600_000_000_00, 11_00},
			level{700_000_000_00, 13_00},
			level{800_000_000_00, 15_00})
	}

	structure := confirmer.ledger.structures[umi.PfxVerGls]

	for lvl := len(levels) - 1; lvl >= 0; lvl-- {
		newLevel := uint8(lvl)
		newInterestRate := levels[lvl].interestRate

		if totalGls >= levels[lvl].balance {
			if structure.LevelInterestRate != newInterestRate {
				structure.Balance = totalGls
				structure.UpdatedAt = confirmer.BlockTimestamp
				structure.Level = newLevel
				structure.LevelInterestRate = newInterestRate

				confirmer.updateLevelAddresses(umi.PfxVerGls)
			}

			break
		}
	}
}

func (confirmer *Confirmer) checkStructureLevel() {
	levels := [...]level{
		10: {1_000_000_000_00, 41_00},
		9:  {500_000_000_00, 39_00},
		8:  {100_000_000_00, 37_00},
		7:  {50_000_000_00, 36_00},
		6:  {10_000_000_00, 35_00},
		5:  {5_000_000_00, 30_00},
		4:  {1_000_000_00, 25_00},
		3:  {500_000_00, 20_00},
		2:  {100_000_00, 15_00},
		1:  {50_000_00, 10_00},
		0:  {0, 0},
	}

	for pfx, structure := range confirmer.ledger.structures {
		switch pfx {
		case umi.PfxVerUmi, umi.PfxVerGlz, umi.PfxVerGls:
			continue
		}

		timestamp := confirmer.BlockTimestamp
		balance := structure.BalanceAt(timestamp)

		for lvl := 10; lvl >= 0; lvl-- {
			newLevel := uint8(lvl)
			newInterestRate := levels[lvl].interestRate

			if balance >= levels[lvl].balance {
				if structure.Level != newLevel {
					structure.Balance = balance
					structure.UpdatedAt = timestamp
					structure.Level = newLevel
					structure.LevelInterestRate = newInterestRate

					confirmer.updateLevelAddresses(pfx)
				}

				break
			}
		}
	}
}

func (confirmer *Confirmer) updateLevelAddresses(pfx umi.Prefix) {
	timestamp := confirmer.BlockTimestamp
	structure := confirmer.ledger.structures[pfx]

	for _, acc := range confirmer.ledger.accounts[pfx] {
		acc.SetInterestRate(structure.InterestRate(acc.Type), timestamp)
	}
}
