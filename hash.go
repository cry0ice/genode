package main

import (
	"encoding/base64"
	"net/url"

	"github.com/Jeffail/gabs"
)

func hash(s string) (string, error) {
	u, err := url.Parse(s)
	if err != nil {
		return "", err
	}

	switch u.Scheme {
	case "vmess":
		bs, err := base64.StdEncoding.DecodeString(u.Host)
		if err != nil {
			return "", err
		}
		jsonObj, err := gabs.ParseJSON(bs)
		if err != nil {
			return "", err
		}
		jsonObj.Set("", "ps")
		return jsonObj.String(), nil
	default:
		u.Fragment = ""
		return u.String(), nil
	}
}
