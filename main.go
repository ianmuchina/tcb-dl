package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

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
	Index       string  `json:",omitempty"`
	Title       string  `json:",omitempty"`
	Url         string  `json:",omitempty"`
	Description string  `json:",omitempty"`
	Images      []Image `json:",omitempty"`
}

type Image struct {
	Src string
	Alt string
}

type CubariManga struct {
	Title       string                   `json:"title"`
	Description string                   `json:"description"`
	Artist      string                   `json:"artist"`
	Author      string                   `json:"author"`
	Cover       string                   `json:"cover"`
	Chapters    map[string]CubariChapter `json:"chapters"`
}

type CubariChapter struct {
	Title        string              `json:"title"`
	Volume       string              `json:"volume"`
	Groups       map[string][]string `json:"groups"`
	Last_updated string
}

const base = "https://onepiecechapters.com"

// HTTP Client
var httpClient = &http.Client{
	Timeout: time.Second * 10,
}

var ChapterMap map[string]Chapter = make(map[string]Chapter, 4000)

func main() {
	P := loadLocalData()
	updateChapterMap(P)

	P = fetchNewChapters(P)
	P = CleanData(P)

	genCubariData(P)
	saveProjectsToDisk(P)
}

func updateChapterMap(P []Project) {
	for _, p := range P {
		for _, c := range p.Chapters {
			ChapterMap[c.Url] = c
		}
	}
}

func fetchNewChapters(P []Project) []Project {
	Projects := updateProjects()

	for i := 0; i < len(Projects); i++ {
		Projects[i].Chapters = getProjectChapters(Projects[i].Url)
		for c := 0; c < len(Projects[i].Chapters); c++ {

			ch := Projects[i].Chapters[c]
			// If new chapter
			if _, ok := ChapterMap[ch.Url]; !ok {
				// fetch images
				Projects[i].Chapters[c].Images = getChapterImages(Projects[i].Chapters[c].Url)
				fmt.Println("New", ch.Title)
			} else {
				// Already seen
				Projects[i].Chapters[c] = ChapterMap[ch.Url]
			}
		}
	}
	return Projects
}

func CleanData(Projects []Project) []Project {
	re := regexp.MustCompile("[0-9.]+")

	for i, p := range Projects {
		for j, c := range p.Chapters {

			x, err := strconv.ParseFloat(re.FindAllString(c.Title, -1)[0], 32)
			if err != nil {
				log.Fatal(err)
			}
			Projects[i].Chapters[j].Index = fmt.Sprintf("%g", x)
		}
	}

	saveProjectsToDisk(Projects)
	return Projects
}

func loadLocalData() []Project {
	data, err := os.ReadFile("./projects.json")
	if err != nil {
		log.Fatal(err)
	}

	tmp := []Project{}
	err = json.Unmarshal(data, &tmp)
	if err != nil {
		log.Fatal(err)
	}
	return tmp
}

func genCubariData(P []Project) {
	for _, m := range ExportCubariData(P) {
		os.Mkdir("./data", 0755)
		filename := "./data/" + m.Title + ".json"
		filename = strings.ReplaceAll(filename, " ", "_")
		fmt.Println(filename)
		json, err := json.MarshalIndent(m, "", "\t")
		if err != nil {
			log.Fatal(err)
		}

		err = os.WriteFile(filename, json, 0644)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func ExportCubariData(Projects []Project) []CubariManga {
	var Result []CubariManga
	for _, P := range Projects {
		Manga := CubariManga{
			Title:       P.Title,
			Description: P.Description,
			Artist:      "Unknown",
			Author:      "Unknown",
			Cover:       P.Image,
			Chapters:    make(map[string]CubariChapter, len(P.Chapters)),
		}

		for _, C := range P.Chapters {
			// Sort Chapter Images
			var imgs []string
			for _, I := range C.Images {
				imgs = append(imgs, I.Src)
			}
			G := map[string][]string{
				"tcbscans": imgs,
			}

			Chp := CubariChapter{
				Title:  C.Title,
				Groups: G,
			}

			Manga.Chapters[C.Index] = Chp
		}

		Result = append(Result, Manga)
	}

	return Result
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
	res, err := httpClient.Get(url)
	if err != nil {
		log.Fatal(err, "Failed to get Latest Chapters")
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		body, _ := io.ReadAll(res.Body)
		res.Body.Close()
		log.Fatalf("status code error: %d %s\n%s", res.StatusCode, res.Status, body)
	}
	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}
	return doc

}
