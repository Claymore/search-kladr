package main

import (
	"database/sql"
	"fmt"
	_ "github.com/bmizerany/pq"
	"html/template"
	"net/http"
	"strings"
)

type geoObject struct {
	Name string
	Type string
	Id   string
}

type Region struct {
	geoObject
	Areas  []geoObject
	Cities []geoObject
}

type City struct {
	geoObject
	Streets []geoObject
}

type Settlement struct {
	geoObject
	Streets []geoObject
}

type Area struct {
	geoObject
	Cities      []geoObject
	Settlements []geoObject
}

type SearchResult struct {
	Regions  []geoObject
	Areas  []geoObject
	Cities      []geoObject
	Settlements []geoObject
	Query string
}

var templates = template.Must(template.ParseFiles("templates/index.html", "templates/region.html", "templates/area.html", "templates/city.html", "templates/settlement.html", "templates/search.html"))

func renderTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
	err := templates.ExecuteTemplate(w, tmpl+".html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func searchGeoObjectsByName(name string) (result *SearchResult, err error) {
	db, err := sql.Open("postgres", "user=postgres password=billing host=localhost port=5432 dbname=kladr")
	if err != nil {
		return result, err
	}
	defer db.Close()
	objectStmt, err := db.Prepare("SELECT name, socr, code FROM kladr WHERE code LIKE '___________00' AND name LIKE $1")
	if err != nil {
		return result, err
	}
	defer objectStmt.Close()
	rows, err := objectStmt.Query(name)
	if err != nil {
		return result, err
	}
	result = &SearchResult{Query: name}
	result.Regions = make([]geoObject, 0)
	result.Areas = make([]geoObject, 0)
	result.Cities = make([]geoObject, 0)
	result.Settlements = make([]geoObject, 0)
	for rows.Next() {
		var object geoObject
		err = rows.Scan(&object.Name, &object.Type, &object.Id)
		if err != nil {
			return result, err
		}
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
	return result, nil
}

func getRegions() (objects []geoObject, err error) {
	db, err := sql.Open("postgres", "user=postgres password=billing host=localhost port=5432 dbname=kladr")
	if err != nil {
		return objects, err
	}
	defer db.Close()
	objectStmt, err := db.Prepare("SELECT name, socr, code FROM kladr WHERE code LIKE '__00000000000'")
	if err != nil {
		return objects, err
	}
	defer objectStmt.Close()
	rows, err := objectStmt.Query()
	if err != nil {
		return objects, err
	}
	objects = make([]geoObject, 0)
	for rows.Next() {
		var object geoObject
		err = rows.Scan(&object.Name, &object.Type, &object.Id)
		if err != nil {
			return objects, err
		}
		objects = append(objects, object)
	}
	return objects, nil
}

func getRegion(id string) (region *Region, err error) {
	db, err := sql.Open("postgres", "user=postgres password=billing host=localhost port=5432 dbname=kladr")
	if err != nil {
		return region, err
	}
	defer db.Close()
	objectStmt, err := db.Prepare("SELECT name, socr, code FROM kladr WHERE code LIKE $1 OR code LIKE $2")
	if err != nil {
		return region, err
	}
	defer objectStmt.Close()
	areaCode := fmt.Sprintf("%s___00000000", id[0:2])
	cityCode := fmt.Sprintf("%s000___00000", id[0:2])
	rows, err := objectStmt.Query(areaCode, cityCode)
	if err != nil {
		return region, err
	}
	region = new(Region)
	region.Areas = make([]geoObject, 0)
	region.Cities = make([]geoObject, 0)
	for rows.Next() {
		var object geoObject
		err = rows.Scan(&object.Name, &object.Type, &object.Id)
		if err != nil {
			return region, err
		}
		switch {
		case object.Id == id:
			region.Name = object.Name
			region.Type = object.Type
			region.Id = object.Id
		case strings.HasSuffix(object.Id, "00000000"):
			region.Areas = append(region.Areas, object)
		default:
			region.Cities = append(region.Cities, object)
		}
	}
	return region, nil
}

func getArea(id string) (area *Area, err error) {
	db, err := sql.Open("postgres", "user=postgres password=billing host=localhost port=5432 dbname=kladr")
	if err != nil {
		return area, err
	}
	defer db.Close()
	objectStmt, err := db.Prepare("SELECT name, socr, code FROM kladr WHERE code LIKE $1")
	if err != nil {
		return area, err
	}
	defer objectStmt.Close()
	code := fmt.Sprintf("%s______00", id[0:5])
	rows, err := objectStmt.Query(code)
	if err != nil {
		return area, err
	}
	area = new(Area)
	area.Cities = make([]geoObject, 0)
	area.Settlements = make([]geoObject, 0)
	for rows.Next() {
		var object geoObject
		err = rows.Scan(&object.Name, &object.Type, &object.Id)
		if err != nil {
			return area, err
		}
		switch {
		case object.Id == id:
			area.Name = object.Name
			area.Type = object.Type
			area.Id = object.Id
		case strings.HasSuffix(object.Id, "00000"):
			area.Cities = append(area.Cities, object)
		default:
			area.Settlements = append(area.Settlements, object)
		}
	}
	return area, nil
}

func getCity(id string) (city *City, err error) {
	db, err := sql.Open("postgres", "user=postgres password=billing host=localhost port=5432 dbname=kladr")
	if err != nil {
		return city, err
	}
	defer db.Close()
	cityStmt, err := db.Prepare("SELECT name, socr, code FROM kladr WHERE code = $1")
	if err != nil {
		return city, err
	}
	defer cityStmt.Close()
	rows, err := cityStmt.Query(id)
	if err != nil {
		return city, err
	}
	city = new(City)
	city.Streets = make([]geoObject, 0)
	for rows.Next() {
		err = rows.Scan(&city.Name, &city.Type, &city.Id)
		if err != nil {
			return city, err
		}
	}
	objectStmt, err := db.Prepare("SELECT name, socr, code FROM street WHERE code LIKE $1")
	if err != nil {
		return city, err
	}
	defer objectStmt.Close()
	code := fmt.Sprintf("%s____00", id[0:11])
	rows, err = objectStmt.Query(code)
	if err != nil {
		return city, err
	}
	for rows.Next() {
		var object geoObject
		err = rows.Scan(&object.Name, &object.Type, &object.Id)
		if err != nil {
			return city, err
		}
		city.Streets = append(city.Streets, object)
	}
	return city, nil
}

func getSettlement(id string) (settlement *Settlement, err error) {
	db, err := sql.Open("postgres", "user=postgres password=billing host=localhost port=5432 dbname=kladr")
	if err != nil {
		return settlement, err
	}
	defer db.Close()
	settlementStmt, err := db.Prepare("SELECT name, socr, code FROM kladr WHERE code = $1")
	if err != nil {
		return settlement, err
	}
	defer settlementStmt.Close()
	rows, err := settlementStmt.Query(id)
	if err != nil {
		return settlement, err
	}
	settlement = new(Settlement)
	settlement.Streets = make([]geoObject, 0)
	for rows.Next() {
		err = rows.Scan(&settlement.Name, &settlement.Type, &settlement.Id)
		if err != nil {
			return settlement, err
		}
	}
	objectStmt, err := db.Prepare("SELECT name, socr, code FROM street WHERE code LIKE $1")
	if err != nil {
		return settlement, err
	}
	defer objectStmt.Close()
	code := fmt.Sprintf("%s____00", id[0:11])
	rows, err = objectStmt.Query(code)
	if err != nil {
		return settlement, err
	}

	for rows.Next() {
		var object geoObject
		err = rows.Scan(&object.Name, &object.Type, &object.Id)
		if err != nil {
			return settlement, err
		}
		settlement.Streets = append(settlement.Streets, object)
	}
	return settlement, nil
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	objects, err := getRegions()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	renderTemplate(w, "index", objects)
}

func searchHandler(w http.ResponseWriter, r *http.Request) {
	result, err := searchGeoObjectsByName(r.FormValue("geo_object_name"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	renderTemplate(w, "search", result)
}

func regionHandler(w http.ResponseWriter, r *http.Request, code string) {
	region, err := getRegion(code)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	renderTemplate(w, "region", region)
}

func areaHandler(w http.ResponseWriter, r *http.Request, code string) {
	area, err := getArea(code)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	renderTemplate(w, "area", area)
}

func cityHandler(w http.ResponseWriter, r *http.Request, code string) {
	city, err := getCity(code)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	renderTemplate(w, "city", city)
}

func settlementHandler(w http.ResponseWriter, r *http.Request, code string) {
	settlement, err := getSettlement(code)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	renderTemplate(w, "settlement", settlement)
}

func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slashIndex := strings.IndexRune(r.URL.Path[1:], '/')
		code := r.URL.Path[slashIndex+2:]
		fn(w, r, code)
	}
}

func main() {
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/region/", makeHandler(regionHandler))
	http.HandleFunc("/city/", makeHandler(cityHandler))
	http.HandleFunc("/area/", makeHandler(areaHandler))
	http.HandleFunc("/settlement/", makeHandler(settlementHandler))
	http.HandleFunc("/search/", searchHandler)
	http.ListenAndServe(":8080", nil)
}
