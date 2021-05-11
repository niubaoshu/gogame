package goGame

import (
	"math/rand"
	"time"
)

type zobrist struct {
	boardHash [3][]int
	player    [3]int
	hash      int
}

func newZobrist(long int) *zobrist {
	rd := rand.New(rand.NewSource(time.Now().UnixNano()))
	z := new(zobrist)
	z.hash = rd.Int()
	for i, bh := range z.boardHash {
		bh = make([]int, long)
		for j := 0; j < long; j++ {
			bh[j] = rd.Int()
		}
		z.boardHash[i] = bh
	}
	return z
}

func (z *zobrist) calcBoardHash(perHash, pos int, oc, nc Color) int {
	return perHash ^ z.boardHash[oc][pos] ^ z.boardHash[nc][pos]
}
