package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

type UrlResponse struct {
	Url        string        `json:"url"`
	StatusCode int           `json:"status_code"`
	ResTime    time.Duration `json:"duration"`
}

func main() {
	var wg sync.WaitGroup
	resChan := make(chan UrlResponse)

	urls, err := RetrieveUrls("urls.txt")
	if err != nil {
		log.Printf("Error retrieving urls : %v", err)
	}

	for _, u := range urls {
		wg.Add(1)
		go func(url string) {
			defer wg.Done()
			SendReq(url, resChan)
		}(u)
	}

	go func() {
		wg.Wait()
		close(resChan)
	}()

	for url := range resChan {

		WriteToFile("results.txt", url)

		fmt.Printf("Url : %s, Status Code : %d, Response Time : %d \n", url.Url, url.StatusCode, url.ResTime)
	}

}

func RetrieveUrls(url string) ([]string, error) {
	urls, err := os.ReadFile(url)
	if err != nil {
		return []string{}, err
	}

	sliceOfUrls := strings.Split(string(urls), "\n")

	return sliceOfUrls, nil
}

func SendReq(url string, c chan<- UrlResponse) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Printf("Error : %v", err)
	}

	client := &http.Client{
		Timeout: time.Second * 5,
	}
	start := time.Now()
	res, err := client.Do(req)
	if err != nil {
		log.Printf("Network error : %v", err)
	}

	if res != nil {
		defer res.Body.Close()
	} else {
		log.Printf("Response is nil for URL : %s", url)
	}

	duration := time.Since(start)

	if res.StatusCode != http.StatusOK {
		log.Printf("GET request failed : %d", res.StatusCode)
	}

	if res != nil {
		c <- UrlResponse{
			Url:        url,
			StatusCode: res.StatusCode,
			ResTime:    time.Duration(duration.Milliseconds()),
		}
	} else {
		c <- UrlResponse{}
	}
}

func WriteToFile(filename string, uRes UrlResponse) {

	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("Error opening file: %v", err)
		return
	}
	defer file.Close()

	data, err := json.MarshalIndent(uRes, "", "    ")
	if err != nil {
		log.Printf("Error in encoding file : %v", err)
		return
	}

	if _, err := file.Write(append(data, '\n')); err != nil {
		log.Printf("Error inserting data into file: %v", err)
	}
}
