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
	var num int64 = 10000000
	sum := num
	size := 19
	n := 100
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
		err := openFile("goGame.data", in, out, done)
		if err != nil {
			fmt.Println(err)
		}
	}()
	for i := 0; i < 4*n; i++ {
		out <- make([]byte, 2*size*size)
	}
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()
			g := goGame.NewBoard(size)
			for atomic.AddInt64(&num, -1) >= 0 {
				g.RandRun(false)
				if err := g.CheckError(); err != nil {
					fmt.Println(err)
				} else {
					in <- g.Bytes()
				}
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
			fmt.Println(float64(num)/float64(sum), perNum-num)
			perNum = num
		}
	}
}

func openFile(name string, in, out chan []byte, done chan struct{}) error {
	f, err := os.OpenFile(name, os.O_CREATE|os.O_RDWR, 0700)
	if err != nil {
		return err
	}
	defer f.Close()
breakLabel:
	for {
		select {
		case data, ok := <-in:
			if ok {
				//fmt.Println(data)
				//_, err := f.Write(data)
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
