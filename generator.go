package main

import (
	"bufio"
	"encoding/base64"
	"io"
	"os"
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

func getNodes(set mapset.Set[string], chNodes chan string, client *req.Client, u string) error {
	resp, err := client.R().Get(u)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	fullBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	clearBytes := make([]byte, base64.StdEncoding.DecodedLen(len(fullBytes)))
	_, err = base64.StdEncoding.Decode(clearBytes, fullBytes)
	if err != nil {
		clearBytes = fullBytes
	}

	splited := strings.Split(strings.ReplaceAll(string(clearBytes), "\r\n", "\n"), "\n")
	for _, link := range splited {
		hash, err := hash(link)
		if err != nil {
			continue
		}
		if set.ContainsOne(hash) {
			continue
		}
		set.Add(hash)
		chNodes <- link
	}
	return nil
}
