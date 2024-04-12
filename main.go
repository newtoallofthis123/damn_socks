package main

import (
	"fmt"

	dbsubs "github.com/newtoallofthis123/damn_socks/db_subs"
)

func main() {
	fmt.Println("Listening on port :1234")

	server := dbsubs.NewSubWithDB()
	server.StartNewSubWithDBServer(":1234")
}
