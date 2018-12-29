package main

import (
	"bytes"
	"flag"
	"io/ioutil"
	"log"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/andybalholm/cascadia"
	"github.com/mattn/go-zglob"
	"golang.org/x/net/html"
)

var (
	flagRoot  = flag.String("root", "./public", "root directory containing static html site")
	flagHTML  = flag.String("html", "**/*.html", "pattern for finding HTML files")
	flagDebug = flag.Bool("debug", false, "enable debug logging")
	flagIndex = flag.String("index", "index.html", "name of index.html file")
)

func main() {
	flag.Parse()

	if *flagHTML == "" {
		log.Fatalf("must specify html pattern")
	}
	if *flagDebug {
		log.Printf("using pattern %q", *flagHTML)
	}
	root := filepath.Clean(*flagRoot)
	if *flagDebug {
		log.Printf("root %q", root)
	}
	files, err := zglob.Glob(filepath.Join(root, *flagHTML))
	if err != nil {
		log.Fatalf("FAIL: %s", err)
	}
	uris := map[string]bool{}

	for _, f := range files {
		log.Printf("reading %s", f)
		uri := f[len(root):]
		if *flagIndex != "" && strings.HasSuffix(uri, *flagIndex) {
			uri = uri[:len(uri)-len(*flagIndex)]
		}
		log.Printf("uri %s", uri)
		uris[uri] = true
	}
	log.Printf("Found %d uris", len(uris))

	matcher, err := cascadia.Compile(`a[href]`)
	if err != nil {
		log.Fatalf("internal error compiling: %s", err)
	}
	for _, f := range files {
		log.Printf("reading %s", f)
		r, err := ioutil.ReadFile(f)
		if err != nil {
			log.Fatalf("FAIL: %s", err)
		}
		doc, err := html.Parse(bytes.NewReader(r))
		if err != nil {
			log.Fatalf("unable to parse html: %s", err)
		}
		for _, node := range matcher.MatchAll(doc) {
			for _, attr := range node.Attr {
				if attr.Key == "href" {
					link, err := url.Parse(attr.Val)
					if err != nil {
						log.Fatalf("unable to parse %q, %s", attr.Val, err)
					}
					if link.Scheme == "" && link.Host == "" {
						if !uris[link.Path] {
							log.Printf("    didn't find relative link: %q", link.Path)
						}
					}
				}
			}
		}
	}
}
