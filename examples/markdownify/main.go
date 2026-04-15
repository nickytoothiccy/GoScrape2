package main

import (
	"context"
	"fmt"
	"log"

	"stealthfetch/pkg/scrapegraph"
)

func main() {
	html := `<html><body><h1>Example</h1><p>Convert this page to markdown.</p></body></html>`
	g := scrapegraph.NewMarkdownifyGraph(html, nil)
	out, err := g.Run(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(out)
}
