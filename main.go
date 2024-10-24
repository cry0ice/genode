package main

import (
	"flag"
	"log"
	"os"
	"sync"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/imroc/req/v3"
)

const userAgent = "Mozilla/5.0 (Windows; U; Windows NT 10.3;) AppleWebKit/603.5 (KHTML, like Gecko) Chrome/52.0.1446.309 Safari/603"

var (
	proxy  string
	output string
)

func init() {
	flag.StringVar(&proxy, "proxy", "", "Set proxy URL")
	flag.StringVar(&output, "output", "sub.txt", "Set output path")
	flag.Parse()
}

func main() {
	client := req.NewClient()
	client.SetUserAgent(userAgent)
	if proxy != "" {
		client.SetProxyURL(proxy)
	}
	sourcePath := flag.Arg(0)
	links, err := readURLs(sourcePath)
	if err != nil {
		log.Fatal(err)
	}
	set := mapset.NewSet[string]()
	chNodes := make(chan string, 100)
	go func() {
		outputFile, err := os.OpenFile(output, os.O_CREATE|os.O_WRONLY, os.ModePerm)
		if err != nil {
			log.Fatal(err)
		}
		defer outputFile.Close()
		for node := range chNodes {
			if _, err := outputFile.WriteString(node + "\n"); err != nil {
				log.Fatal(err)
			}
		}
	}()

	var wg sync.WaitGroup
	for _, link := range links {
		wg.Add(1)
		go func() {
			defer wg.Done()
			log.Println("fetch", link)
			if err := getNodes(set, chNodes, client, link); err != nil {
				log.Println("error on fetching", link, err)
			}
		}()
	}
	wg.Wait()
	close(chNodes)
}
