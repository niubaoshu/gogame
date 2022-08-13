package goGame

import (
	"math/rand"
	"time"
)

type zobrist struct {
	boardHash [3][]uint64
	hash      uint64
	histHash  map[uint64]bool //出现过的局面hash

}

func newZobrist(long int) *zobrist {
	rd := rand.New(rand.NewSource(time.Now().UnixNano()))
	z := new(zobrist)
	rd.Int()
	z.hash = rd.Uint64()
	for i := range z.boardHash {
		bh := make([]uint64, long)
		for j := 0; j < long; j++ {
			bh[j] = rd.Uint64()
		}
		z.boardHash[i] = bh
	}
	z.histHash = make(map[uint64]bool, playerNum*long)
	return z
}

func (z *zobrist) calcBoardHash(perHash uint64, pos int, oldColor, newColor byte) uint64 {
	return perHash ^ z.boardHash[oldColor][pos] ^ z.boardHash[newColor][pos]
}
func (z *zobrist) reset() {
	for k, _ := range z.histHash {
		delete(z.histHash, k)
	}
}
