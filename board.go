package goGame

import (
	"fmt"
	"math/rand"
	"time"
)

/*
The Logical Rules
Go is played on a 19x19 square grid of points, by two players called Black and White.
Each point on the grid may be colored black, white or empty.
A point P, not colored C, is said to reach C, if there is a path of (vertically or horizontally)
adjacent points of P's color from P to a point of color C.
Clearing a color is the process of emptying all points of that color that don't reach empty.
Starting with an empty grid, the players alternate turns, starting with Black.
A turn is either a pass; or a move that doesn't repeat an earlier grid coloring.
A move consists of coloring an empty point one's own color;
then clearing the opponent color, and then clearing one's own color.
The game ends after two consecutive passes.
A player's score is the number of points of her color, plus the number of empty points that reach only her color.
The player with the higher score at the end of the game is the winner. Equal scores result in a tie.
*/

type Color int

var (
	colorString = []string{"EMPTY", "WHITE", "BLACK", "WALL", "REACH", "RESULT", "RESULT2"}
)

const (
	EMPTY Color = iota
	WHITE
	BLACK
	WALL
	REACH
	RESULT
	RESULT2

	MAXPLAYSISE = 2
	PASS        = -1
)

type Board struct {
	size       int
	long       int
	board      []Color
	zh         *zobrist
	histHash   map[int]bool
	posNum     int // pos num 黑白子数的和
	histPos    []int
	idxPos     []int
	neighbours [][]int
	block      []int
	colorNum   []int
}

func NewBoard(size int) *Board {
	long := size * size
	b := &Board{
		long:       long,
		size:       size,
		board:      make([]Color, long),
		zh:         newZobrist(long),
		histHash:   make(map[int]bool, MAXPLAYSISE*long),
		histPos:    make([]int, long+1),
		idxPos:     make([]int, long+1),
		block:      make([]int, long),
		colorNum:   make([]int, 3),
		neighbours: make([][]int, long),
	}

	b.colorNum[EMPTY] = long
	b.histHash[b.zh.hash] = true
	for i := 0; i < long; i++ {
		b.neighbours[i] = getNeighbours(i, size)
		b.histPos[i] = i
		b.idxPos[i] = i
	}
	b.histPos[b.long] = PASS
	return b
}

// 获得从pos所在的块（颜色相同），切测试最终能否到达c色
func (b *Board) GetBlockAndIsReachColorByPos(c Color, pos int, findReturn bool) (bool, []int) {
	reach := false
	pc := b.board[pos]
	block := b.block
	l := 0
	if pc == c {
		reach = true
		if findReturn {
			return reach, block[:l]
		}
	}
	block[l] = pos
	b.board[pos] = REACH
	l++
loop:
	for i := 0; i < l; i++ {
		for _, p := range b.neighbours[block[i]] {
			//fmt.Println(b.toXY(p), b.board[p], c)
			if b.board[p] == c {
				reach = true
				if findReturn {
					break loop
				}
			} else if b.board[p] == pc {
				b.board[p] = REACH
				block[l] = p
				l++
			}
		}
	}
	for _, p := range block[:l] {
		b.board[p] = pc
	}
	return reach, block[:l]
}

func (b *Board) randRun() (result int) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	player := BLACK
	p := 0
	prePass := false
	loopNum := 0
	for {
		//if b.posNum > 100 {
		//	time.Sleep(time.Second / 20)
		//	b.print()
		//}
		loopNum++
		p = b.randPos(player, r)
		if p == PASS {
			if prePass {
				break
			}
			prePass = true
		} else {
			prePass = false
			b.move(player, p)
		}
		player = reverseColor(player)
	}
	fmt.Println(loopNum, b.posNum, b.colorNum)
	return b.calcResult()
}

func (b *Board) isMineOrOtherBig(c Color, pos int) bool {
	if b.colorNum[BLACK] == 0 || b.colorNum[WHITE] == 0 {
		return false
	}
	rc := reverseColor(c)
	ok, _ := b.GetBlockAndIsReachColorByPos(c, pos, true)
	ok2, s := b.GetBlockAndIsReachColorByPos(rc, pos, false)
	if ok && !ok2 {
		return true // isMine
	}
	return !ok && ok2 && len(s) > 6 //isOtherBig
}

func (b *Board) calcResult() int {
	result := 0
	perColor := REACH
	for i := 0; i < b.long; i++ {
		if b.board[i] == BLACK {
			perColor = BLACK
			result++
			continue
		}
		if b.board[i] == EMPTY {
			if perColor == BLACK {
				result++
				continue
			}
		}
		if b.board[i] == WHITE {
			perColor = WHITE
		}
	}
	return result
}

func (b *Board) move(c Color, pos int) {
	// state
	b.changeBoard(c, pos)

	rcN := b.TakeColor(reverseColor(c), pos)
	//cN := 0
	if rcN == 0 {
		_ = b.TakeColor(c, pos)
	}
	if rcN >= 360 || b.posNum == 360 {
		//fmt.Println(rcN, cN, c, pos)
		//b.print()
	}
	// history
	b.zh.hash = b.zh.hash ^ b.zh.player[reverseColor(c)]
	if _, has := b.histHash[b.zh.hash]; has {
		fmt.Println("真是hash重复", b.zh.hash)
		panic("局面重复")
	}
	b.histHash[b.zh.hash] = true
}

func (b *Board) changeBoard(c Color, pos int) {
	b.colorNum[b.board[pos]]--
	b.colorNum[c]++
	b.setColorAndHash(c, pos)
	if c == EMPTY {
		b.posNum--
		b.switchIdx(b.idxPos[pos], b.posNum)
	} else if c == BLACK || c == WHITE {
		b.switchIdx(b.idxPos[pos], b.posNum)
		b.posNum++
	} else {
		panic("不可能")
	}
}

