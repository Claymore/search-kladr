package main

import (
	"database/sql"
	"flag"
	"fmt"
	_ "github.com/bmizerany/pq"
	"log"
)

func main() {
	var objName, streetName string
	flag.StringVar(&objName, "obj", "", "search for objects with this name")
	flag.StringVar(&streetName, "street", "", "search for streets with this name")
	flag.Parse()

	db, err := sql.Open("postgres", "user=postgres password=billing host=localhost port=5432 dbname=kladr")
	if err != nil {
		log.Fatalf("Error: %s\n", err)
	}
	defer db.Close()
	objectStmt, err := db.Prepare("SELECT name, socr, substring(code from 1 for 11) FROM kladr WHERE code like '%00' AND name LIKE $1")
	if err != nil {
		log.Fatalf("Error: %s\n", err)
	}
	defer objectStmt.Close()
	streetStmt, err := db.Prepare("SELECT name, socr FROM street WHERE name LIKE $1 AND code like $2 ORDER BY name")
	if err != nil {
		log.Fatalf("Error: %s\n", err)
	}
	defer streetStmt.Close()
	rows, err := objectStmt.Query(objName)
	if err != nil {
		log.Fatalf("Error: %s\n", err)
	}
	for rows.Next() {
		var name, short, code string
		err = rows.Scan(&name, &short, &code)
		if err != nil {
			log.Fatalf("Error: %s\n", err)
		}
		fmt.Printf("%s %s\n", name, short)
		if streetName != "" {
			rows2, err := streetStmt.Query(streetName, fmt.Sprintf("%s____00", code))
			if err != nil {
				log.Fatalf("Error: %s\n", err)
			}
			for rows2.Next() {
				var (
					name2, short2 string
				)
				err = rows2.Scan(&name2, &short2)
				if err != nil {
					log.Fatalf("Error: %s\n", err)
				}
				fmt.Printf("|-- %s %s\n", name2, short2)
			}
		}
	}
}
