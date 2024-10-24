package main

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"io"
	"log"
	"os"
	"regexp"
	"strings"
	"text/template"
	"time"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/imroc/req/v3"
)

const checkBase64Pattern = "^[a-zA-Z0-9+/]*={0,3}$"

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
	isBase64, err := regexp.Match(checkBase64Pattern, fullBytes)
	if err != nil {
		return err
	}

	buffer := bytes.NewBuffer(fullBytes)
	var reader io.Reader
	if isBase64 {
		reader = base64.NewDecoder(base64.StdEncoding, buffer)
	} else {
		reader = buffer
	}

	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		text := scanner.Text()
		hashed, err := hash(text)
		if err != nil {
			log.Println(err)
			continue
		}
		if set.ContainsOne(hashed) {
			continue
		}
		set.Add(hashed)
		chNodes <- text
	}

	return scanner.Err()
}
