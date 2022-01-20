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
	WHITE Color = iota
	EMPTY
	BLACK

	WALL
	REACH
	RESULT
	RESULT2

	playerNum = 2
	PASS      = -1
)

func (c Color) byte() byte {
	return byte(c - EMPTY)
	//return (byte(c) & 0x02) >> 1, byte(c) & 0x01
	//return  (byte(c)&0x02)>>1 + byte('0'), byte(c)&0x01 + byte('0')
}

type Board struct {
	size       int //棋盘大小
	long       int //size*size
	board      []Color
	zh         *zobrist // 用于判断全同
	neighbours [][]int  // 邻居
	blockCache []int    // 块缓存
	deadCache  []int    // 死子缓存
	idxPos     []int    // 索引每个pos在histPos中的index
	histPos    []int    // [:b.posNum]已落子位置
	bytes      []byte   // bytes
	colorNum   [3]int   // 棋盘上黑白空颜色数量
	moveNum    [3]int   // 走子总数
	takeNum    [3]int   // 被提子数
	score      [3]int   // 结果数单位是半子
	passNum    [3]int   // pass统计
	lastPos    int      // 最后一个落子位置
	posNum     int      // 棋盘上总子数，黑加白
}

func NewBoard(size int) *Board {
	long := size * size
	b := &Board{
		long:       long,
		size:       size,
		board:      make([]Color, long),
		zh:         newZobrist(long),
		histPos:    make([]int, long),
		idxPos:     make([]int, long),
		neighbours: make([][]int, long),
		blockCache: make([]int, long),
		deadCache:  make([]int, long),
		lastPos:    -1,
		bytes:      make([]byte, long+2), // x+y
	}

	b.colorNum[EMPTY] = long
	for i := 0; i < long; i++ {
		b.neighbours[i] = getNeighbours(i, size)
		b.histPos[i] = i
		b.idxPos[i] = i
		b.board[i] = EMPTY
	}
	return b
}
func (b *Board) GenGame() []byte {
	b.RandRun()
	b.CalcScore()
	return b.Bytes()
}

func (b *Board) Reset(bytes []byte) {
	b.lastPos = -1
	b.posNum = 0
	for i := WHITE; i <= BLACK; i++ {
		b.colorNum[i] = 0
		b.moveNum[i] = 0
		b.takeNum[i] = 0
		b.score[i] = 0
		b.passNum[i] = 0
	}
	long := b.long
	b.colorNum[EMPTY] = long
	for i := 0; i < long; i++ {
		b.board[i] = EMPTY
		b.histPos[i] = i
		b.idxPos[i] = i
	}
	b.bytes = bytes
	b.zh.reset()
}

// 获得pos所在的块，并测试pos所在的块能否到达c色
func (b *Board) isReachedColorAndGetBlock(c Color, pos int, noBlock bool) (bool, []int) {
	reach := false
	pc := b.board[pos]
	block := b.blockCache
	l := 0 // 被查找位置数量
	if pc == c {
		reach = true
		if noBlock {
			return reach, block[:0]
		}
	}
	b.board[pos] = REACH //搜过的位置进行染色
	block[l] = pos
	l++
loop:
	for i := 0; i < l; i++ {
		for _, nb := range b.neighbours[block[i]] {
			nbc := b.board[nb]
			if !reach && nbc == c {
				reach = true
				if noBlock {
					break loop
				}
			} else if nbc == pc {
				b.board[nb] = REACH //搜过的位置进行染色
				block[l] = nb
				l++
			}
		}
	}
	for _, p := range block[:l] {
		b.board[p] = pc
	}
	return reach, block[:l]
}

func (b *Board) RandRun() {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	player := BLACK
	prePass := false
	for {
		p := b.randPos(player, r)
		b.moveNum[player]++
		if p == PASS {
			b.passNum[player]++
			if prePass {
				break
			}
			prePass = true
		} else {
			b.lastPos = p
			prePass = false
			b.move(player, p)
		}
		player = reverseColor(player)
	}
	return
}

// 数子法
func (b *Board) CalcScore() int {
	wScore := 0
	bScore := 0
	for i := 0; i < b.long; i++ {
		ic := b.board[i]
		if ic == BLACK {
			bScore += 2
		} else if ic == WHITE {
			wScore += 2
		} else if ic == EMPTY {
			reachedBlack, _ := b.isReachedColorAndGetBlock(BLACK, i, true)
			reachedWhite, _ := b.isReachedColorAndGetBlock(WHITE, i, true)
			if reachedBlack && !reachedWhite {
				bScore += 2
			}
			if reachedWhite && !reachedBlack {
				wScore += 2
			}
			if reachedBlack && reachedWhite {
				wScore++
				bScore++
			}
		} else {
			panic("sss")
		}
	}
	b.score[WHITE] = wScore / 2
	b.score[BLACK] = bScore / 2
	sum := (bScore - wScore) / 2
	b.score[EMPTY] = sum
	return sum

}

