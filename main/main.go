package main

import (
	"fmt"
	"goGame"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

func main() {
	var num int64 = 2
	sum := num
	size := 19
	long := size * size
	long2 := long + 2
	n := 50
	in := make(chan []byte, n)
	out := make(chan []byte, 4*n)
	var wg sync.WaitGroup
	done := make(chan struct{})
	go func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Println("Recovered in f", r)
			}
		}()
		err := openFile(long, "goGame.data", in, out, done)
		if err != nil {
			fmt.Println(err)
		}
	}()
	for i := 0; i < 3*n; i++ {
		out <- make([]byte, long2)
	}
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()
			g := goGame.NewBoard(size)
			for atomic.AddInt64(&num, -1) >= 0 {
				in <- g.GenGame()
				g.Reset(<-out)
			}
		}()
	}
	go func() {
		wg.Wait()
		close(in)
	}()
	perNum := sum
	for {
		select {
		case <-done:
			os.Exit(0)
		default:
			time.Sleep(time.Second * 1)
			fmt.Println(1-float64(num)/float64(sum), perNum-num)
			perNum = num
		}
	}
}

func openFile(l int, name string, in, out chan []byte, done chan struct{}) error {
	xf, err := os.OpenFile(name+"x", os.O_CREATE|os.O_RDWR, 0700)
	if err != nil {
		return err
	}
	yf, err := os.OpenFile(name+"y", os.O_CREATE|os.O_RDWR, 0700)
	if err != nil {
		return err
	}
	defer xf.Close()
	defer yf.Close()
breakLabel:
	for {
		select {
		case data, ok := <-in:
			if ok {
				//fmt.Println(data)
				_, err := xf.Write(data[:l])
				if err != nil {
					fmt.Println(err)
				}
				_, err = yf.Write(data[l:])
				if err != nil {
					fmt.Println(err)
				}
				out <- data
			} else {
				break breakLabel
			}
		}
	}
	done <- struct{}{}
	return nil
}
