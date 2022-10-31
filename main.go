package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/PuerkitoBio/goquery"
)

// Todo: Methods instead of functions
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

const base = "https://onepiecechapters.com"

func main() {
	// Download Latest Projects from the internet
	Projects := updateProjects()
	// For each Project, Get all chapter Links and Images
	// TODO: Concurrency
	for p := range Projects {
		Projects[p].Chapters = getProjectChapters(Projects[p].Url)
		for c := range Projects[p].Chapters {
			// Download Chapter Images
			// TODO: Concurrency
			Projects[p].Chapters[c].Images = getChapterImages(Projects[p].Chapters[c].Url)
			fmt.Println("Prj:", p, "Ch.", c)
		}
	}
	saveProjectsToDisk(Projects)
}

func updateProjects() []Project {
	var Result []Project
	// fetch latest chapters
	fetch("https://onepiecechapters.com/projects").
		Find("div .bg-card").
		Each(func(i int, s *goquery.Selection) {
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
	// yaml, _ := yaml.Marshal(Projects)
	// os.WriteFile("projects.yml", yaml, 0644)
}

func getProjectImageAndDescription(url string) (string, string) {
	var description string = ""

	// Fetch Project
	fetch(base + url).
		Find(".order-1").
		Each(func(i int, s *goquery.Selection) {
			url = s.Find("img").AttrOr("src", "")
			description = s.Find("p").Text()
		})

	return url, description
}

// sync
func getProjectChapters(url string) []Chapter {
	var Result []Chapter

	fetch(base + url).
		Find("div .block.border.border-border.bg-card.mb-3.p-3 ").
		// For Each Chapter
		Each(func(i int, s *goquery.Selection) {
			var C Chapter
			// Get Chapter Permalink
			val, exists := s.Attr("href")
			if exists {
				C.Url = val
			} else {
				log.Fatalln("Chapter Does not Exist")
			}
			// Get Chapter Text
			C.Title = s.Find("div .text-lg.font-bold").Text()
			// Get Chapter Description
			C.Description = s.Find("div .text-gray-500").Text()

			Result = append(Result, C)
		})
	return Result
}

// Return array of image sources and their respective permalinks
func getChapterImages(url string) []Image {
	var Result []Image
	fetch(base + url).
		Find("img.fixed-ratio-content").
		Each(func(i int, s *goquery.Selection) {
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

func fetch(url string) *goquery.Document {
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
	return doc

}