func (b *Board) move(c Color, pos int) {
	// state
	b.changeBoard(c, pos)
	rc := reverseColor(c)
	// take dead pos
	dd := b.getDeadPos(reverseColor(c), pos)
	b.takeNum[rc] += len(dd)
	for _, p := range dd {
		b.changeBoard(EMPTY, p)
		//time.Sleep(10 * time.Second)
	}
	// history
	if _, has := b.zh.histHash[b.zh.hash]; has {
		fmt.Println("真是hash重复", b.zh.hash, len(b.zh.histHash))
		panic("局面重复")
	}
	b.zh.histHash[b.zh.hash] = true
}

func (b *Board) changeBoard(c Color, pos int) {
	b.colorNum[b.board[pos]]--
	b.colorNum[c]++
	b.setColorAndHash(c, pos)
	if c == EMPTY { // 提子
		b.posNum--
		b.switchIdx(b.idxPos[pos], b.posNum)
	} else if c == BLACK || c == WHITE { //落子
		b.switchIdx(b.idxPos[pos], b.posNum)
		b.posNum++
	} else {
		panic("不可能")
	}

	for i := 0; i < b.long; i++ {
		if i < b.posNum && b.board[b.histPos[i]] != WHITE && b.board[b.histPos[i]] != BLACK ||
			i >= b.posNum && b.board[b.histPos[i]] != EMPTY {
			fmt.Println("报错。。。了", i, b.posNum, b.board[b.histPos[i]])
			panic("不可能")
		}
	}
}

func (b *Board) setColorAndHash(c Color, pos int) {
	b.zh.hash = b.zh.calcBoardHash(b.zh.hash, pos, b.board[pos], c)
	b.board[pos] = c
}

func (b *Board) randPos(c Color, r *rand.Rand) int {
	start := b.posNum
	end := b.long - b.posNum
	illegalN := 0
	optimizedN := 0
	for start < b.long {
		idx := start + r.Intn(end)
		pos := b.histPos[idx]

		if b.board[pos] != EMPTY {
			fmt.Println("错误", pos, b.board[pos], b.posNum)
			panic("随机到已落子的位置")
		}

		b.board[pos] = c
		isIllegal := b.isIllegalPos(pos, c)
		b.board[pos] = EMPTY
		if isIllegal {
			illegalN++
		}
		optimized := b.optimized(c, pos)
		if optimized {
			optimizedN++
		}
		if isIllegal || optimized {
			b.switchIdx(idx, start)
			start++
			end--
		} else {
			return b.histPos[idx]
		}
	}
	//fmt.Println(b.posNum, b.takeNum, illegalN, optimizedN, float64(optimizedN)/float64(illegalN+1))
	return PASS
}

// optimized 需要确保不会优化掉最优解就可以
func (b *Board) optimized(c Color, pos int) bool {
	if b.posNum < 2 {
		return false
	}
	//return b.isMyEyeOrRivalBigEye(c, pos)
	return b.isMyTrueEye(c, pos)
	return false
}

func (b *Board) isMyTrueEye(c Color, pos int) bool {
	ok, _ := b.isReachedColorAndGetBlock(c, pos, true)
	ok2, _ := b.isReachedColorAndGetBlock(reverseColor(c), pos, true)
	rc := reverseColor(c)
	b.board[pos] = rc
	canTake := b.canTake(c, pos)
	b.board[pos] = EMPTY
	if ok && !ok2 && !canTake { // 只能到达我方颜色且对方落此位置不能提子
		return true // 我方眼位
	}
	return false
	//return !ok && ok2 && len(s) > 7 //对方大眼

}

//判断是否是己方的眼位或对方的大眼
func (b *Board) isMyEyeOrRivalBigEye(c Color, pos int) bool {
	ok, _ := b.isReachedColorAndGetBlock(c, pos, true)
	ok2, s := b.isReachedColorAndGetBlock(reverseColor(c), pos, false)
	if ok && !ok2 { // 只能到达我方颜色
		return true // 我方眼位
	}
	return !ok && ok2 && len(s) > 6 //对方大眼
}

func (b *Board) isIllegalPos(pos int, c Color) bool {
	return b.isSuicide(pos, c) || b.isSuperKO(c, pos)
}

// 判断是不是自杀
func (b *Board) isSuicide(pos int, c Color) bool {
	if ok, _ := b.isReachedColorAndGetBlock(EMPTY, pos, true); !ok { //无气
		if !b.canTake(reverseColor(c), pos) { // 不能提子
			return true // 无气且不能提子，自杀
		}
	}
	return false
}

func (b *Board) isSuperKO(c Color, pos int) bool {
	rc := reverseColor(c)
	hb := b.zh.calcBoardHash(b.zh.hash, pos, EMPTY, c)
	for _, p := range b.getDeadPos(rc, pos) {
		hb = b.zh.calcBoardHash(hb, p, rc, EMPTY)
	}
	return b.zh.histHash[hb]
}

