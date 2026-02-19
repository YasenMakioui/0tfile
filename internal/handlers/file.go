package handlers

import (
	"fmt"
	"net/http"
)

func GetFileHandler(w http.ResponseWriter, r *http.Request) {
	// handle errors inside using http.Error(w, "message", http.StatusInternalServerError)
	fmt.Fprintf(w, "Get file")
}

func PostFileHandler(w http.ResponseWriter, r *http.Request) {
	// The way of saving files here is quite simple

	// A file comes and we read the size
	// If its passed the limit, we return an error
	// We collect the filemetadata and mimetype
	// If its not a compressed file, we compress it
	// We encrypt the file using a generated secret
	// We generate a hash using the filename and something else
	// We save it into the uploads dir with the name hash.json
	// If all went ok, we return a the download url and password url

	fmt.Fprintf(w, "Post file")
}
