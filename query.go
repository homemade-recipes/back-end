package main

import (
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func getRecipe(lang string, title string) (r recipe, err error) {
	return r, recipes.FindOneAndUpdate(
		ctx,
		bson.M{"title": title, "language": lang},
		bson.M{"$inc": bson.M{"visits": 1}},
	).Decode(&r)
}

func getByIngredients(lang string, search []string) (rs []recipe, err error) {
	// Split words
	for i := 0; i < len(search); i++ {
		ws := strings.Split(search[i], ",")
		search[i] = strings.TrimSpace(ws[0])
		if len(ws) > 1 {
			search = append(search, ws[1:]...)
		}

		ws = strings.Split(search[i], " e ")
		search[i] = strings.TrimSpace(ws[0])
		if len(ws) > 1 {
			search = append(search, ws[1:]...)
		}

		ws = strings.Split(search[i], " and ")
		search[i] = strings.TrimSpace(ws[0])
		if len(ws) > 1 {
			search = append(search, ws[1:]...)
		}
	}

	pipe := buildQuery(lang, search)

	c := options.Collation{Strength: 1, Locale: lang}
	opts := options.AggregateOptions{Collation: &c}
	cur, err := recipes.Aggregate(ctx, pipe, &opts)
	if err != nil {
		return
	}

	return rs, cur.All(ctx, &rs)
}

func getMostVisited(lang string, page int) (rs []recipe, err error) {
	// Query options
	lim := int64(20)
	skip := int64(20*page)
	opts := options.FindOptions{
		Limit: &lim,
		Sort: bson.M{"visits": -1},
		Skip: &skip,
	}

	// Create the mongoDB filter from the map
	cur, err := recipes.Find(ctx, bson.M{"language": lang}, &opts)
	if err != nil {
		return
	}

	return rs, cur.All(ctx, &rs)
}

func getByName(lang, name string, page int) (rs []recipe, err error) {
	// Query options
	lim := int64(20)
	s:= int64(20*page)
	p := bson.M{"score": bson.M{"$meta": "textScore"}}
	os := options.FindOptions{Limit: &lim, Sort: p, Projection: p, Skip: &s}

	// Create the mongoDB filter from the map
	cur, err := recipes.Find(
		ctx,
		bson.M{"language": lang, "$text": bson.M{"$search": name}},
		&os,
	)
	if err != nil {
		return
	}

	return rs, cur.All(ctx, &rs)
}

// TODO: What to do with accents?
func buildQuery(language string, fields []string) []interface{} {
	// Build $and array input
	and := bson.A{}
	for _, f := range fields {
		if f != "e" && f != "and" {
			and = append(and,
				bson.M{
					"ingredients": bson.M{
						"$regex":   strings.TrimSpace(f),
						"$options": "i",
					},
				},
			)
		}
	}

	// Match part
	match := bson.M{"$match": bson.M{"language": language, "$and": and}}

	// Projection part
	proj := bson.M{
		"$project": bson.M{
			"_id":          0,
			"category":     1,
			"instructions": 1,
			"author":       1,
			"notes":        1,
			"picture":      1,
			"source":       1,
			"ingredients":  1,
			"title":        1,
			"score": bson.M{
				"$abs": bson.M{
					"$subtract": bson.A{
						bson.M{
							"$size": "$ingredients",
						},
						len(fields),
					},
				},
			},
		},
	}

	// Sort part
	sorting := bson.M{
		"$sort": bson.M{"score": 1},
	}

	// Limit part
	limit := bson.M{"$limit": 50}

	// Agregate needs an array
	return bson.A{match, proj, sorting, limit}
}
