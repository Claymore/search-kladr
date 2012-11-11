package main

import (
	"html/template"
	"net/http"
	"regexp"
	"strings"
)

type SearchResult struct {
	Regions     []geoObject
	Areas       []geoObject
	Cities      []geoObject
	Settlements []geoObject
	Query       string
}

var templates = template.Must(template.ParseFiles("templates/index.html", "templates/region.html", "templates/area.html", "templates/city.html", "templates/settlement.html", "templates/search.html"))

func renderTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
	err := templates.ExecuteTemplate(w, tmpl+".html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func indexHandler(w http.ResponseWriter, r *http.Request, repo RepoHandler) {
	objects, err := repo.getRegions()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	renderTemplate(w, "index", objects)
}

func searchHandler(w http.ResponseWriter, r *http.Request, repo RepoHandler) {
	code := ""
	if r.FormValue("local_search") == "on" {
		code = r.FormValue("code")
	}
	name := r.FormValue("geo_object_name")
	objects, err := repo.searchGeoObjectsByName(name, code)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	result := &SearchResult{Query: name}
	result.Regions = make([]geoObject, 0)
	result.Areas = make([]geoObject, 0)
	result.Cities = make([]geoObject, 0)
	result.Settlements = make([]geoObject, 0)
	for _, object := range objects {
		switch {
		case strings.HasSuffix(object.Id, "0000000000"):
			result.Regions = append(result.Regions, object)
		case strings.HasSuffix(object.Id, "00000000"):
			result.Areas = append(result.Areas, object)
		case strings.HasSuffix(object.Id, "00000"):
			result.Cities = append(result.Cities, object)
		default:
			result.Settlements = append(result.Settlements, object)
		}
	}
	renderTemplate(w, "search", result)
}

func regionHandler(w http.ResponseWriter, r *http.Request, code string, repo RepoHandler) {
	region, err := repo.getRegion(code)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	renderTemplate(w, "region", region)
}

func areaHandler(w http.ResponseWriter, r *http.Request, code string, repo RepoHandler) {
	area, err := repo.getArea(code)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	renderTemplate(w, "area", area)
}

func cityHandler(w http.ResponseWriter, r *http.Request, code string, repo RepoHandler) {
	city, err := repo.getCity(code)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	renderTemplate(w, "city", city)
}

func settlementHandler(w http.ResponseWriter, r *http.Request, code string, repo RepoHandler) {
	settlement, err := repo.getSettlement(code)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	renderTemplate(w, "settlement", settlement)
}

var urlValidator = regexp.MustCompile("^/(region|city|area|settlement)/[0-9]{13}$")

func makeGeoObjectHandler(fn func(http.ResponseWriter, *http.Request, string, RepoHandler), repo RepoHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !urlValidator.MatchString(r.URL.Path) {
			http.NotFound(w, r)
			return
		}
		slashIndex := strings.IndexRune(r.URL.Path[1:], '/')
		code := r.URL.Path[slashIndex+2:]
		fn(w, r, code, repo)
	}
}

func makeHandler(fn func(http.ResponseWriter, *http.Request, RepoHandler), repo RepoHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fn(w, r, repo)
	}
}

func runServer(address string, repo RepoHandler) {
	http.HandleFunc("/", makeHandler(indexHandler, repo))
	http.HandleFunc("/region/", makeGeoObjectHandler(regionHandler, repo))
	http.HandleFunc("/city/", makeGeoObjectHandler(cityHandler, repo))
	http.HandleFunc("/area/", makeGeoObjectHandler(areaHandler, repo))
	http.HandleFunc("/settlement/", makeGeoObjectHandler(settlementHandler, repo))
	http.HandleFunc("/search/", makeHandler(searchHandler, repo))
	http.ListenAndServe(address, nil)
}
