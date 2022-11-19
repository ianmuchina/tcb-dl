package lib

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
)

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

// Save data to json readable by cubari.moe
func SaveCubariData(P []Project) {
	for _, m := range GenCubariData(P) {
		os.Mkdir("./data", 0755)
		filename := "./data/" + m.Title + ".json"
		filename = strings.ReplaceAll(filename, " ", "_")
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

// Convert data to format acceptable by cubari.moe
func GenCubariData(Data []Project) []CubariManga {
	var Projects []CubariManga

	for _, p := range Data {
		Project := CubariManga{
			Title:       p.Title,
			Description: p.Description,
			Artist:      "Unknown",
			Author:      "Unknown",
			Cover:       p.Image,
			Chapters:    make(map[string]CubariChapter, len(p.Chapters)),
		}

		for _, C := range p.Chapters {
			var imgs []string
			for _, I := range C.Images {
				imgs = append(imgs, I.Src)
			}

			chapter := CubariChapter{
				Title:  C.Title,
				Groups: map[string][]string{"tcbscans": imgs},
			}

			Project.Chapters[fmt.Sprint(C.Index)] = chapter
		}

		Projects = append(Projects, Project)
	}

	return Projects
}
