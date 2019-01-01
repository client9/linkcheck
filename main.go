package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"time"

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

// Severity of linter message
type Severity string

// Linter message severity levels.
const (
	Error   Severity = "error"
	Warning Severity = "warning"
)

type Issue struct {
	Linter   string   `json:"linter"`
	Severity Severity `json:"severity"`
	Path     string   `json:"path"`
	Line     int      `json:"line"`
	Col      int      `json:"col"`
	Message  string   `json:"message"`
}

func printIssue(i Issue) {
	log.Printf("%s: %s", i.Path, i.Message)
}

func IssueError(format string, a ...interface{}) Issue {
	return Issue{
		Severity: Error,
		Message:  fmt.Sprintf(format, a...),
	}
}

func IssueWarning(format string, a ...interface{}) Issue {
	return Issue{
		Severity: Warning,
		Message:  fmt.Sprintf(format, a...),
	}
}

func getHref(node *html.Node) string {
	for _, attr := range node.Attr {
		if attr.Key == "href" {
			return attr.Val
		}
	}
	return ""
}

type LinkCheck struct {
	matcher  cascadia.Selector
	external map[string]bool
}

func NewLinkCheck() *LinkCheck {
	return &LinkCheck{
		matcher:  cascadia.MustCompile(`a[href]`),
		external: make(map[string]bool),
	}
}
func (c *LinkCheck) CheckFile(fname string, uris map[string]bool) []Issue {
	raw, err := ioutil.ReadFile(fname)
	if err != nil {
		issue := IssueError("unable to read: %s", err)
		issue.Path = fname
		return []Issue{issue}
	}
	issues := c.CheckHTML(raw, uris)
	for i, _ := range issues {
		issues[i].Path = fname
	}
	return issues
}

func (c *LinkCheck) CheckHTML(raw []byte, uris map[string]bool) []Issue {
	tr := &http.Transport{
		ResponseHeaderTimeout: 10 * time.Second,
	}
	client := &http.Client{Transport: tr}

	var issues []Issue
	doc, err := html.Parse(bytes.NewReader(raw))
	if err != nil {
		return append(issues, IssueError("unable to parse HTML: %s", err))
	}
	for _, node := range c.matcher.MatchAll(doc) {
		href := getHref(node)
		if href == "" {
			// really should never happen
			continue
		}
		link, err := url.Parse(href)
		if err != nil {
			issues = append(issues, IssueError("unable to parse url %q, %s", href, err))
			continue
		}

		// if internal reference
		if link.Scheme == "" && link.Host == "" {
			if !uris[link.Path] {
				issues = append(issues, IssueError("didn't find relative link: %q", link.Path))
			}
			continue
		}

		if link.Scheme != "" && link.Scheme != "http" && link.Scheme != "https" {
			issues = append(issues, IssueError("found bogus scheme %q", href))
			continue
		}

		// assume https
		if link.Scheme == "" {
			href = "https:" + href
		}

		if _, ok := c.external[href]; ok {
			continue
		}

		//
		log.Printf("Checking %s", href)
		res, err := client.Get(href)
		if err != nil {
			issues = append(issues, IssueWarning("external link %q failed: %s", href, err))
			c.external[href] = false
			continue
		}
		res.Body.Close()
		if res.StatusCode != 200 {
			issues = append(issues, IssueWarning("external link %q returned status %s", href, res.Status))
			c.external[href] = false
			continue
		}
		c.external[href] = true

	}
	return issues
}

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

	// get html files
	files, err := zglob.Glob(filepath.Join(root, *flagHTML))
	if err != nil {
		log.Fatalf("FAIL: %s", err)
	}

	// compute set of URIs
	uris := map[string]bool{}
	for _, f := range files {
		uri := f[len(root):]
		if *flagIndex != "" && strings.HasSuffix(uri, *flagIndex) {
			uri = uri[:len(uri)-len(*flagIndex)]
		}
		uris[uri] = true
	}
	log.Printf("Found %d uris", len(uris))

	// check
	checker := NewLinkCheck()
	counts := make(map[Severity]int)
	for _, f := range files {
		issues := checker.CheckFile(f, uris)
		for _, issue := range issues {
			counts[issue.Severity]++
			printIssue(issue)
		}
	}

	// done
	if counts[Error] > 0 {
		log.Fatalf("linkcheck failed with %d errors", counts[Error])
	}
	log.Printf("linkcheck: ok")
}
