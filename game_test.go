package goGame

import (
	"testing"
)

func TestGame(t *testing.T) {
	b := NewBoard(19)
	testBoard(t, b, true)
	b.Reset(b.bytes)
	testBoard(t, b, true)
	//testBoard(t, 9, false)
	//testBoard(t, 10, false)
}

func testBoard(t *testing.T, b *Board, display bool) {
	t.Log(b.RandRun(display))
	if (b.moveNum[WHITE] + b.moveNum[BLACK] - b.passNum[WHITE] - b.passNum[BLACK] - b.posNum - b.takeNum[BLACK] - b.takeNum[WHITE]) != 0 {
		t.Error("报错了")
	}

	if (b.score[BLACK] + b.score[WHITE]) != b.long*2 {
		t.Error("score 不对")
	}

	if b.posNum != b.colorNum[WHITE]+b.colorNum[BLACK] || b.posNum+b.colorNum[EMPTY] != b.long {
		t.Error("报错了")
	}
}
func TestGame2(t *testing.T) {
	//b := NewBoard(19)
	//for i := 0; i < b.long-1; i++ {
	//	b.changeBoard(WHITE, i)
	//}
	//b.randRun()
	//b.print()
}

func TestGameHash(t *testing.T) {
	b := NewBoard(19)
	h := b.zh.hash
	b.changeBoard(WHITE, 0)
	b.changeBoard(EMPTY, 0)
	if b.zh.hash != h {
		t.Error("hash 计算错误")
	}
}
func TestZobrist(t *testing.T) {
	z := newZobrist(19 * 19)
	h1 := z.hash
	z.hash = z.calcBoardHash(h1, 0, EMPTY, WHITE)
	h2 := z.calcBoardHash(z.hash, 0, WHITE, EMPTY)
	if h1 != h2 {
		t.Error("hash 计算错误")
	}
}
