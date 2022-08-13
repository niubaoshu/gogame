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
	trainFile := "goGame_train.data"
	testFile := "goGame_test.data"
	var trainNum int64 = 100000
	var testNum int64 = 100000
	sum := trainNum + testNum
	d2 := make(chan struct{})

	go func() {
		done := make(chan struct{})
		genData(trainFile, 19, &trainNum, done)
		trainNum = 0
		<-done
		genData(testFile, 19, &testNum, done)
		d2 <- <-done
	}()

	perNum := sum
	for {
		select {
		case <-d2:
			os.Exit(0)
		default:
			time.Sleep(time.Second * 1)
			s := trainNum + testNum
			fmt.Println(1-float64(s)/float64(sum), perNum-s)
			perNum = s
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

func genData(fileName string, size int, num *int64, done chan struct{}) {
	var wg = sync.WaitGroup{}

	long := size * size
	long2 := long + 2
	n := 20
	in := make(chan []byte, 2*n)
	out := make(chan []byte, 2*n)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Println("Recovered in f", r)
			}
		}()
		err := openFile(long, fileName, in, out, done)
		if err != nil {
			fmt.Println(err)
		}
	}()
	for i := 0; i < n; i++ {
		b := make([]byte, long2)
		for i := 0; i < long2; i++ {
			b[i] = goGame.EMPTY
		}
		out <- b
	}
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func() {
			g := goGame.NewBoard(size)
			defer func() {
				wg.Done()
				//if err := recover(); err != nil {
				//	g.Display()
				//	runtime.StartTrace()
				//	fmt.Println(err)
				//}
			}()
			for atomic.AddInt64(num, -1) >= 0 {
				in <- g.GenGame()
				//g.Display()
				g.Reset(<-out)
			}
		}()
	}
	wg.Wait()
	close(in)
}
