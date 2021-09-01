package main

import (
	"context"
	"embed"
	"net/http"
	"os"
	"text/template"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	connString = os.Getenv("CONN_STRING")
	port       = os.Getenv("PORT")
	ctx        = context.Background()

	recipes  *mongo.Collection
	uploads  *mongo.Collection
	pictures *mongo.Collection

	//go:embed static/*
	pages embed.FS
)

type recipe struct {
	URL          string `bson:"url"`
	Category     string `bson:"category"`
	Picture      string `bson:"picture"`
	Title        string `bson:"title"`
	Language     string `bson:"language"`
	Servings     string
	Time         string
	Author       string   `bson:"author"`
	Source       string   `bson:"source"`
	Ingredients  []string `bson:"ingredients"`
	Instructions []string `bson:"instructions"`
	Notes        []string `bson:"notes"`
	Visits       int      `bson:"visits"`
}

var pageTemplate *template.Template

func main() {
	// Initiate a session with Mongo
	conn, err := mongo.Connect(ctx, options.Client().ApplyURI(connString))
	if err != nil {
		panic("mongo connection " + err.Error())
	}
	recipes = conn.Database("recipes_app").Collection("recipes")
	uploads = conn.Database("recipes_app").Collection("uploads")
	pictures = conn.Database("recipes_app").Collection("pictures")

	// Parse templates
	pageTemplate, err = template.ParseFS(pages, "static/*.*")
	if err != nil {
		panic("parse templates: " + err.Error())
	}

	// For compatibility
	http.HandleFunc("/images/", imageHandler)
	http.HandleFunc("/api", apiHandler)
	http.HandleFunc("/upload.html", uploadHandler)
	http.HandleFunc("/", staticHandler)

	http.ListenAndServe(":"+port, nil)
}
