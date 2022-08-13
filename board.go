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

var (
	colorString = []string{WHITE: "WHITE", EMPTY: "EMPTY", BLACK: "BLACK", WALL: "WALL", REACH: "REACH"}
)

const (
	WHITE byte = iota
	EMPTY
	BLACK
	WALL
	REACH

	playerNum = 2
)

type Board struct {
	size       int //棋盘大小
	long       int //size*size
	board      []byte
	zh         *zobrist // 用于判断全同
	neighbours [][]int  // 邻居
	blockCache []int    // 块缓存
	deadCache  []int    // 死子缓存
	idxPos     []int    // 索引每个pos在histPos中的index
	histPos    []int    // [b.emptyNum:]已落子位置
	colorNum   [3]int   // 棋盘上黑白空颜色数量
	moveNum    [3]int   // 走子总数
	takeNum    [3]int   // 被提子数
	score      [3]int   // 结果数单位是半子
	passNum    [3]int   // pass统计
	lastPos    int      // 最后一个落子位置
	simpleKo   [][2]int // 简单劫
	koIdx      int
}

func NewBoard(size int) *Board {
	long := size * size
	b := &Board{
		long:       long,
		size:       size,
		board:      make([]byte, long+2),
		zh:         newZobrist(long),
		histPos:    make([]int, long),
		idxPos:     make([]int, long),
		neighbours: make([][]int, long),
		blockCache: make([]int, long),
		deadCache:  make([]int, long),
		lastPos:    -1,
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
		bytes[i] = EMPTY
		b.histPos[i] = i
		b.idxPos[i] = i
	}
	b.board = bytes
	b.zh.reset()
}

// 获得pos所在的块，并测试pos所在的块能否到达c色
func (b *Board) isReachedColorAndGetBlock(c byte, pos int, noBlock bool) (bool, []int) {
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
		end := b.colorNum[EMPTY]
		if b.randMove(player, r, end) {
			b.passNum[player]++
			if prePass { // 两pass终局
				break
			}
			prePass = true
		} else {
			prePass = false
		}
		player = reverseColor(player)
	}
}

func (b *Board) move(c byte, pos int) {
	// state
	b.moveNum[c]++
	b.lastPos = pos
	b.changeBoard(c, pos) // 落子
	rc := reverseColor(c)
	deadPos := b.getDeadPos(rc, pos)
	b.takeNum[int(rc)] += len(deadPos)
	for _, p := range deadPos {
		b.changeBoard(EMPTY, p) // 提子
	}

	//测试
	b.checkSuperKO(c, pos)
	b.CheckError(c, pos)

	b.zh.histHash[b.zh.hash] = true
}

func (b *Board) checkSuperKO(c byte, pos int) {
	if b.takeNum[BLACK] <= 0 && b.takeNum[WHITE] <= 0 {
		return
	}
	if b.zh.histHash[b.zh.hash] {
		fmt.Println("劫出现了，", colorString[c], b.ToXY(pos))
	}
}

func (b *Board) changeBoard(c byte, pos int) {
	b.colorNum[b.board[pos]]--
	b.zh.hash = b.zh.calcBoardHash(b.zh.hash, pos, b.board[pos], c)
	b.board[pos] = c
	b.switchIdx(b.idxPos[pos], b.colorNum[EMPTY])
	b.colorNum[c]++
}

func (b *Board) randMove(c byte, r *rand.Rand, end int) (isPass bool) {
	for end > 0 {
		idx := r.Intn(end)
		pos := b.histPos[idx]

		if b.isIllegalPos(c, pos) || b.optimized(c, pos) {
			end--
			b.switchIdx(idx, end)
		} else {
			b.move(c, pos)
			return false
		}
	}
	//fmt.Println(b.posNum, b.takeNum, illegalN, optimizedN, float64(optimizedN)/float64(illegalN+1))
	return true
}

func (b *Board) searchTree(c byte, end int) int {
	end--
	for end > 0 {
		pos := b.histPos[end]
		if b.isIllegalPos(c, pos) || b.optimized(c, pos) {
			end--
		} else {
			b.simpleMove(c, pos, end)
		}

	}
	return 0
}

func (b *Board) simpleMove(c byte, pos int, end int) int {
	b.board[pos] = c
	b.switchIdx(b.idxPos[pos], end)
	end--
	rc := reverseColor(c)
	dead := b.getDeadPos(rc, pos)
	for _, p := range dead {
		b.board[pos] = EMPTY
		b.switchIdx(b.idxPos[p], end)
		end--
	}
	//b.searchTree(rc, end)
	return 0
}

func (b *Board) isIllegalPos(c byte, pos int) bool {
	b.board[pos] = c
	defer func() { b.board[pos] = EMPTY }()
	//isSuicide := b.isSuicide(c, pos)
	//return isSuicide
	return b.isSuicide(c, pos) || b.isSuperKO(c, pos)
}

// 判断是不是自杀
func (b *Board) isSuicide(c byte, pos int) bool {
	ok, _ := b.isReachedColorAndGetBlock(EMPTY, pos, true)
	return !ok && !b.canTake(reverseColor(c), pos) // 无气 && 不能提子
}

