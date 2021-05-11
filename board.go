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
	size     int
	long     int
	board    []Color
	stack    *stack
	zh       *zobrist
	histHash map[int]bool
	posNum   int // pos num 黑白子数的和
	histPos  []int
	idxPos   []int
}

func NewBoard(size int) *Board {
	long := size * size
	b := &Board{
		long:     long,
		size:     size,
		board:    make([]Color, long),
		stack:    &stack{s: make([]int, 0, long/2)},
		zh:       newZobrist(long),
		histHash: make(map[int]bool, MAXPLAYSISE*long),
		histPos:  make([]int, long+1),
		idxPos:   make([]int, long+1),
	}
	b.histHash[b.zh.hash] = true
	for i := 0; i < long; i++ {
		b.histPos[i] = i
		b.idxPos[i] = i
	}
	b.histPos[b.long] = PASS
	return b
}

// pos 能否到达c,如果不能把所有pos返回
func (b *Board) isReachColorAndTint(c Color, pos int, tint Color) (bool, []int) {
	b.stack.reset()
	pc := b.board[pos]
	result := make([]int, 0, 2)
	b.stack.push(pos)
	b.board[pos] = tint
	result = append(result, pos)
	for !b.stack.isEmpty() {
		p := b.stack.pop()
		for _, p := range b.getNeighbours(p) {
			if b.board[p] == c {
				return true, result
			} else if b.board[p] == pc {
				b.board[p] = tint
				result = append(result, p)
				b.stack.push(p)
			}
		}
	}
	return false, result
}

func (b *Board) IsReachColor(c Color, pos int) (bool, []int) {
	pc := b.board[pos]
	ok, reachPos := b.isReachColorAndTint(c, pos, REACH)
	l := len(reachPos)
	for i := 0; i < l; i++ {
		p := reachPos[i]
		if b.board[p] == REACH {
			b.board[p] = pc
		} else {
			reachPos[i], reachPos[l-1] = reachPos[l-1], reachPos[i]
			i--
			l--
		}
	}
	return ok, reachPos[:l]
}

func (b *Board) randRun() (result int) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	player := BLACK
	p := 0
	prePass := false
	loopNum := 0
	for {
		//if b.posNum > 300 {
		//	fmt.Println(loopNum, b.posNum)
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
	return b.calcResult()
}

func (b *Board) isOwer(c Color, pos int) bool {
	ok, _ := b.IsReachColor(c, pos)
	ok2, _ := b.IsReachColor(reverseColor(c), pos)
	return ok && !ok2
}

func (b *Board) isOtherBig(c Color, pos int) bool {
	rc := reverseColor(c)
	ok, s := b.IsReachColor(rc, pos)
	ok2, _ := b.IsReachColor(c, pos)
	if ok && !ok2 && len(s) > 6 && b.posNum > 100 {
		fmt.Println(len(s), ok, ok2, c, pos)
		b.print()
	}
	return ok && !ok2 && len(s) > 6 && b.posNum > 100
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
	}
	b.histHash[b.zh.hash] = true
}

func (b *Board) changeBoard(c Color, pos int) {
	b.setColorAndHash(c, pos)
	//b.recordPos(pos, b.posNum)
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
			//if start == b.long {
			//return PASS
			//}
			continue
		}
		pos := b.histPos[idx]
		if b.isForbidden(pos, c) || b.getTakeNum(c, pos) < 0 || b.isOwer(c, pos) ||
			b.isOtherBig(c, pos) {
			b.switchIdx(idx, start)
			start++
			end--
		} else {
			return b.histPos[idx]
		}
	}

	return PASS
}

