package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/PuerkitoBio/goquery"
	yaml "gopkg.in/yaml.v3"
)

type Project struct {
	Title       string    `json:"Title"`
	Description string    `json:"Description,omitempty"`
	Url         string    `json:"Url"`
	Image       string    `json:"Image,omitempty"`
	Chapters    []Chapter `json:"Chapters,omitempty"`
}

type Chapter struct {
	Title       string  `json:",omitempty"`
	Url         string  `json:",omitempty"`
	Description string  `json:",omitempty"`
	Images      []Image `json:",omitempty"`
}

type Image struct {
	Src string
	Alt string
}

func main() {
	Projects := updateProjects()
	for i := range Projects {
		Projects[i].Chapters = getProjectChapters(Projects[i].Url)
		for j := range Projects[i].Chapters {
			Projects[i].Chapters[j].Images = getChapterImages(Projects[i].Chapters[j].Url)
			fmt.Println("Ch.", i, j)
		}
	}
	saveProjectsToDisk(Projects)
}

func updateProjects() []Project {
	var Result []Project
	// fetch latest chapters
	res, err := http.Get("https://onepiecechapters.com/projects")
	if err != nil {
		log.Fatal(err, "Failed to get Latest Chapters")
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	// find all projects using css selectors
	doc.Find("div .bg-card").Each(func(i int, s *goquery.Selection) {
		var P Project
		// Title
		P.Title = s.Find(".font-bold").Text()
		// Url
		url, attrExists := s.Find("a").Attr("href")
		if attrExists {
			P.Url = url
		} else {
			log.Fatalln("Error getting Project Url")
		}

		// Chapter Image
		image := s.Find("img")
		src, imageExists := image.Attr("src")
		if imageExists {
			P.Image = src
		} else {
			log.Fatalln("Error getting project image")
		}
		// Get Project Image & Description
		// async
		P.Image, P.Description = getProjectImageAndDescription(P.Url)
		// Append Project
		Result = append(Result, P)
	})

	// wait
	return Result
}

func saveProjectsToDisk(Projects []Project) {
	json, _ := json.MarshalIndent(Projects, "", "\t")
	os.WriteFile("projects.json", json, 0644)
	yaml, _ := yaml.Marshal(Projects)
	os.WriteFile("projects.yml", yaml, 0644)
}

func getProjectImageAndDescription(url string) (string, string) {
	url = "https://onepiecechapters.com" + url
	description := ""

	// Fetch Project
	res, err := http.Get(url)
	if err != nil {
		log.Fatal(err, "Failed to get Latest Chapters")
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	// find all projects using css selectors
	doc.Find(".order-1").Each(func(i int, s *goquery.Selection) {
		url = s.Find("img").AttrOr("src", "")
		description = s.Find("p").Text()
	})

	return url, description
}

// sync
func getProjectChapters(url string) []Chapter {
	Result := []Chapter{}
	url = "https://onepiecechapters.com" + url

	// Fetch Project
	res, err := http.Get(url)
	if err != nil {
		log.Fatal(err, "Failed to get Latest Chapters")
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	// Find all Chapters using css selectors
	doc.Find("div .block.border.border-border.bg-card.mb-3.p-3 ").Each(func(i int, s *goquery.Selection) {
		var C Chapter
		C.Url = s.AttrOr("href", "")
		C.Title = s.Find("div .text-lg.font-bold").Text()
		C.Description = s.Find("div .text-gray-500").Text()
		Result = append(Result, C)
	})
	return Result
}

func getChapterImages(url string) []Image {
	Result := []Image{}
	url = "https://onepiecechapters.com" + url

	res, err := http.Get(url)
	if err != nil {
		log.Fatal(err, "Failed to get Latest Chapters")
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}
	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	doc.Find("img.fixed-ratio-content").Each(func(i int, s *goquery.Selection) {
		var I Image
		src, imageExists := s.Attr("src")
		if imageExists {
			I.Src = src
		} else {
			log.Fatal("image has no src attribute")
		}

		alt, altExists := s.Attr("alt")
		// Alt is Optional
		if altExists {
			I.Alt = alt
		}
		Result = append(Result, I)
	})
	return Result
}
