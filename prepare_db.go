package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
)

func main() {
	database := "kladr"
	hostname := "localhost"
	username := "postgres"

	file, err := os.OpenFile("init.sql", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprintf(file, "DROP DATABASE %s;\n", database)
	fmt.Fprintf(file, "CREATE DATABASE %s;\n", database)
	file.Close()
	file, err = os.OpenFile("update.sql", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprintln(file, "ALTER TABLE kladr ADD CONSTRAINT pk_kladr PRIMARY KEY (code);")
	fmt.Fprintln(file, "ALTER TABLE street ADD CONSTRAINT pk_street PRIMARY KEY (code);")
	fmt.Fprintln(file, "CREATE INDEX i_kladr_name ON kladr (UPPER(name), code);")
	fmt.Fprintln(file, "CREATE INDEX i_street_name ON street (name, code);")
	file.Close()
	log.Println("Downloading...")
	res, err := http.Get("http://www.gnivc.ru/html/gnivcsoft/KLADR/BASE.7z")
	if err != nil {
		log.Fatal(err)
	}
	data, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		log.Fatal(err)
	}
	err = ioutil.WriteFile("BASE.7z", data, 0644)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Download finished.")
	log.Println("Extracting...")
	cmd := exec.Command("7z", "e", "BASE.7z", "KLADR.DBF", "STREET.DBF")
	err = cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Extraction finished.")
	log.Println("Converting...")
	cmd = exec.Command("pgdbf", "-cT", "-s", "cp866", "KLADR.DBF")
	data, err = cmd.Output()
	if err != nil {
		log.Fatal(err)
	}
	err = ioutil.WriteFile("kladr.sql", data, 0644)
	if err != nil {
		log.Fatal(err)
	}
	cmd = exec.Command("pgdbf", "-cT", "-s", "cp866", "STREET.DBF")
	data, err = cmd.Output()
	if err != nil {
		log.Fatal(err)
	}
	err = ioutil.WriteFile("street.sql", data, 0644)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Convertion finished.")
	cmd = exec.Command("psql", "-f", "init.sql", "-U", username, "-h", hostname)
	err = cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Importing table 'kladr'...")
	cmd = exec.Command("psql", "-f", "kladr.sql", "-U", username, "-h", hostname, "-d", database)
	err = cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Import finished.")
	log.Println("Importing table 'street'...")
	cmd = exec.Command("psql", "-f", "street.sql", "-U", username, "-h", hostname, "-d", database)
	err = cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Import finished.")
	log.Println("Updating database...")
	cmd = exec.Command("psql", "-f", "update.sql", "-U", username, "-h", hostname, "-d", database)
	err = cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Update finished.")
}
