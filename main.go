package main

import (
	"flag"
)

func main() {
	var dbHandler = &DbRepoHandler{User: "postgres", Password: "billing", Host: "localhost", Port: 5432, Name: "kladr"}
	var address string
	flag.StringVar(&dbHandler.User, "dbuser", "postgres", "database user")
	flag.StringVar(&dbHandler.Password, "dbpassword", "billing", "database password")
	flag.StringVar(&dbHandler.Host, "dbhost", "localhost", "database host")
	flag.StringVar(&dbHandler.Name, "dbname", "kladr", "database name")
	flag.Uint64Var(&dbHandler.Port, "dbport", 5432, "database port")
	flag.StringVar(&address, "address", ":8080", "listen and serve on this address")
	flag.Parse()

	runServer(address, dbHandler)
}
