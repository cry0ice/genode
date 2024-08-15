package main

import (
	"fmt"
	"time"

	"github.com/deckarep/golang-set/v2"
	"github.com/imroc/req/v3"
	"github.com/spf13/cobra"
)

const userAgent = "Mozilla/5.0 (Windows; U; Windows NT 10.3;) AppleWebKit/603.5 (KHTML, like Gecko) Chrome/52.0.1446.309 Safari/603"

var (
	outputDir string
	proxy     string
)

var root = &cobra.Command{
	Use:     "genode",
	Short:   "Generate node list",
	Version: "0.1.0",
	RunE: func(cmd *cobra.Command, args []string) error {
		urls, err := readURLs(args[0])
		if err != nil {
			return err
		}

		client := req.C().SetUserAgent(userAgent).SetTimeout(5 * time.Second).SetTLSFingerprintSafari().SetProxyURL(proxy)
		set := mapset.NewSet[string]()

		nodes := mapset.NewSet[string]()

		for _, url := range urls {
			fmt.Println("fetch", url)
			ns, err := getNodes(set, client, url)
			if err != nil {
				fmt.Println(err)
				continue
			}
			nodes.Append(ns...)
		}

		return writeToFile(nodes, outputDir)
	},
}

func init() {
	root.Flags().StringVarP(&outputDir, "output", "o", "", "Output directory")
	root.Flags().StringVar(&proxy, "proxy", "", "Proxy URL")
}
