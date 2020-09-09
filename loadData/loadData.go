package loadData

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

type LoadData struct {
	UrlEntitieChanel chan urlEntitie
	Conf             config
}

func New() LoadData {
	var ld  LoadData
	ld.UrlEntitieChanel, ld.Conf =loadData()
	return ld
}

type urlEntitie struct {
	Url   string
	Path  string
	State bool
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

type config struct {
	MaxGoRoutine int
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

func loadData() (chan urlEntitie, config) {
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