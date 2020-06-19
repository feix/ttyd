package main

import (
	"net/http"

	"github.com/feix/ttyd/server"
)

func main() {
	router := server.NewRouter()
	_ = http.ListenAndServe(":7681", router)
}
