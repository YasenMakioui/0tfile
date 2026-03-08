package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path"

	"github.com/YasenMakioui/0tfile/internal/config"
	"github.com/YasenMakioui/0tfile/internal/handlers"
)

func ensureDir(path string) error {
	log.Printf("ensuring dir %s", path)
	_, err := os.Stat(path)
	// dir does not exist, create it
	if err != nil {
		log.Printf("%s does not exist", path)
		log.Printf("creating dir %s", path)
		if err := os.MkdirAll(path, 0755); err != nil {
			log.Print(err)
			log.Printf("could not create dir %s", path)
			return err
		}
	}

	return nil
}

func main() {

	router := http.NewServeMux()

	log.Println("===starting 0tfile server===")

	// get configuration

	log.Println("Loading configuration values")

	cfg := config.Load()

	jsonCfg, _ := json.Marshal(cfg)

	log.Println(string(jsonCfg))

	// instantiate handler

	fileHandler := handlers.NewFileHandler(cfg)
	secretHandler := handlers.NewSecretHandler(cfg)

	router.HandleFunc("GET /f/{hash}", fileHandler.GetFileHandler)
	router.HandleFunc("POST /f", fileHandler.PostFileHandlerStream)
	router.HandleFunc("POST /f/{hash}", fileHandler.DeleteFileHandler)
	router.HandleFunc("GET /s/{hash}", secretHandler.GetSecretHandler)

	uploadsDir := path.Join(cfg.UploadPath, "uploads")
	metaDir := path.Join(uploadsDir, "meta")
	orphansDir := path.Join(uploadsDir, "orphans")

	// check if necessary dirs are created, if not, create them

	log.Printf("Checking for %s dir", uploadsDir)

	if err := ensureDir(uploadsDir); err != nil {
		panic("failed startup due to failure on dir creation")
	}

	log.Printf("Checking for %s dir", metaDir)

	if err := ensureDir(metaDir); err != nil {
		panic("failed startup due to failure on dir creation")
	}

	log.Printf("Checking for %s dir", orphansDir)

	if err := ensureDir(orphansDir); err != nil {
		panic("failed startup due to failure on dir creation")
	}

	if err := http.ListenAndServe(":3000", router); err != nil {
		panic("could not start server")
	}
}