func (b *Board) canTake(c Color, pos int) bool {
	for _, nb := range b.neighbours[pos] {
		if b.board[nb] == c {
			if ok, _ := b.isReachedColorAndGetBlock(EMPTY, nb, true); !ok {
				return true
			}
		}
	}
	return false
}

//获得pos相邻的c色死子
func (b *Board) getDeadPos(c Color, pos int) []int {
	first := true
	dead := b.deadCache
	l := 0
	for _, nb := range b.neighbours[pos] {
		if b.board[nb] == c {
			if ok, block := b.isReachedColorAndGetBlock(EMPTY, nb, false); !ok {
				if first {
					_ = append(dead[:0], block...)
					l = len(block)
				} else {
					n := l
				bl:
					for _, dp := range block {
						for i := 0; i < n; i++ {
							if dead[i] == dp {
								continue bl
							}
						}
						dead[l] = dp
						l++
					}
				}
				first = false
			}
		}
	}
	return dead[:l]
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

// 交换i1，i2位置上的值，并更新值的位置
func (b *Board) switchIdx(i1, i2 int) {
	v1 := b.histPos[i1]
	v2 := b.histPos[i2]
	b.histPos[i1], b.histPos[i2] = v2, v1
	b.idxPos[v1] = i2
	b.idxPos[v2] = i1

}

func reverseColor(c Color) Color {
	if c == BLACK {
		return WHITE
	} else {
		return BLACK
	}
}

func (b *Board) ToXY(pos int) string {
	return fmt.Sprintf("%d,%d", pos/b.size+1, pos%b.size+1)
}

func (b *Board) Display() {
	fmt.Printf("moveN:%3d ,moveB:%3d ,moveW:%3d ,posN:%4d ,posB:%4d ,posW:%4d\n", b.moveNum[WHITE]+b.moveNum[BLACK], b.moveNum[WHITE], b.moveNum[BLACK],
		b.posNum, b.colorNum[BLACK], b.colorNum[WHITE])
	fmt.Printf("takeN:%3d ,takeB:%3d ,takeW:%3d ,scrN:%3d ,scrB:%3d ,scrW:%3d\n", b.takeNum[WHITE]+b.takeNum[BLACK], b.takeNum[WHITE], b.takeNum[BLACK],
		b.score[EMPTY], b.score[BLACK], b.score[WHITE])
	fmt.Printf("passN:%3d ,passB:%3d ,passW:%3d ,boardhash:%d\n ", b.passNum[WHITE]+b.passNum[BLACK], b.passNum[WHITE], b.passNum[BLACK], b.zh.hash)
	fmt.Println()
	fmt.Print("  ")
	for i := 0; i < b.size; i++ {
		fmt.Printf("%2d", i+1)
	}
	for i := 0; i < b.long; i++ {
		if i%b.size == 0 {
			fmt.Printf("\n%2d", i/b.size+1)
		}
		if i == b.lastPos {
			if b.board[i] == BLACK {
				fmt.Printf(" X")
			} else if b.board[i] == WHITE {
				fmt.Printf(" Y")
			}
		} else if b.board[i] == BLACK {
			fmt.Printf(" B")
		} else if b.board[i] == WHITE {
			fmt.Print(" W")
		} else if b.board[i] == EMPTY {
			fmt.Print(" _")
		} else if b.board[i] == REACH {
			fmt.Print(" R", b.board[i])
		} else {
			fmt.Print(b.board[i])
		}
	}
	fmt.Printf("\n\n")
	//fmt.Println(b.Bytes())
}
func (c Color) String() string {
	return colorString[c]
}

func (b *Board) Bytes() []byte {
	bs := b.bytes
	l := b.long
	for i := 0; i < l; i++ {
		bs[i] = byte(b.board[i] - EMPTY)
	}
	score := b.score[EMPTY]
	bs[l] = byte(score >> 8)
	bs[l+1] = byte(score)
	return bs
}

func (b *Board) CheckError() error {
	if (b.moveNum[WHITE] + b.moveNum[BLACK] - b.passNum[WHITE] - b.passNum[BLACK] - b.posNum - b.takeNum[BLACK] - b.takeNum[WHITE]) != 0 {
		fmt.Println(b.moveNum, b.passNum, b.takeNum)
		return fmt.Errorf("走子，落子，和提子不对")
	}
	if (b.score[BLACK] + b.score[WHITE]) != b.long {
		return fmt.Errorf("得分不对")
	}
	if b.posNum != b.colorNum[WHITE]+b.colorNum[BLACK] || b.posNum+b.colorNum[EMPTY] != b.long {
		return fmt.Errorf("棋盘上的石头数量不对")
	}
	return nil
}