func (b *Board) isForbidden(pos int, c Color) bool {
	if b.board[pos] != EMPTY {
		fmt.Println("错误", pos, b.board[pos], b.posNum)
		//time.Sleep(time.Second * 10)
		return true
	}
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
	pc := reverseColor(c)
	hb := b.zh.calcBoardHash(b.zh.hash, pos, EMPTY, c) ^ b.zh.player[pc]
	b.board[pos] = c
	defer func() { b.board[pos] = EMPTY }()
	d1 := b.getDeadPos(reverseColor(c), pos)
	if len(d1) == 0 {
		pc = c
		d1 = b.getDeadPos(c, pos)
	}
	for _, p := range d1 {
		//if b.board[p] == EMPTY {
		//	continue
		//}
		hb = b.zh.calcBoardHash(hb, p, pc, EMPTY)
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
	pc := b.board[pos]
	if len(b.getNeighbours(pos)) == len(b.getNeighboursByColor(rc, pos)) {
		//if pos%b.size != 0 && b.board[pos-1] == rc && // have left Neighbours
		//	pos%b.size != b.size-1 && b.board[pos+1] == rc && // have right Neighbours
		//	pos >= b.size && b.board[pos-b.size] == rc && // have up Neighbours
		//	pos <= b.long-b.size && b.board[pos+b.size] == rc { //have down Neighbours
		b.board[pos] = c
		defer func() { b.board[pos] = pc }()
		n := len(b.getDeadPos(rc, pos))
		if n <= 0 {
			return true
		} else {
			//fmt.Println("无气落子，提n颗", n, c, pos)
		}
	}
	return false
}

func (b *Board) getTakeNum(c Color, pos int) int {
	b.board[pos] = c
	defer func() { b.board[pos] = EMPTY }()
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

func (b *Board) getDeadPos(c Color, pos int) []int {
	deadPos := make([]int, 0, 8)
	nbs := b.getNeighboursByColor(c, pos)
	tm := make(map[int]bool, 10)
	for _, p := range nbs {
		if ok, r := b.IsReachColor(EMPTY, p); !ok {
			for _, pp := range r {
				if tm[pp] {
					break
				} else {
					tm[pp] = true
					deadPos = append(deadPos, pp)
				}
			}
		}
	}
	return deadPos
}

func (b *Board) getSameColorNeighbours(pos int) []int {
	return b.getNeighboursByColor(b.board[pos], pos)
}
func (b *Board) getNeighbours(pos int) []int {
	var nbs = make([]int, 0, 4)
	if pos >= b.long && pos < 0 {
		return nbs
	}
	// have left
	if pos%b.size != 0 {
		nbs = append(nbs, pos-1)
	}
	// have right
	if pos%b.size != b.size-1 {
		nbs = append(nbs, pos+1)
	}
	// have up
	if pos >= b.size {
		nbs = append(nbs, pos-b.size)
	}
	// have down
	if pos < b.long-b.size {
		nbs = append(nbs, pos+b.size)
	}
	return nbs
}

func (b *Board) getNeighboursByColor(c Color, pos int) []int {
	var nbs = make([]int, 0, 4)
	if pos >= b.long && pos < 0 {
		return nbs
	}
	// have left
	if pos%b.size != 0 && b.board[pos-1] == c {
		nbs = append(nbs, pos-1)
	}
	// have right
	if pos%b.size != b.size-1 && b.board[pos+1] == c {
		nbs = append(nbs, pos+1)
	}
	// have up
	if pos >= b.size && b.board[pos-b.size] == c {
		nbs = append(nbs, pos-b.size)
	}
	// have down
	if pos < b.long-b.size && b.board[pos+b.size] == c {
		nbs = append(nbs, pos+b.size)
	}
	return nbs
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

type stack struct {
	s    []int
	long int
}

func (s *stack) isEmpty() bool {
	return s.long == 0
}

func (s *stack) reset() {
	s.long = 0
	s.s = s.s[0:0]
}

func (s *stack) pop() int {
	s.long--
	e := s.s[s.long]
	s.s = s.s[:s.long]
	return e
}
func (s *stack) push(e int) {
	s.long++
	s.s = append(s.s, e)
}

func (c Color) String() string {
	return colorString[c]
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
		} else {
			fmt.Print(b.board[i])
		}
	}
	fmt.Println()
}
