package lib

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"sync"
	"time"

	"go.uber.org/ratelimit"
	"gopkg.in/yaml.v3"

	"github.com/PuerkitoBio/goquery"
)

type ChapterDict map[float64]Chapter

type Project struct {
	Title       string
	Description string
	Url         string
	Image       string
	Chapters    map[float64]Chapter
	Id          int
}

type Chapter struct {
	Index       float64
	Title       string
	Url         string
	Description string
	Images      []Image
	Date        string
}

type Image struct {
	Src string
	Alt string
}

const base = "https://onepiecechapters.com"

var commit_msg string = ""

// per second
var rl = ratelimit.New(25)

var httpClient = &http.Client{
	Timeout: time.Second * 10,
}

var ProjectsArr []Project

var newChapters = []Chapter{}

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

func SyncNew() {
	var P = []Project{}

	P = loadLocalData()
	updateExistingChapters(P)
	fmt.Println(len(existingChapters), "Chapters")

	P = fetchNewData(P)
	updateExistingChapters(P)
	fmt.Println(len(existingChapters))

	saveProjectsToDisk(P)
	SaveCubariData(P)
	genCommitMsg()
}

var existingChapters = make(map[string]Chapter, 4000)

func updateExistingChapters(P []Project) {

	for _, p := range P {
		for _, c := range p.Chapters {
			existingChapters[c.Url] = c
		}
	}

}

// Faster than fetchAllData
func fetchNewData(P []Project) []Project {
	Projects := P

	var wg sync.WaitGroup
	wg.Add(len(Projects))

	for projectId, project := range Projects {
		go func(projectId int, project Project) {
			defer wg.Done()
			for chapterId, chapter := range fetchProjectChapters(project.Url) {
				if _, ok := existingChapters[chapter.Url]; !ok {
					chapter.Images = fetchChapterImages(chapter.Url)
					chapter.Date = getDate(chapter)

					Projects[projectId].Chapters[chapterId] = chapter
					newChapters = append(newChapters, chapter)

					fmt.Println("new", chapter.Title)
				} else {
					Projects[projectId].Chapters[chapterId] = existingChapters[chapter.Url]
				}
			}
		}(projectId, project)
	}
	wg.Wait()

	return Projects
}

func SyncAll() {
	Projects := fetchAllData()

	// SaveCubariData(Projects)
	saveProjectsToDisk(Projects)

	// err := os.WriteFile("commit_msg", []byte(commit_msg), 0644)
	// if err != nil {
	// 	log.Fatal(err)
	// }
}

// Go through every chapter and fetch every image
func fetchAllData() []Project {
	Projects := fetchProjects()
	// commit_msg = "update data"
	var wg sync.WaitGroup
	wg.Add(len(Projects))

	for i := range Projects {
		go func(i int) {
			Projects[i].Chapters = fetchProjectChapters(Projects[i].Url)

			var wg2 sync.WaitGroup
			for ch_id, ch := range Projects[i].Chapters {
				wg2.Add(len(Projects[i].Chapters))
				go func(ch_id float64, ch Chapter) {
					ch.Images = fetchChapterImages(ch.Url)
					Projects[i].Chapters[ch_id] = ch
					fmt.Println(ch.Title)
					wg2.Done()
				}(ch_id, ch)
				wg2.Wait()
			}
			wg.Done()
		}(i)
	}
	wg.Wait()

	return Projects
}

func genCommitMsg() {
	var suffix string = ""
	if len(newChapters) > 1 {
		suffix = "s"
	}

	var commit_msg = fmt.Sprintln("Added", len(newChapters), "New Chapter", suffix)
	for _, ch := range newChapters {
		commit_msg = fmt.Sprintln(commit_msg, ch.Title)
	}
	fmt.Println(commit_msg)
	err := os.WriteFile("commit_msg", []byte(commit_msg), 0644)
	checkErr(err)
}

var re = regexp.MustCompile("[0-9.]+")

func loadLocalData() []Project {
	if yamlIn {

		Result := []Project{}
		f, err := os.ReadFile("./projects.yml")
		checkErr(err)
		err = yaml.Unmarshal(f, &Result)
		checkErr(err)

		return Result
	} else {

		Result := []Project{}
		file, err := os.ReadFile("projects.bin")
		checkErr(err)

		buf := bytes.NewBuffer(file)

		decoder := gob.NewDecoder(buf)
		err = decoder.Decode(&Result)
		checkErr(err)

		return Result
	}
}

