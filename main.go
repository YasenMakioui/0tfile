package main

import (
	"log"
	"net/http"

	"github.com/YasenMakioui/0tfile/internal/handlers"
)

func main() {

	router := http.NewServeMux()

	router.HandleFunc("GET /file", handlers.GetFileHandler)
	router.HandleFunc("POST /file", handlers.PostFileHandler)

	log.Println("===starting 0tfile server===")

	if err := http.ListenAndServe(":3000", router); err != nil {
		panic("could not start server")
	}
}
