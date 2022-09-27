package common

import (
	"math/big"
)

func DeepEqualMapBigInt1D[K1 string | uint64](a, b map[K1]*big.Int) bool {
	aCopy := make(map[K1]*big.Int)
	for k, v := range a {
		aCopy[k] = v
	}

	bCopy := make(map[K1]*big.Int)
	for k, v := range b {
		bCopy[k] = v
	}

	for k, aV := range aCopy {
		bV, ok := b[k]
		if !ok || aV.Cmp(bV) != 0 {
			return false
		}
		delete(aCopy, k)
		delete(bCopy, k)
	}

	return len(aCopy) == 0 && len(bCopy) == 0
}

func DeepEqualMapBigInt2D[K1 string | uint64, K2 string | uint64](a, b map[K1]map[K2]*big.Int) bool {
	aCopy := make(map[K1]map[K2]*big.Int)
	for k, v := range a {
		aCopy[k] = v
	}

	bCopy := make(map[K1]map[K2]*big.Int)
	for k, v := range b {
		bCopy[k] = v
	}

	for k, aV := range aCopy {
		bV, ok := b[k]
		if !ok || !DeepEqualMapBigInt1D(aV, bV) {
			return false
		}
		delete(aCopy, k)
		delete(bCopy, k)
	}

	return len(aCopy) == 0 && len(bCopy) == 0
}

func DeepEqualMapBigInt3D[K1 string | uint64, K2 string | uint64, K3 string | uint64](a, b map[K1]map[K2]map[K3]*big.Int) bool {
	aCopy := make(map[K1]map[K2]map[K3]*big.Int)
	for k, v := range a {
		aCopy[k] = v
	}

	bCopy := make(map[K1]map[K2]map[K3]*big.Int)
	for k, v := range b {
		bCopy[k] = v
	}

	for k, aV := range aCopy {
		bV, ok := b[k]
		if !ok || !DeepEqualMapBigInt2D(aV, bV) {
			return false
		}
		delete(aCopy, k)
		delete(bCopy, k)
	}

	return len(aCopy) == 0 && len(bCopy) == 0
}
