package handlers

import (
	"log"
	"net/http"

	"github.com/YasenMakioui/0tfile/internal/config"
)

type SecretHandler struct {
	Cfg *config.Config
}

func NewSecretHandler(config *config.Config) *SecretHandler {
	return &SecretHandler{
		Cfg: config,
	}
}

func (sh *SecretHandler) GetSecretHandler(w http.ResponseWriter, r *http.Request) {
	secretHash := r.PathValue("hash")

	log.Printf("got request for secret %s", secretHash)

}

func (sh *SecretHandler) PostSecretHandler(w http.ResponseWriter, r *http.Request) {
	// get secret from the body
	// generate a hash and save it into a <hash>_secret.json file in uploads

	// generateHash

	//defer r.Body.Close()

	//secretFileName := path.Join(sh.Cfg.UploadPath, "uploads")

	//file, err := os.Create()
}
