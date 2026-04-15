package main

import (
	"context"
	"fmt"
	"log"

	"stealthfetch/pkg/scrapegraph"
)

func main() {
	html := `<html><body><h1>Lite Example</h1><p>Use the lightweight workflow for a quick extraction.</p></body></html>`
	g := scrapegraph.NewSmartScraperLiteGraph("Extract the page title and summary as JSON", html, nil, "")
	result, err := g.Run(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(result))
}
