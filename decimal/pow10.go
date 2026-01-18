package decimal

import (
	"math/big"
	"sync"
)

var (
	pow10Lock  sync.Mutex
	pow10Cache = []big.Int{*big.NewInt(1)}
	bigTen     = big.NewInt(10)
)

func pow10(n int32) *big.Int {
	if n < 0 {
		return big.NewInt(1)
	}

	pow10Lock.Lock()
	defer pow10Lock.Unlock()

	if int(n) < len(pow10Cache) {
		return &pow10Cache[n]
	}

	for i := len(pow10Cache); i <= int(n); i++ {
		var next big.Int
		next.Mul(&pow10Cache[i-1], bigTen)
		pow10Cache = append(pow10Cache, next)
	}

	return &pow10Cache[n]
}