var yamlOut = true
var yamlIn = true

func fileExists(filename string) bool {
	_, err := os.Stat("/path/to/whatever")
	return !errors.Is(err, os.ErrNotExist)
}

func saveProjectsToDisk(Projects []Project) {
	if yamlOut {

		yaml, err := yaml.Marshal(Projects)
		checkErr(err)
		os.WriteFile("projects.yml", yaml, 0644)

	} else {

		var b bytes.Buffer
		e := gob.NewEncoder(&b)
		if err := e.Encode(Projects); err != nil {
			log.Fatal(err)
		}

		os.WriteFile("projects.bin", b.Bytes(), 0644)

	}
}

func fetchProjects() []Project {
	var Result []Project
	var wg sync.WaitGroup
	fetchDoc("https://onepiecechapters.com/projects").
		Find("div .bg-card").
		Each(
			func(i int, s *goquery.Selection) {
				wg.Add(1)
				go func(i int, s *goquery.Selection) {

					P := Project{}
					// Title
					P.Title = s.Find(".font-bold").Text()

					// Url
					url, attrExists := s.Find("a").Attr("href")
					if attrExists {
						P.Url = url
					} else {
						log.Fatalln("Error Getting Project Url")
					}

					image := s.Find("img")
					src, imageExists := image.Attr("src")
					if imageExists {
						P.Image = src
					} else {
						log.Fatalln("Error Getting project image")
					}

					P.Image, P.Description = getProjectImageAndDescription(P.Url)

					P.Chapters = make(ChapterDict)

					Result = append(Result, P)
					fmt.Println(P.Title)
					wg.Done()
				}(i, s)
			})
	wg.Wait()
	fmt.Println("updated project list")
	return Result
}

func getProjectImageAndDescription(url string) (string, string) {
	var description string = ""
	// Fetch Project
	fetchDoc(base + url).
		Find(".order-1").
		Each(func(i int, s *goquery.Selection) {
			url = s.Find("img").AttrOr("src", "")
			description = s.Find("p").Text()
		})
	return url, description
}

// sync
func fetchProjectChapters(url string) ChapterDict {

	var Result map[float64]Chapter = make(map[float64]Chapter, 3000)

	fetchDoc(base + url).
		Find("div .block.border.border-border.bg-card.mb-3.p-3 ").
		// For Each Chapter
		Each(func(i int, s *goquery.Selection) {
			var C Chapter
			val, exists := s.Attr("href")
			if !exists {
				log.Fatalln("Chapter Does not Exist")
			}
			C.Url = val

			C.Title = s.Find("div .text-lg.font-bold").Text()
			C.Description = s.Find("div .text-gray-500").Text()
			C.Index = ParseChapterId(C.Title)

			Result[C.Index] = C
		})

	return Result
}

// Return array of image sources and their respective permalinks
func fetchChapterImages(url string) []Image {
	var Result []Image
	fetchDoc(base + url).
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

func fetchDoc(url string) *goquery.Document {
	rl.Take()
	res, err := httpClient.Get(url)
	fmt.Println(url)

	if err != nil {
		log.Fatal(err, "Failed to Get Latest Chapters")
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		body, err := io.ReadAll(res.Body)
		if err != nil {
			log.Fatal(err)
		}
		res.Body.Close()
		log.Fatalf("status code error: %d %s\n%s", res.StatusCode, res.Status, body)
	}
	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(res.Body)

	checkErr(err)
	return doc

}

func FindChapter(ProjectId int, ChapterId float64) Chapter {
	Projects := loadLocalData()

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
func GetLatestChapter(ProjectId int) Chapter {
	var Data []Project = loadLocalData()
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

func ParseChapterId(Title string) float64 {
	id, err := strconv.ParseFloat(re.FindAllString(Title, -1)[0], 32)
	checkErr(err)
	return id
}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func getDate(C Chapter) string {
	rl.Take()
	res, err := httpClient.Get(C.Images[0].Src)
	checkErr(err)

	defer res.Body.Close()
	if res.StatusCode != 200 {
		body, err := io.ReadAll(res.Body)
		checkErr(err)
		res.Body.Close()
		log.Fatalf("status code error: %d %s\n%s", res.StatusCode, res.Status, body)
	}

	return res.Header.Get("x-bz-upload-timestamp")
}

func hasInternet() bool {
	return false
}
