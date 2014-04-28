package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
)

type Payload struct {
	Cells []*Cell
	Mazes []*Maze
	Links map[string]*Link
}

type Link struct {
	Href string
	Type string
}

type Cell struct {
	Element
}

type Maze struct {
	Element
}

type Element struct {
	Exit_Link  string
	Id         string
	Name       string
	ReadableId int
	Abandon    string
	Type       string
	Links      map[string]string
}

var directions map[string]string
var visited map[string]bool
var linkRegex *regexp.Regexp = regexp.MustCompile("{(.*)}")

func main() {
	url := os.Args[1]

	directions = make(map[string]string)
	visited = make(map[string]bool)

	process(url)
}

func process(url string) {
	if _, ok := visited[url]; !ok {
		result := getJson(url)
		fillMap(result)
		visitLinks(result)
	}
}

func visitLinks(payload *Payload) {
	var id string
	if len(payload.Cells) > 0 {
		id = payload.Cells[0].Id
	}
	if len(payload.Mazes) > 0 {
		id = payload.Mazes[0].Id
	}

	for _, link := range payload.Links {
		key := linkRegex.FindAllStringSubmatch(link.Href, 1)
		if key != nil {
			dir, ok := directions[key[0][1]+"."+id]
			if ok {
				url := linkRegex.ReplaceAllString(link.Href, dir)
				process(url)
			}
		}
	}
}

func fillMap(payload *Payload) {
	if payload.Cells != nil {
		for _, cell := range payload.Cells {
			if cell.Exit_Link != "" {
				log.Println(cell.Exit_Link)
				os.Exit(0)
			}
			for key, value := range cell.Links {
				directions["cells."+key+"."+cell.Id] = value
			}
		}
	}

	if payload.Mazes != nil {
		for _, maze := range payload.Mazes {
			for key, value := range maze.Links {
				directions["mazes."+key+"."+maze.Id] = value
			}
		}
	}
}

func getJson(url string) *Payload {
	var p Payload
	log.Printf("opening", url)
	response, err := http.Get(url)
	if err != nil {
		log.Fatal("couldn't open the url", url)
	}

	if response.StatusCode != 200 {
		log.Printf("%#v\n", directions)
		log.Fatal("we got a wrong response", response.Body, response.Status, url)
	} else {
		visited[url] = true
		content, err := ioutil.ReadAll(response.Body)
		if err != nil {
			log.Fatal("couldn't read the body", url)
		} else {
			err := json.Unmarshal(content, &p)
			if err != nil {
				log.Fatal("couldn't parse the json", string(content), url, err)
			} else {
				log.Println(string(content))
			}
		}
	}

	return &p
}
