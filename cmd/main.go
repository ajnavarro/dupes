package main

import (
	"log"
	"os"

	"github.com/ajnavarro/dupes"
	"github.com/jessevdk/go-flags"
	bblfsh "gopkg.in/bblfsh/client-go.v2"
)

type Options struct {
	Path string `short:"p" long:"path" description:"Path to folder to analyze" required:"true"`
}

var options Options
var parser = flags.NewParser(&options, flags.Default)

func main() {
	if _, err := parser.Parse(); err != nil {
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			os.Exit(0)
		} else {
			os.Exit(1)
		}
	}

	client, err := bblfsh.NewClient("0.0.0.0:9432")
	if err != nil {
		panic(err)
	}

	p := dupes.NewParser(client, options.Path)

	res, err := p.Parse()
	if err != nil {
		panic(err)
	}

	for _, e := range res.Errs {
		log.Println("Error on file ", e.Filename, "ERROR:", e.Err)
	}

	for k, dups := range res.Dupes {
		log.Println("similar code found!", k)
		for _, d := range dups {
			log.Println("	- file: ", d.Filename, "lines:", d.LineFrom, d.LineTo)
		}
	}
}
