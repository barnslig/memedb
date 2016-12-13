package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
)

var (
	configPath = flag.String("c", "config.json", "Config file path")
	reindexAll = flag.Bool("reindex", false, "Reindex all memes")
)

func parseConfig(filePath string) (cfg AppConfig, err error) {
	file, err := ioutil.ReadFile(filePath)
	if err != nil {
		return
	}

	err = json.Unmarshal(file, &cfg)
	if err != nil {
		return
	}

	return
}

func main() {
	flag.Parse()

	cfg, err := parseConfig(*configPath)
	if err != nil {
		log.Fatal(err)
	}

	app := CreateApp(cfg)

	if *reindexAll {
		app.reindexMemes()
		return
	}

	app.ListenAndServe()
}
