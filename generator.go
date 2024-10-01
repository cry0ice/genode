package main

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"io"
	"net/url"
	"os"
	"path"
	"strings"
	"text/template"
	"time"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/imroc/req/v3"
)

func formatDate(layout string) string {
	return time.Now().Format(layout)
}

func readURLs(name string) ([]string, error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	tmpl := template.New("").Funcs(template.FuncMap{
		"date": formatDate,
	})

	var urls []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if text := scanner.Text(); text != "" {
			t, err := tmpl.Parse(text)
			if err != nil {
				continue
			}
			var builder strings.Builder
			if err := t.Execute(&builder, nil); err != nil {
				continue
			}

			urls = append(urls, builder.String())
		}
	}

	return urls, scanner.Err()
}

func getNodes(set mapset.Set[string], client *req.Client, u string) ([]string, error) {
	cutted, found := strings.CutPrefix(u, "clear:")

	var reader io.Reader
	if found {
		resp, err := client.R().Get(cutted)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		reader = resp.Body
	} else {
		resp, err := client.R().Get(u)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		decoder := base64.NewDecoder(base64.StdEncoding, resp.Body)
		reader = decoder
	}

	var nodes []string
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		if text := scanner.Text(); text != "" {
			hash, err := hash(text)
			if err != nil {
				fmt.Println(err)
				continue
			}

			if set.ContainsOne(hash) {
				continue
			}
			set.Add(hash)
			nodes = append(nodes, text)
		}
	}

	return nodes, scanner.Err()
}

func writeToFile(nodes mapset.Set[string], outputDir string) error {
	protocolMap := make(map[string]io.Writer)

	allList, err := os.OpenFile(path.Join(outputDir, "all.txt"), os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		return err
	}
	defer allList.Close()
	allListEncoder := base64.NewEncoder(base64.StdEncoding, allList)
	defer allListEncoder.Close()

	for node := range nodes.Iter() {
		allListEncoder.Write([]byte(node + "\n"))

		parsedURL, err := url.Parse(node)
		if err != nil {
			continue
		}
		writer, ok := protocolMap[parsedURL.Scheme]
		if !ok {
			f, err := os.OpenFile(path.Join(outputDir, parsedURL.Scheme+".txt"), os.O_CREATE|os.O_WRONLY, os.ModePerm)
			if err != nil {
				continue
			}
			defer f.Close()
			encoder := base64.NewEncoder(base64.StdEncoding, f)
			defer encoder.Close()
			protocolMap[parsedURL.Scheme] = encoder
			writer = encoder
		}

		writer.Write([]byte(node + "\n"))
	}

	return nil
}
