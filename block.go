package goGame

import "fmt"

type block struct {
	pNode *block
	cNode map[*block]bool
	//nNode []*block
	ps    map[int]bool
	color byte
}

type board struct {
	root       *block
	size       int
	long       int
	wallSize   int
	wallLong   int
	board      []byte
	pTob       []*block
	neighbours [][4]int
	blocks     []*block
	blocksNum  int
}

func newBoard(size int) *board {
	long := size * size
	wallSize := size + 2
	wallLong := wallSize * wallSize
	r := &block{
		cNode: make(map[*block]bool),
		ps:    make(map[int]bool, wallSize*4-4),
		color: WALL,
	}
	bpn := long / 4
	b := &board{
		size:     size,
		long:     long,
		wallSize: wallSize,
		wallLong: wallLong,
		root:     r,
		board:    make([]byte, wallLong),
		pTob:     make([]*block, wallLong),
		blocks:   make([]*block, bpn),
	}

	blkPool := make([]block, bpn)
	for i := 0; i < bpn; i++ {
		b.blocks[i] = &blkPool[i]
	}
	e := b.getBlock().init(r, EMPTY, 0, nil)
	e.ps = make(map[int]bool, long)
	r.cNode[e] = true
	for i := 0; i < wallLong; i++ {
		b.neighbours[i] = b.getNeighbours(i)
		if isWall(wallSize, i) {
			b.board[i] = WALL
			b.pTob[i] = r
			r.ps[i] = true
		} else {
			b.board[i] = EMPTY
			b.pTob[i] = e
			e.ps[i] = true
		}
	}

	return b
}

func (blk *block) init(pNode *block, color byte, p int, nNodes []*block) *block {
	blk.pNode = pNode
	blk.color = color
	blk.ps[p] = true
	//blk.nNode = append(blk.nNode, nNodes...)
	return blk

}
func (b *board) merge(blk, blk2 *block) {
	for p, _ := range blk2.ps {
		b.pTob[p] = blk
	}
	delete(blk.pNode.cNode, blk2)
	b.putBlock(blk2)
}
func (b *board) changeBlock(c byte, p int) {
	nb := b.neighbours[p]
	pn := b.pTob[p]
	delete(pn.ps, p)
	var snb = [4][]int{make([]int, 0, 4), make([]int, 0, 4), make([]int, 0, 4), make([]int, 0, 4)}
	snb = b.sameNb(nb, snb)
	snbC := snb[c]
	cl := len(snbC)
	//snbE := snb[EMPTY]
	//el := len(snbE)

	if cl == 0 { // 周围没有同色 , 新落子独立成块，要处理好块间关系
		blk := b.getBlock().init(nil, c, p, nil)
		if len(snb[WALL]) > 0 { // 靠墙
			blk.pNode = b.root
			b.root.cNode[blk] = true
		} else { // 不靠墙，新块的
			blk.pNode = pn
			pn.cNode[blk] = true
		}
	} else if cl == 1 { // 周围有一个同色，将新的颜色添加到同色的块中即可
		b.pTob[snbC[0]].ps[p] = true
	} else { // 周围有多个同色，由于落子，导致需要块合并
		fb := b.pTob[snbC[0]]
		for i := 1; i < cl; i++ { //合并同色块
			blk := b.pTob[snbC[i]]
			if blk != fb {
				b.merge(fb, blk) // 合并到第一个块色上
			}
		}
	}

	//分割白棋需要检测白点的联通性。

	// 处理原块去pos
	// 处理切empty块的问题

	// 新增独立块
	// 新增独立块并导致分割无色块
	// 沾附到自己块上
	// 沾附到自己块上并导致分割无色块（紧挨对手色或墙）
	// 挨到对手快上导致分割无色块
	// 链接自己色的块
	// 链接自己色的块 且导致分割无色块
	// 分割无色块
	//  提子总是多导致一个对手块变成无色块产生
	//b.pTob[p]
}
func isWall(size int, pi int) bool {
	long := size * size
	return pi < size || pi >= long-size || pi%size == 0 || pi%size == size-1
}

func (b *board) getBlock() *block {
	b.blocksNum--
	return b.blocks[b.blocksNum]
}

func (b *board) putBlock(blk *block) {
	b.blocks[b.blocksNum] = blk
	b.blocksNum++
}

func (b *board) Display() {
	//fmt.Printf("moveN:%3d ,moveB:%3d ,moveW:%3d ,posN:%3d ,posB:%3d ,posW:%3d\n", b.moveNum[WHITE]+b.moveNum[BLACK], b.moveNum[WHITE], b.moveNum[BLACK],
	//	b.long-b.colorNum[EMPTY], b.colorNum[BLACK], b.colorNum[WHITE])
	//fmt.Printf("takeN:%3d ,takeB:%3d ,takeW:%3d ,scrN:%3d ,scrB:%3d ,scrW:%3d\n", b.takeNum[WHITE]+b.takeNum[BLACK], b.takeNum[WHITE], b.takeNum[BLACK],
	//	b.score[EMPTY], b.score[BLACK], b.score[WHITE])
	//fmt.Printf("passN:%3d ,passB:%3d ,passW:%3d ,boardhash:%d\n ", b.passNum[WHITE]+b.passNum[BLACK], b.passNum[WHITE], b.passNum[BLACK], b.zh.hash)
	fmt.Println()
	fmt.Print("    ")
	for i := 0; i < b.size; i++ {
		fmt.Printf("%2d", i+1)
	}
	for i := 0; i < b.wallLong; i++ {
		if i%b.wallSize == 0 {
			if i == 0 || i == b.wallLong-b.wallSize {
				fmt.Printf("\n  ")
			} else {
				fmt.Printf("\n%2d", i/b.wallSize)
			}
		}
		//if i == b.lastPos {
		//	if b.board[i] == BLACK {
		//		fmt.Printf(" X")
		//	} else if b.board[i] == WHITE {
		//		fmt.Printf(" Y")
		//	}
		//} else
		if b.board[i] == BLACK {
			fmt.Printf(" B")
		} else if b.board[i] == WHITE {
			fmt.Print(" W")
		} else if b.board[i] == EMPTY {
			fmt.Print(" _")
		} else if b.board[i] == WALL {
			fmt.Print(" " + string(rune(0x2588)))
		} else {
			fmt.Print(b.board[i])
		}
	}
	fmt.Printf("\n\n")
	//fmt.Println(b.Bytes())
}

func (b *board) getNeighbours(pos int) (nb [4]int) {
	size := b.wallSize

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
	return nb
}

func (b *board) sameNb(nb [4]int, ret [4][]int) [4][]int {
	ret[b.board[nb[0]]] = append(ret[b.board[nb[0]]], nb[0])
	ret[b.board[nb[1]]] = append(ret[b.board[nb[1]]], nb[1])
	ret[b.board[nb[2]]] = append(ret[b.board[nb[2]]], nb[2])
	ret[b.board[nb[3]]] = append(ret[b.board[nb[3]]], nb[3])
	return ret
}
