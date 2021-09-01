package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"mime"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

func imageHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("received request for", r.URL.Path)

	ex := mime.TypeByExtension(r.URL.Path[strings.LastIndex(r.URL.Path, "."):])
	w.Header()["Content-Type"] = []string{ex}

	content, err := pages.ReadFile("static/" + r.URL.Path[1:])
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Cache-Control", "public, max-age=604800")
	w.Write(content)
}

func apiHandler(w http.ResponseWriter, req *http.Request) {
	log.Println("received api request")
	lang, ok := req.URL.Query()["lang"]
	if !ok {
		lang = []string{"pt"}
	}

	pg := req.URL.Query()["page"]
	if len(pg) == 0 {
		pg = []string{"0"}
	}
	page, _ := strconv.Atoi(pg[0])

	name := req.URL.Query()["name"]
	ingredients := req.URL.Query()["ingredients"]

	// The query part
	var results []recipe
	var err error
	if len(name) > 0 {
		log.Println("getting recipes by name for", lang, name[0])
		results, err = getByName(lang[0], name[0], page)
	} else if len(ingredients) > 0 {
		log.Println("getting recipes by ingredient for", lang, ingredients[0])
		results, err = getByIngredients(lang[0], ingredients)
	} else {
		results, err = getMostVisited(lang[0], page)
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(results) == 0 {
		http.Error(w, "", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(results)
}

func staticHandler(w http.ResponseWriter, req *http.Request) {
	log.Println("received request for", req.URL.Path)
	path := req.URL.Path[1:]
	if path == "" {
		path = "index.html"
	}

	ext := mime.TypeByExtension(path[strings.LastIndex(path, "."):])
	w.Header()["Content-Type"] = []string{ext}

	// Add recipes
	var data interface{}
	var err error
	lang := "pt"
	if req.Host == "en.feitaemcasa.com" {
		lang = "en"
	}

	switch path {
	case "index.html":
		by := req.URL.Query()["by"]
		if len(by) == 0 {
			by = []string{"visits"}
		}

		search, ok := req.URL.Query()["search"]
		if !ok && by[0] != "visits" {
			http.Error(w, "need a query for "+by[0], http.StatusBadRequest)
			return
		}

		pg := req.URL.Query()["page"]
		if len(pg) == 0 {
			pg = []string{"0"}
		}
		page, _ := strconv.Atoi(pg[0])

		// The query part
		var results []recipe
		switch by[0] {
		case "name":
			log.Println("getting recipes by name for", lang, search)
			results, err = getByName(lang, search[0], page)

		case "ingredients":
			log.Println("getting recipes by ingredient for", lang, search)
			results, err = getByIngredients(lang, search)

		case "visits":
			log.Println("getting most seen recipes for", lang)
			results, err = getMostVisited(lang, page)

		default:
			http.Error(w, "unknown search type", http.StatusBadRequest)
			return
		}

		if err != nil {
			log.Println("error getting recipes", by, search, err)
		}

		// Check pages
		next := page + 1
		if len(results) < 20 {
			next = 0
		}

		data = struct {
			Recipes []recipe
			Page    int
			Prev    int
			Next    int
		}{
			Recipes: results,
			Page:    page,
			Prev:    page - 1,
			Next:    next,
		}

	case "recipe.html":
		title := req.URL.Query()["title"]
		log.Println("getting recipe", title, "language", lang)
		if len(title) == 0 {
			http.Error(w, "need a title", http.StatusBadRequest)
			return
		}

		data, err = getRecipe(lang, title[0])
		if err != nil {
			// TODO: Send nice page with error
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	if lang == "pt" {
		path = strings.ReplaceAll(path, ".html", "-pt.html")
	}
	err = pageTemplate.ExecuteTemplate(w, path, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
}

func uploadHandler(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "wrong method", http.StatusBadRequest)
		return

	}
	err := req.ParseMultipartForm(100000)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validation
	lang := "pt"
	if req.Host == "en.feitaemcasa.com" {
		lang = "en"
	}
	category := req.MultipartForm.Value["category"]
	title := req.MultipartForm.Value["title"]
	author := req.MultipartForm.Value["author"]
	time := req.MultipartForm.Value["time"]
	servings := req.MultipartForm.Value["servings"]
	if len(category)+len(title)+len(author) < 3 {
		http.Error(w, "missing required fields", http.StatusBadRequest)
		return
	}

	const baseURL = "feitaemcasa.com/recipe.html?title="
	newRecipe := recipe{
		Category: category[0],
		Title:    title[0],
		Author:   author[0],
		Time:     time[0],
		Servings: servings[0],
		Source:   "Feita em casa",
		URL:      baseURL + url.QueryEscape(title[0]),
		Language: lang,
	}

	path := "upload-pt.html"
	if lang == "en" {
		newRecipe.Source = "Homemade Recipes"
		path = "upload.html"
	}

	ingredients := req.MultipartForm.Value["ingredients"]
	instructions := req.MultipartForm.Value["instructions"]
	if len(ingredients[0])+len(instructions[0]) < 2 {
		err = fmt.Errorf("ingredients and instructions must not be empty")
		pageTemplate.ExecuteTemplate(w, path, err)
		return
	}
	ingredients[0] = strings.TrimSpace(ingredients[0])
	instructions[0] = strings.TrimSpace(instructions[0])
	newRecipe.Ingredients = strings.Split(ingredients[0], "\n")
	newRecipe.Instructions = strings.Split(instructions[0], "\n")

	notes := req.MultipartForm.Value["notes"]
	if len(notes) > 0 {
		notes[0] = strings.TrimSpace(notes[0])
		newRecipe.Notes = strings.Split(notes[0], "\n")
	}

	// Get the picture
	for _, file := range req.MultipartForm.File {
		picture, _ := file[0].Open()
		content := make([]byte, file[0].Size)
		_, err = picture.Read(content)
		if err != nil {
			pageTemplate.ExecuteTemplate(w, path, err)
			return
		}

		// Encode base64
		picName := strings.ReplaceAll(newRecipe.Title, " ", "-")
		pic := map[string]interface{}{
			"name":    picName,
			"content": base64.RawStdEncoding.EncodeToString(content),
		}

		newRecipe.Picture = "pictures/" + picName
		pictures.InsertOne(ctx, pic)
	}

	// Insert new recipe
	_, err = uploads.InsertOne(ctx, newRecipe)

	// Return html
	err = pageTemplate.ExecuteTemplate(w, path, err)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
}
