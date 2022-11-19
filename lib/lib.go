package lib

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type ChapterDict map[float64]Chapter

type Project struct {
	Title       string              `json:"Title,omitempty"`
	Description string              `json:"Description,omitempty"`
	Url         string              `json:"Url,omitempty"`
	Image       string              `json:"Image,omitempty"`
	Chapters    map[float64]Chapter `json:"Chapters,omitempty"`
	Id          int                 `json:"Id,omitempty"`
}

type Chapter struct {
	Index       float64 `json:"Index,omitempty"`
	Title       string  `json:"Title,omitempty"`
	Url         string  `json:"Url,omitempty"`
	Description string  `json:"Description,omitempty"`
	Images      []Image `json:"Images,omitempty"`
}

type Image struct {
	Src string
	Alt string
}

const base = "https://onepiecechapters.com"

var commit_msg string = ""

// HTTP Client
var httpClient = &http.Client{
	Timeout: time.Second * 10,
}

var ProjectsArr []Project

var ExistingChapters map[string]Chapter = make(map[string]Chapter, 4000)

// TODO: Map to Actual Structs
var ProjectsMap = map[int]string{
	1:  "/mangas/1/ace-novel-manga-adaptation",
	2:  "/mangas/2/bleach",
	3:  "/mangas/3/black-clover",
	4:  "/mangas/4/jujutsu-kaisen",
	5:  "/mangas/5/one-piece",
	6:  "/mangas/6/my-hero-academia",
	8:  "/mangas/8/attack-on-titan",
	9:  "/mangas/9/build-king",
	10: "/mangas/10/one-punch-man",
	11: "/mangas/11/haikyu-special",
	12: "/mangas/12/one-piece-nami-vs-kalifa-by-boichi",
	13: "/mangas/13/chainsaw-man",
	15: "/mangas/15/hunter-x-hunter",
}

func SyncAll() {
	//P := LoadLocalData()
	//UpdateChapterMap(P)

	P := FetchAllData()

	SaveCubariData(P)
	SaveProjectsToDisk(P)
	os.WriteFile("commit_msg", []byte(commit_msg), 0644)
}

func SyncNew() {
	P := LoadLocalData()
	UpdateChapterMap(P)

	P = FetchNewChapters(P)

	SaveCubariData(P)
	os.WriteFile("commit_msg", []byte(commit_msg), 0644)
}

func UpdateChapterMap(P []Project) {
	for _, p := range P {
		for _, c := range p.Chapters {
			ExistingChapters[c.Url] = c
		}
	}
}

func FetchNewChapters(P []Project) []Project {
	Projects := FetchProjects()

	for i := 0; i < len(Projects); i++ {
		Projects[i].Chapters = FetchProjectChapters(Projects[i].Url)

		for ch_id, ch := range Projects[i].Chapters {

			// If new chapter
			if _, ok := ExistingChapters[ch.Url]; !ok {
				// fetch images
				ch.Images = FetchChapterImages(ch.Url)
				Projects[i].Chapters[ch_id] = ch
				fmt.Println("New", ch.Title)
				commit_msg = commit_msg + ch.Title + "\n"
			} else {
				// Already seen
				Projects[i].Chapters[ch_id] = ExistingChapters[ch.Url]
			}
		}
	}

	return Projects
}

// Go through every chapter and fetch every image
func FetchAllData() []Project {
	Projects := FetchProjects()
	commit_msg = "update data"
	for i := 0; i < len(Projects); i++ {
		// Chapters
		Projects[i].Chapters = FetchProjectChapters(Projects[i].Url)
		for ch_id, ch := range Projects[i].Chapters {
			ch.Images = FetchChapterImages(ch.Url)
			Projects[i].Chapters[ch_id] = ch
			fmt.Println("New", ch.Title)
			//commit_msg = commit_msg + ch.Title + "\n"
		}
	}

	return Projects
}

var re = regexp.MustCompile("[0-9.]+")

func getChapterIdFromTitle(Title string) float64 {
	id, err := strconv.ParseFloat(re.FindAllString(Title, -1)[0], 32)
	if err != nil {
		log.Fatal(err)
	}
	return id
}

func LoadLocalData() []Project {
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

func FetchProjects() []Project {
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
				log.Fatalln("Error Getting Project Url")
			}

			// Chapter Image
			image := s.Find("img")
			src, imageExists := image.Attr("src")
			if imageExists {
				P.Image = src
			} else {
				log.Fatalln("Error Getting project image")
			}
			// Get Project Image & Description
			// async
			P.Image, P.Description = GetProjectImageAndDescription(P.Url)
			// Append Project
			Result = append(Result, P)
		})

	// wait
	return Result
}

func SaveProjectsToDisk(Projects []Project) {
	json, _ := json.MarshalIndent(Projects, "", "\t")
	os.WriteFile("projects.json", json, 0644)
	// yaml, _ := yaml.Marshal(Projects)
	// os.WriteFile("projects.yml", yaml, 0644)
}

func GetProjectImageAndDescription(url string) (string, string) {
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
func FetchProjectChapters(url string) ChapterDict {
	var Result map[float64]Chapter = make(map[float64]Chapter)

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
			// Get Chapter Uext
			C.Title = s.Find("div .text-lg.font-bold").Text()
			// Get Chapter Description
			C.Description = s.Find("div .text-gray-500").Text()

			C.Index = getChapterIdFromTitle(C.Title)
			Result[C.Index] = C
		})
	return Result
}

// Return array of image sources and their respective permalinks
func FetchChapterImages(url string) []Image {
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
		log.Fatal(err, "Failed to Get Latest Chapters")
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

func FindChapter(ProjectId int, ChapterId float64) Chapter {
	Projects := LoadLocalData()

	var Project Project
	var Chapter Chapter

	// Select Project
	// TODO: Make Faster
	for _, p := range Projects {
		if p.Url == ProjectsMap[ProjectId] {
			Project = p
		}
	}

	// Select Chapter
	// TODO: Make Faster
	for _, ch := range Project.Chapters {
		if ChapterId == ch.Index {
			Chapter = ch
		}
	}

	return Chapter
}

// Returns Latest Chapter
func GetLatest(ProjectId int) Chapter {
	var Data []Project = LoadLocalData()
	var Project Project

	// Select Project
	for _, p := range Data {
		if p.Url == ProjectsMap[ProjectId] {
			Project = p
		}
	}

	var max float64 = 0
	var maxCh Chapter
	// Get largest Chapter
	for _, ch := range Project.Chapters {
		if ch.Index > max {
			max = ch.Index
			maxCh = ch
		}
	}

	return maxCh
}
