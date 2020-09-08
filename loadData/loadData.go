package loadData

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
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

func loadData() []urlEntitie {
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
	validUrltoDownload := onlyValidUrlEntities(urlToDownload)
	fmt.Println(validUrltoDownload)
	return validUrltoDownload
}
