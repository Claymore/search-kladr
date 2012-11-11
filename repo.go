package main

import (
	"database/sql"
	"fmt"
	_ "github.com/bmizerany/pq"
	"strings"
)

type RepoHandler interface {
	searchGeoObjectsByName(name string, code string) (objects []geoObject, err error)
	getRegions() (objects []geoObject, err error)
	getRegion(id string) (region *Region, err error)
	getArea(id string) (area *Area, err error)
	getSettlement(id string) (settlement *Settlement, err error)
	getCity(id string) (city *City, err error)
}

type DbRepoHandler struct {
	User     string
	Password string
	Host     string
	Port     uint64
	Name     string
}

func (h *DbRepoHandler) getConnectionString() string {
	return fmt.Sprintf("user=%s password=%s host=%s port=%d dbname=%s", h.User, h.Password, h.Host, h.Port, h.Name)
}

func (h *DbRepoHandler) searchGeoObjectsByName(name string, code string) (objects []geoObject, err error) {
	db, err := sql.Open("postgres", h.getConnectionString())
	if err != nil {
		return objects, err
	}
	defer db.Close()
	var stmt string
	switch {
	case strings.HasSuffix(code, "0000000000"):
		stmt = fmt.Sprintf("SELECT name, socr, code FROM kladr WHERE code LIKE '%s_________00' AND UPPER(name) LIKE UPPER($1)", code[:2])
	case strings.HasSuffix(code, "00000000"):
		stmt = fmt.Sprintf("SELECT name, socr, code FROM kladr WHERE code LIKE '%s______00' AND UPPER(name) LIKE UPPER($1)", code[:5])
	default:
		stmt = "SELECT name, socr, code FROM kladr WHERE code LIKE '___________00' AND UPPER(name) LIKE UPPER($1)"
	}
	objectStmt, err := db.Prepare(stmt)
	if err != nil {
		return objects, err
	}
	defer objectStmt.Close()
	rows, err := objectStmt.Query(name)
	if err != nil {
		return objects, err
	}
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

func (h *DbRepoHandler) getRegions() (objects []geoObject, err error) {
	db, err := sql.Open("postgres", h.getConnectionString())
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

func (h *DbRepoHandler) getRegion(id string) (region *Region, err error) {
	db, err := sql.Open("postgres", h.getConnectionString())
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

func (h *DbRepoHandler) getArea(id string) (area *Area, err error) {
	db, err := sql.Open("postgres", h.getConnectionString())
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

func (h *DbRepoHandler) getCity(id string) (city *City, err error) {
	db, err := sql.Open("postgres", h.getConnectionString())
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

func (h *DbRepoHandler) getSettlement(id string) (settlement *Settlement, err error) {
	db, err := sql.Open("postgres", h.getConnectionString())
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
