package goGame

import (
	"testing"
)

func TestGame(t *testing.T) {
	b := NewBoard(19)
	testBoard(t, b)
	b.Reset(b.bytes)
	testBoard(t, b)
	//testBoard(t, 9, false)
	//testBoard(t, 10, false)
}

func testBoard(t *testing.T, b *Board) {
	b.RandRun()
	b.CalcScore()
	if err := b.CheckError(); err != nil {
		t.Error("报错了", err)
	}
	b.Display()
	t.Log(b.Bytes())
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