func (b *Board) isSuperKO(c byte, pos int) bool {
	if b.takeNum[BLACK] <= 0 && b.takeNum[WHITE] <= 0 {
		return false
	}
	rc := reverseColor(c)
	hb := b.zh.calcBoardHash(b.zh.hash, pos, EMPTY, c)
	for _, p := range b.getDeadPos(rc, pos) {
		hb = b.zh.calcBoardHash(hb, p, rc, EMPTY)
	}
	return b.zh.histHash[hb]
}

// optimized 需要确保优化掉的局面不会比pass一手差
func (b *Board) optimized(c byte, pos int) bool {
	if b.moveNum[BLACK] < 100 {
		return false
	}
	return b.isMyTrueEye(c, pos)
}

func (b *Board) isMyTrueEye(c byte, pos int) bool {
	rc := reverseColor(c)
	ok, _ := b.isReachedColorAndGetBlock(c, pos, true)
	ok2, _ := b.isReachedColorAndGetBlock(rc, pos, true)
	//ok2, s := b.isReachedColorAndGetBlock(rc, pos, false)
	b.board[pos] = rc
	canTake := b.canTake(c, pos)
	b.board[pos] = EMPTY
	//return ok && !ok2 && !canTake || !ok && ok2 && len(s) > 7 // 只能到达我方颜色且对方落此位置不能提子 或对方大眼
	return ok && !ok2 && !canTake // 只能到达我方颜色且对方落此位置不能提子
}

func (b *Board) canTake(c byte, pos int) bool {
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
func (b *Board) getDeadPos(c byte, pos int) []int {
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

// 交换i1，i2位置上的值，并更新值的位置
func (b *Board) switchIdx(i1, i2 int) {
	// 拿到位置的值
	v1 := b.histPos[i1]
	v2 := b.histPos[i2]
	// 换位置
	b.histPos[i1], b.histPos[i2] = v2, v1
	//构建新索引
	b.idxPos[v1] = i2
	b.idxPos[v2] = i1

}

func reverseColor(c byte) byte {
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
	fmt.Printf("moveN:%3d ,moveB:%3d ,moveW:%3d ,posN:%3d ,posB:%3d ,posW:%3d\n", b.moveNum[WHITE]+b.moveNum[BLACK], b.moveNum[WHITE], b.moveNum[BLACK],
		b.long-b.colorNum[EMPTY], b.colorNum[BLACK], b.colorNum[WHITE])
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

func (b *Board) Bytes() []byte {
	bs := b.board
	l := b.long
	score := b.score[EMPTY] + b.long
	bs[l] = byte(score >> 8)
	bs[l+1] = byte(score)
	return bs
}

func (b *Board) CheckError(c byte, pos int) {
	for i := 0; i < b.long; i++ {
		if i >= b.colorNum[EMPTY] && b.board[b.histPos[i]] != WHITE && b.board[b.histPos[i]] != BLACK ||
			i < b.colorNum[EMPTY] && b.board[b.histPos[i]] != EMPTY {
			fmt.Println("报错。。。了,empty 数量不对", i, b.colorNum[EMPTY], b.board[b.histPos[i]], c, b.ToXY(pos), b.ToXY(b.histPos[i]))
			b.Display()
			panic("不可能")
		}
	}
	if (b.moveNum[WHITE] + b.moveNum[BLACK] - b.colorNum[WHITE] - b.colorNum[BLACK] - b.takeNum[BLACK] - b.takeNum[WHITE]) != 0 {
		fmt.Println(b.moveNum, b.passNum, b.takeNum)
		fmt.Println("走子，落子，和提子不对")
		b.Display()
		panic("不可能")
	}

	if (b.score[BLACK]+b.score[WHITE]) != b.long && b.score[BLACK] != 0 && b.score[WHITE] != 0 && b.score[EMPTY] != 0 {
		fmt.Println("得分不对", b.score[BLACK], b.score[WHITE], b.score[EMPTY])
		panic("不可能")
	}

	if b.colorNum[BLACK]+b.colorNum[WHITE]+b.colorNum[EMPTY] != b.long {
		b.Display()

		fmt.Printf("棋盘上的石头空白总数量不对,%d,%d,%d,%d\n", b.colorNum[BLACK], b.colorNum[WHITE], b.colorNum[EMPTY], b.long)
		panic("不可能")
	}
	wr := b.passNum[WHITE] + b.colorNum[WHITE] + b.takeNum[WHITE]
	br := b.passNum[BLACK] + b.colorNum[BLACK] + b.takeNum[BLACK]
	if b.board[b.lastPos] == BLACK {
		wr += 1
	}
	if br != wr {
		b.Display()
		fmt.Println("两人手数不对")
		panic("不可能")
	}
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
			//reachedBlack, _ := b.isReachedColorAndGetBlock(BLACK, i, true)
			//reachedWhite, _ := b.isReachedColorAndGetBlock(WHITE, i, true)
			//if reachedBlack && !reachedWhite {
			if b.board[b.neighbours[i][0]] == BLACK {
				bScore += 2
			}
			//if reachedWhite && !reachedBlack {
			if b.board[b.neighbours[i][0]] == WHITE {
				wScore += 2
			}
			//if reachedBlack && reachedWhite {
			//	wScore++
			//	bScore++
			//}
		}
	}
	b.score[WHITE] = wScore / 2
	b.score[BLACK] = bScore / 2
	sum := (bScore - wScore) / 2
	b.score[EMPTY] = sum
	return sum

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
