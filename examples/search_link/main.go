package main

import (
	"context"
	"fmt"
	"log"

	"stealthfetch/pkg/scrapegraph"
)

func main() {
	html := `<html><body><a href="https://example.com/docs">Docs</a><a href="https://example.com/blog">Blog</a></body></html>`
	graph := scrapegraph.NewSearchLinkGraph(scrapegraph.SearchLinkGraphConfig{
		Source:      html,
		MaxLinks:    10,
		FilterByLLM: false,
	})
	links, err := graph.Run(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	for _, link := range links {
		fmt.Println(link)
	}
}
