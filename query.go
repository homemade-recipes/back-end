package main

import (
	"sort"
	"strings"
)

func getRecipe(lang string, title string) recipe {
	r := titleIndex[title]
	if r.Language != lang {
		return recipe{}
	}
	return r
}

func getByIngredients(lang string, search []string) []recipe {
	search = cleanSearch(search)

	results := searchIngredientIndex(search)
	for i := 0; i < len(results); i++ {
		if results[i].Language != lang {
			results[i] = results[len(results)-1]
			results = results[:len(results)-1]
			i--
		}
	}
	return results
}

// each page has 20 recipes
func getMostVisited(lang string, page int) []recipe {
	var recs []recipe
	for _, r := range recipes {
		if len(recs) == 20 {
			if page == 0 {
				break
			}
			page--
			recs = []recipe{}
		}

		if r.Language == lang {
			recs = append(recs, r)
		}
	}

	return recs
}

func getByName(lang, name string, page int) []recipe {
	search := cleanSearch(strings.Fields(name))
	return searchNameIndex(search)
}

// TODO: What to do with accents?
func searchIngredientIndex(search []string) []recipe {
	results := []recipe{}
	round := []recipe{}

	for _, r := range search {
		round = append(round, ingredientIndex[r]...)
	}

	for _, r := range round {
		hasAll := true
		for _, s := range search {
			if !strings.Contains(strings.ToLower(strings.Join(r.Ingredients, " ")), s) {
				hasAll = false
				break
			}
		}

		if hasAll {
			results = append(results, r)
		}
	}

	sort.SliceStable(
		results,
		func(i, j int) bool {
			li := len(search) - len(results[i].Ingredients)
			lj := len(search) - len(results[j].Ingredients)
			return li*li < lj*lj
		},
	)

	return results
}

// TODO: What to do with accents?
func searchNameIndex(search []string) []recipe {
	matches := []struct {
		Recipe recipe
		Count  int
	}{}

	for _, r := range recipes {
		match := struct {
			Recipe recipe
			Count  int
		}{
			Recipe: r,
			Count:  0,
		}

		for _, w := range search {
			if strings.Contains(strings.ToLower(r.Title), w) {
				match.Count++
			}
		}

		if match.Count > 0 {
			matches = append(matches, match)
		}
	}

	sort.SliceStable(
		matches,
		func(i, j int) bool {
			return matches[i].Count > matches[j].Count
		},
	)

	l := len(matches)
	if l > 50 {
		l = 50
	}
	results := make([]recipe, 0, l)
	for i := 0; i < l; i++ {
		results = append(results, matches[i].Recipe)
	}

	return results
}

func cleanSearch(search []string) []string {
	for i := 0; i < len(search); i++ {
		search[i] = strings.ToLower(search[i])

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
	return search
}