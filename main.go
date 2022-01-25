package main

import (
	"embed"
	"encoding/json"
	"net/http"
	"os"
	"sort"
	"strings"
	"text/template"
)

var (
	port        = os.Getenv("PORT")
	githubToken = os.Getenv("GITHUB_TOKEN")

	//go:embed static/*
	pages embed.FS

	// recipe indexes
	titleIndex      = map[string]recipe{}
	ingredientIndex = map[string][]recipe{}
)

type recipe struct {
	URL          string
	Category     string
	Picture      string
	Title        string
	Language     string
	Servings     string
	Time         string
	Author       string
	Source       string
	Ingredients  []string
	Instructions []string
	Visits       int
	Notes        []string
}

var pageTemplate *template.Template
var recipes []recipe

func main() {
	if port == "" {
		port = "8080"
	}

	var err error
	pageTemplate, err = template.ParseFS(pages, "static/*.*")
	if err != nil {
		panic("parse templates: " + err.Error())
	}

	// parse recipes
	db, err := os.Open("db.json")
	if err != nil {
		panic("open recipes: " + err.Error())
	}
	err = json.NewDecoder(db).Decode(&recipes)
	if err != nil {
		panic("decode recipes: " + err.Error())
	}

	// build indexes
	go func() {
		for _, r := range recipes {
			titleIndex[r.Title] = r

			ingreds := map[string]struct{}{}
			for _, ingred := range r.Ingredients {
				ings := strings.Fields(ingred)
				for _, i := range ings {
					ingreds[strings.ToLower(i)] = struct{}{}
				}
			}
			for i := range ingreds {
				ingredientIndex[i] = append(ingredientIndex[i], r)
			}
		}
		sort.Slice(recipes, func(i, j int) bool { return recipes[i].Visits > recipes[j].Visits })
	}()

	// For compatibility
	http.HandleFunc("/images/", imageHandler)
	http.HandleFunc("/api", apiHandler)
	http.HandleFunc("/upload.html", uploadHandler)
	http.HandleFunc("/", staticHandler)

	http.ListenAndServe(":"+port, nil)
}
