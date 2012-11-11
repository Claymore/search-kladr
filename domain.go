package main

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
