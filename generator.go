package main

import (
	"bufio"
	"encoding/base64"
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
	resp, err := client.R().Get(u)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var nodes []string
	decoder := base64.NewDecoder(base64.StdEncoding, resp.Body)
	scanner := bufio.NewScanner(decoder)
	for scanner.Scan() {
		if text := scanner.Text(); text != "" {
			parsedUrl, err := url.Parse(text)
			if err != nil {
				continue
			}
			parsedUrl.Fragment = ""
			parsedUrl.RawFragment = ""
			noFragmentUrl := parsedUrl.String()

			if set.ContainsOne(noFragmentUrl) {
				continue
			}
			set.Add(noFragmentUrl)
			nodes = append(nodes, text)
		}
	}

	return nodes, scanner.Err()
}

func writeToFile(nodes []string, outputDir string) error {
	protocolMap := make(map[string]io.Writer)

	allList, err := os.OpenFile(path.Join(outputDir, "all.txt"), os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		return err
	}
	defer allList.Close()
	allListEncoder := base64.NewEncoder(base64.StdEncoding, allList)
	defer allListEncoder.Close()

	for _, node := range nodes {
		allListEncoder.Write([]byte(node + "\n"))

		parsedUrl, err := url.Parse(node)
		if err != nil {
			continue
		}
		writer, ok := protocolMap[parsedUrl.Scheme]
		if !ok {
			f, err := os.OpenFile(path.Join(outputDir, parsedUrl.Scheme+".txt"), os.O_CREATE|os.O_WRONLY, os.ModePerm)
			if err != nil {
				continue
			}
			defer f.Close()
			encoder := base64.NewEncoder(base64.StdEncoding, f)
			defer encoder.Close()
			protocolMap[parsedUrl.Scheme] = encoder
			writer = encoder
		}

		writer.Write([]byte(node + "\n"))
	}

	return nil
}
