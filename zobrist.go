package goGame

import (
	"math/rand"
	"time"
)

type zobrist struct {
	boardHash [3][]int
	hash      int
	histHash  map[int]bool //出现过的局面hash

}

func newZobrist(long int) *zobrist {
	rd := rand.New(rand.NewSource(time.Now().UnixNano()))
	z := new(zobrist)
	z.hash = rd.Int()
	for i := range z.boardHash {
		bh := make([]int, long)
		for j := 0; j < long; j++ {
			bh[j] = rd.Int()
		}
		z.boardHash[i] = bh
	}
	z.histHash = make(map[int]bool, playerNum*long)
	return z
}

func (z *zobrist) calcBoardHash(perHash, pos int, oldColor, newColor Color) int {
	return perHash ^ z.boardHash[oldColor][pos] ^ z.boardHash[newColor][pos]
}
