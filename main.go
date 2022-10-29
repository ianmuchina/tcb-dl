package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
)

type Project struct {
	Title       string
	Description string
	Url         string
}

type Chapter struct {
	Title       string
	Description string
	Images      []string
}

var Projects = []Project{
	{Title: "Ace Novel - Manga Adaptation", Url: "/mangas/1/ace-novel-manga-adaptation"},
	{Title: "Attack on Titan", Url: "/mangas/8/attack-on-titan"},
	{Title: "Black Clover", Url: "/mangas/3/black-clover"},
	{Title: "Bleach", Url: "/mangas/2/bleach"},
	{Title: "Build King", Url: "/mangas/9/build-king"},
	{Title: "ChainSaw Man", Url: "/mangas/13/chainsaw-man"},
	{Title: "Haikyuu!! (New Special!)", Url: "/mangas/11/haikyu-special"},
	{Title: "Jujutsu Kaisen", Url: "/mangas/4/jujutsu-kaisen"},
	{Title: "My Hero Academia", Url: "/mangas/6/my-hero-academia"},
	{Title: "One Piece", Url: "/mangas/5/one-piece"},
	{Title: "One Piece - Nami vs Kalifa by Boichi", Url: "/mangas/12/one-piece-nami-vs-kalifa-by-boichi"},
	{Title: "One-Punch Man", Url: "/mangas/10/one-punch-man"},
}

func main() {
	p, _ := json.MarshalIndent(Projects, "", "\t")

	projects_file, err := os.Create("projects.json")

	if err != nil {
		log.Fatal(err)
	}

	defer projects_file.Close()

	w := bufio.NewWriter(projects_file)
	fmt.Fprintf(w, "%s", p)
	w.Flush()
}
