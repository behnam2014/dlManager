package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

type urlEntitie struct {
	Url   string
	Path  string
	State bool
}

func tagValidUrl(urlEntities *[]urlEntitie) {
	var cleanUrls []urlEntitie
	for _, urlItem := range *urlEntities {
		resp, err := http.Head(urlItem.Url)
		if err != nil {
			urlItem.State = false
			log.Println(err)
		} else if resp.StatusCode != http.StatusOK {
			urlItem.State = false
			log.Println(err)
		} else {
			urlItem.State = true
		}
		cleanUrls = append(cleanUrls, urlItem)
	}
	*urlEntities = cleanUrls
}

func onlyValidUrlEntities(entities []urlEntitie) []urlEntitie {
	var entitiesWithValidUrl []urlEntitie
	for _, v := range entities {
		if v.State == true {
			entitiesWithValidUrl = append(entitiesWithValidUrl, v)
		}
	}
	return entitiesWithValidUrl
}

type downloadWriter struct {
	fileName string
	total    uint64
	size     uint64
}

func (wc *downloadWriter) Write(p []byte) (int, error) {
	n := len(p)
	wc.total += uint64(n)
	fmt.Printf("\rDownloading %s... %f %c ", wc.fileName, (float64(wc.total)/float64(wc.size))*100, '%')
	return n, nil
}

func Download(url string, path string, wg *sync.WaitGroup) error {
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
	start := time.Now()
	counter := &downloadWriter{fileName: path, total: 0, size: uint64(size)}
	if _, err = io.Copy(out, io.TeeReader(resp.Body, counter)); err != nil {
		_ = out.Close()
		return err
	}

	_ = out.Close()
	elapsed := time.Since(start)
	log.Printf("Download completed in %s", elapsed)
	return nil
}


func loadData() (chan urlEntitie,config) {
	dat, err := ioutil.ReadFile("config/FilesToDownload.json")
	var urlToDownload []urlEntitie
	if err != nil {
		log.Println(err)
	} else {
		json.Unmarshal(dat, &urlToDownload)
	}
	log.Println("\n----------------", "\n", "Here is the list of corrupted urls:")
	tagValidUrl(&urlToDownload)
	log.Println("\n----------------", "\n", "Here is the list of urls to download:")
	validUrlstoDownload := onlyValidUrlEntities(urlToDownload)
	fmt.Println(validUrlstoDownload)
	urlEntitieChanel := make(chan urlEntitie)
	go func() {
		for _, urlItem := range validUrlstoDownload {
			urlEntitieChanel <- urlItem
		}
		close(urlEntitieChanel)
	}()
	var conf config
	datConfig, errConfig := ioutil.ReadFile("config/maxGoRoutine.json")
	if errConfig != nil {
		log.Println(errConfig)
	} else {
		json.Unmarshal(datConfig, &conf)
	}

	return urlEntitieChanel, conf
}

type config struct{
	MaxGoRoutine int
}


func main() {
	urlEntitieChanel,conf := loadData()
	log.Println("\n----------------", "\n", "Progress of downloads:")

	maxGoRoutine := conf.MaxGoRoutine
	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

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
					Download(valueFromChanel.Url, valueFromChanel.Path, &wg)
					time.Sleep(time.Millisecond*2)
				}

			}
		}()
	}
	wg.Wait()

	fmt.Println("Download Finished")
}
