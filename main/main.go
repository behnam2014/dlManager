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

var ld = loadData.New()

type downloadWriter struct {
	fileName    string
	total       uint64
	size        uint64
	printIndex  int64
	startTimeDl time.Time
	wg *sync.WaitGroup
	reportString string
}

func (wc *downloadWriter) Write(p []byte) (int, error) {
	n := len(p)
	wc.total += uint64(n)
	elapsed := time.Since(wc.startTimeDl)
	wc.reportString=fmt.Sprintf("\rDownloading %s... TIME: %s, Compeleted: %2.2f %c", wc.fileName, elapsed, (float64(wc.total)/float64(wc.size))*100,'%')
	fmt.Printf(wc.reportString)
	return n, nil
}

func Download(url string, path string, wg *sync.WaitGroup, printIndex int64) error {
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

	counter := &downloadWriter{fileName: path, total: 0, size: uint64(size), printIndex: printIndex, startTimeDl: time.Now(), wg: wg}
	defer fmt.Println(counter.reportString)
	if _, err = io.Copy(out, io.TeeReader(resp.Body, counter)); err != nil {
		_ = out.Close()
		return err
	}

	_ = out.Close()

	return nil
}

func main() {

	urlEntitieChanel := ld.UrlEntitieChanel
	conf := ld.Conf
	log.Println("\n----------------", "\n", "Progress of downloads:")

	maxGoRoutine := conf.MaxGoRoutine
	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var counter int64
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
					Download(valueFromChanel.Url, valueFromChanel.Path, &wg, atomic.LoadInt64(&counter))
				}

			}
		}()
	}
	wg.Wait()
	time.Sleep(time.Millisecond)
	fmt.Println("Download Finished")

}