func (b *Board) setColorAndHash(c Color, pos int) {
	b.zh.hash = b.zh.calcBoardHash(b.zh.hash, pos, b.board[pos], c)
	b.board[pos] = c
}

func (b *Board) randPos(c Color, r *rand.Rand) int {
	start := b.posNum
	end := len(b.histPos) - b.posNum
	for start < b.long {
		idx := start + r.Intn(end)
		if b.histPos[idx] == PASS {
			continue
		}
		pos := b.histPos[idx]

		if b.board[pos] != EMPTY {
			fmt.Println("错误", pos, b.board[pos], b.posNum)
		}

		isMineOrOtherBig := b.isMineOrOtherBig(c, pos)
		pc := b.board[pos]
		b.board[pos] = c
		if b.isForbidden(pos, c) || b.getTakeNum(c, pos) < 0 || isMineOrOtherBig {
			b.board[pos] = pc
			b.switchIdx(idx, start)
			start++
			end--
		} else {
			b.board[pos] = pc
			return b.histPos[idx]
		}
	}

	return PASS
}

func (b *Board) isForbidden(pos int, c Color) bool {
	if b.isOneSuicide(pos, c) {
		//fmt.Println(b.posNum, c, pos, "自杀")
		return true
	}
	if b.isSuperKO(c, pos) {
		//fmt.Println(b.posNum, c, pos, "全同,len", len(b.histHash))
		//b.print()
		return true
	}
	return false
}
func (b *Board) isSuperKO(c Color, pos int) bool {
	rc := reverseColor(c)
	hb := b.zh.calcBoardHash(b.zh.hash, pos, EMPTY, c) ^ b.zh.player[rc]
	deadPos := b.getDeadPos(rc, pos)
	deadColor := rc
	if len(deadPos) == 0 {
		deadColor = c
		deadPos = b.getDeadPos(c, pos)
	}
	for _, p := range deadPos {
		//if b.board[p] == EMPTY {
		//	continue
		//}
		hb = b.zh.calcBoardHash(hb, p, deadColor, EMPTY)
	}

	if _, has := b.histHash[hb]; has {
		//fmt.Println("重复", hb)
		return true
	} else {
		//fmt.Println("预测hash未重复", hb)
		return false
	}
}

func (b *Board) isOneSuicide(pos int, c Color) bool {
	rc := reverseColor(c)
	for _, p := range b.neighbours[pos] {
		if b.board[p] == EMPTY || b.board[p] == c {
			return false
		}
	}
	n := len(b.getDeadPos(rc, pos))
	if n <= 0 {
		return true
	} else {
		//fmt.Println("无气落子，提n颗", n, c, pos)
	}
	return false
}

func (b *Board) getTakeNum(c Color, pos int) int {
	n := len(b.getDeadPos(reverseColor(c), pos))
	if n > 0 {
		return n
	} else {
		return -len(b.getDeadPos(c, pos))
	}
}

func (b *Board) TakeColor(c Color, pos int) int {
	deadPos := b.getDeadPos(c, pos)
	ret := 0
	for _, p := range deadPos {
		if b.board[p] == EMPTY {
			continue
		}
		ret++
		b.changeBoard(EMPTY, p)
	}
	return ret
}

//获得pos相邻的c色死子
func (b *Board) getDeadPos(c Color, pos int) []int {
	deadPos := make([]int, 0, 8)
	nbs := b.neighbours[pos]
	tm := make(map[int]bool, 10)
	for _, p := range nbs {
		if b.board[p] == c {
			if ok, block := b.GetBlockAndIsReachColorByPos(EMPTY, p, true); !ok {
				for _, pp := range block {
					if tm[pp] {
						break
					} else {
						tm[pp] = true
						deadPos = append(deadPos, pp)
					}
				}
			}
		}
	}
	return deadPos
}

func getNeighbours(pos, size int) []int {
	nb := make([]int, 4)
	i := 0
	//have left
	if pos%size != 0 {
		nb[i] = pos - 1
		i++
	}
	//have right
	if pos%size != size-1 {
		nb[i] = pos + 1
		i++
	}
	//have up
	if pos >= size {
		nb[i] = pos - size
		i++
	}
	//have down
	if pos < size*(size-1) {
		nb[i] = pos + size
		i++
	}
	return nb[:i]
}
func (b *Board) switchIdx(i1, i2 int) {
	v1 := b.histPos[i1]
	v2 := b.histPos[i2]
	b.histPos[i1], b.histPos[i2] = v2, v1
	b.idxPos[v1] = i2
	b.idxPos[v2] = i1

	for i := b.posNum + 1; i < b.long; i++ {
		if b.board[b.histPos[i]] != EMPTY {
			fmt.Println("报错。。。了")
		}
	}
}

func reverseColor(c Color) Color {
	if c == BLACK {
		return WHITE
	} else {
		return BLACK
	}
}

func (b *Board) toXY(pos int) string {
	return fmt.Sprintf("%d,%d", pos/b.size+1, pos%b.size+1)
}

func (b *Board) print() {
	fmt.Print("   1 2 3 4 5 6 7 8 910111213141516171819")
	for i := 0; i < b.long; i++ {
		if i%b.size == 0 {
			fmt.Printf("\n%2d", i/b.size+1)
		}
		if b.board[i] == BLACK {
			fmt.Printf(" B")
		} else if b.board[i] == WHITE {
			fmt.Print(" W")
		} else if b.board[i] == EMPTY {
			fmt.Print(" _")
		} else if b.board[i] == REACH {
			fmt.Print(" R")
		} else {
			fmt.Print(b.board[i])
		}
	}
	fmt.Println()
}
func (c Color) String() string {
	return colorString[c]
}
