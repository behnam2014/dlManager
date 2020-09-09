package main

import (
	"context"
	"fmt"
	"goland/dlManagerV1/loadData"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

var ld = loadData.NewLoadData()

type downloadWriter struct {
	fileName    string
	total       uint64
	size        uint64
	startTimeDl time.Time
	printIndex  int64
	wg          *sync.WaitGroup
	printChan   chan<- reportProgress
}

type reportProgress struct {
	printIndex   int64
	reportString string
}

func (wc *downloadWriter) Write(p []byte) (int, error) {
	n := len(p)
	wc.total += uint64(n)
	elapsed := time.Since(wc.startTimeDl)
	reportString := fmt.Sprintf("Downloading %s... TIME: %s, Compeleted: %2.2f %%", wc.fileName, elapsed, (float64(wc.total)/float64(wc.size))*100)
	wc.printChan <- reportProgress{printIndex: wc.printIndex, reportString: reportString}
	return n, nil
}

func Download(url string, path string, wg *sync.WaitGroup, printIndex int64, printChan chan<- reportProgress) error {
	out, err := os.Create(path)
	if err != nil {
		return err
	}

	resp, err := http.Get(url)
	if err != nil {
		_ = out.Close()
		return err
	}
	defer resp.Body.Close()

	size, err := strconv.Atoi(resp.Header.Get("Content-Length"))

	counter := &downloadWriter{fileName: path, total: 0, size: uint64(size), startTimeDl: time.Now(), wg: wg, printChan: printChan, printIndex: printIndex-1}
	if _, err = io.Copy(out, io.TeeReader(resp.Body, counter)); err != nil {
		_ = out.Close()
		return err
	}

	_ = out.Close()

	return nil
}

func main() {
	printChan := make(chan reportProgress, 3)
	urlEntitieChanel := ld.UrlEntitieChanel
	conf := ld.Conf
	log.Println("\n----------------", "\n", "Progress of downloads:")

	maxGoRoutine := conf.MaxGoRoutine
	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var counter int64
	upCode:="\u001b[%dA"
	go func() {
		reports := make([]string, ld.NumOfUrls)
		for v := range printChan {
			//fmt.Println(v)
			reports[v.printIndex] = v.reportString
			fmt.Printf(upCode,ld.NumOfUrls)
			for _,j:= range reports{
				fmt.Println(j)
			}
		}
	}()
	for i := 0; i < maxGoRoutine; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case valueFromChanel, ok := <-urlEntitieChanel:
					if !ok {
						return
					}
					atomic.AddInt64(&counter, 1)
					Download(valueFromChanel.Url, valueFromChanel.Path, &wg, atomic.LoadInt64(&counter), printChan)

				}

			}
		}()
	}
	wg.Wait()
	time.Sleep(time.Millisecond)
	fmt.Println("Download Finished")

}
