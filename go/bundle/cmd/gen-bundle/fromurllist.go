package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/WICG/webpackage/go/bundle"
)

func fromURLList(urlListFile string, es []*bundle.Exchange) ([]*bundle.Exchange, error) {
	input, err := os.Open(urlListFile)
	if err != nil {
		return nil, fmt.Errorf("Failed to open %q: %v", urlListFile, err)
	}
	defer input.Close()
	scanner := bufio.NewScanner(input)

	urls := make(map[string]bool)
	for _, e := range es {
		urls[e.Request.URL.String()] = true
	}

	for scanner.Scan() {
		rawURL := strings.TrimSpace(scanner.Text())
		// Skip blank lines and comments.
		if len(rawURL) == 0 || rawURL[0] == '#' {
			continue
		}

		if _, ok := urls[rawURL]; ok {
			log.Printf("Skipping %q", rawURL)
			continue
		}
		urls[rawURL] = true
		log.Printf("Processing %q", rawURL)

		parsedURL, err := url.Parse(rawURL)
		if err != nil {
			return nil, fmt.Errorf("Failed to parse URL %q: %v", rawURL, err)
		}
		resp, err := http.Get(rawURL)
		if err != nil {
			log.Printf("Failed to fetch %q: %v", rawURL, err)
			continue
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("Error reading response body of %q: %v", rawURL, err)
		}
		e := &bundle.Exchange{
			Request: bundle.Request{
				URL: parsedURL,
			},
			Response: bundle.Response{
				Status: resp.StatusCode,
				Header: resp.Header,
				Body: body,
			},
		}
		es = append(es, e)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("Error reading %q: %v", urlListFile, err)
	}

	return es, nil
}
