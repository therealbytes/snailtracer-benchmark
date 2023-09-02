package snailtracer

import (
	"math/big"
	"testing"

	"github.com/holiman/uint256"
)

var (
	bigintNums = []*big.Int{
		big.NewInt(5),
		big.NewInt(10),
		big.NewInt(-5),
		big.NewInt(-10),
	}
	uint256Nums = []*uint256.Int{
		uint256.NewInt(5),
		uint256.NewInt(10),
		new(uint256.Int).Neg(uint256.NewInt(5)),
		new(uint256.Int).Neg(uint256.NewInt(10)),
	}
)

func toInt64(u *uint256.Int) int64 {
	if u.Sign() < 0 {
		return -int64(new(uint256.Int).Neg(u).Uint64())
	}
	return int64(u.Uint64())
}

func TestOps(t *testing.T) {
	for i := 0; i < len(bigintNums); i++ {
		for j := 0; j < len(bigintNums); j++ {
			// Add
			{
				b := new(big.Int).Add(bigintNums[i], bigintNums[j])
				u := new(uint256.Int).Add(uint256Nums[i], uint256Nums[j])
				if b.Int64() != toInt64(u) {
					t.Errorf("%v != %v", b, toInt64(u))
				}
			}
			// Sub
			{
				b := new(big.Int).Sub(bigintNums[i], bigintNums[j])
				u := new(uint256.Int).Sub(uint256Nums[i], uint256Nums[j])
				if b.Int64() != toInt64(u) {
					t.Errorf("%v != %v", b, toInt64(u))
				}
			}
			// Mul
			{
				b := new(big.Int).Mul(bigintNums[i], bigintNums[j])
				u := new(uint256.Int).Mul(uint256Nums[i], uint256Nums[j])
				if b.Int64() != toInt64(u) {
					t.Errorf("%v != %v", b, toInt64(u))
				}
			}
			// Div
			{
				b := new(big.Int).Quo(bigintNums[i], bigintNums[j])
				u := new(uint256.Int).SDiv(uint256Nums[i], uint256Nums[j])
				if b.Int64() != toInt64(u) {
					t.Errorf("%v != %v", b, toInt64(u))
				}
			}
			// Rem
			{
				b := new(big.Int).Rem(bigintNums[i], bigintNums[j])
				u := new(uint256.Int).SMod(uint256Nums[i], uint256Nums[j])
				if b.Int64() != toInt64(u) {
					t.Errorf("%v != %v", b, toInt64(u))
				}
			}
		}
	}
}
