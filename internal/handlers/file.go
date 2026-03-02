package handlers

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/big"
	"net/http"
	"os"
	"path"
	"strconv"
	"time"

	"github.com/YasenMakioui/0tfile/internal/config"
	"github.com/YasenMakioui/0tfile/internal/models"
)

type FileHandler struct {
	Cfg *config.Config
}

func NewFileHandler(config *config.Config) *FileHandler {
	return &FileHandler{
		Cfg: config,
	}
}

// Will return the file, but also check if the max download count
// arrived to 0 or less (not probable), or if the time expired.
// This handler does not remove the already expired files.
func (fh *FileHandler) GetFileHandler(w http.ResponseWriter, r *http.Request) {

	fileHash := r.PathValue("hash")

	log.Printf("got request for file %s", fileHash)

	metadataFile := path.Join(fh.Cfg.UploadPath, "uploads", "meta", fileHash+".json")

	if _, err := os.Stat(metadataFile); err != nil {
		log.Println("file does not exist")
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "file not found")
		return
	}

	// open metadata file and generate the filemeta struct

	jsonFile, err := os.Open(metadataFile)

	if err != nil {
		log.Printf("could not open file %s", metadataFile)
		log.Println(err)
		http.Error(w, "could not download file", http.StatusInternalServerError)
		return
	}

	defer jsonFile.Close()

	byteValue, err := io.ReadAll(jsonFile)

	if err != nil {
		log.Printf("could not read file %s", metadataFile)
		log.Println(err)
		http.Error(w, "could not download file", http.StatusInternalServerError)
		return
	}

	var fileMeta models.FileMeta

	if err := json.Unmarshal(byteValue, &fileMeta); err != nil {
		log.Println("failed unmarshaling json file")
		log.Println(err)
		http.Error(w, "could not download file", http.StatusInternalServerError)
		return
	}

	log.Println(fileMeta)

	// check if maxdownloads is more than 0

	if fileMeta.MaxDownloadCount <= 0 {
		log.Printf("can't return file with maxdownloadcount of %s", fileMeta.MaxDownloadCount)
		http.Error(w, "file expired", http.StatusGone)
		return
	}

	fileMeta.MaxDownloadCount -= 1

	// check for the date limit

	if fileMeta.ExpiresAt.Before(time.Now()) {
		log.Printf("can't return file with expired date %s", fileMeta.ExpiresAt.String())
		http.Error(w, "file expired", http.StatusGone)
		return
	}

	// dump the updated file meta to the meta file

	metadataFileContent, err := os.OpenFile(metadataFile, os.O_WRONLY|os.O_TRUNC, 0644)

	if err != nil {
		log.Printf("could not open metadata file %s for update", metadataFile)
		log.Print(err)
		http.Error(w, "could not download file", http.StatusInternalServerError)
		return
	}

	defer metadataFileContent.Close()

	jsonFileMeta, err := json.MarshalIndent(fileMeta, "", "  ")

	if err != nil {
		log.Println("failed marshaling data")
		log.Println(err)
		http.Error(w, "could not download file", http.StatusInternalServerError)
		return
	}

	_, err = metadataFileContent.Write(jsonFileMeta)

	if err != nil {
		log.Println("could not dump metadata into metadata file")
		log.Println(err)
		http.Error(w, "could not download file", http.StatusInternalServerError)
		return
	}

	// return the file

	file, err := os.Open(fileMeta.Path)

	if err != nil {
		log.Printf("could not open file %s for download", fileMeta.Path)
		http.Error(w, "could not download file", http.StatusInternalServerError)
		return
	}

	defer file.Close()

	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", fileMeta.OriginalName))
	w.Header().Set("Content-Type", "application/octet-stream")

	_, err = io.Copy(w, file)

	if err != nil {
		log.Printf("could not send file data of %s to client", fileMeta.Path)
		http.Error(w, "could not download file", http.StatusInternalServerError)
		return
	}
}

// upload file fully streaming, not loading data to memory.
// This handler will stream the data to a file with a unique name.
// It then creates a json file with metadata about the constraints of the file.
func (fh *FileHandler) PostFileHandlerStream(w http.ResponseWriter, r *http.Request) {

	r.Body = http.MaxBytesReader(w, r.Body, int64(fh.Cfg.MaxUploadSize)<<20)
	//r.Body = http.MaxBytesReader(w, r.Body, 1)

	// Max download count handling
	maxDownloadCount := r.Header.Get("Max-Download-Count")
	maxDownloadCountInt, err := strconv.Atoi(maxDownloadCount)

	if err != nil {
		maxDownloadCountInt = fh.Cfg.MaxDownloadCount
	}

	if maxDownloadCountInt > fh.Cfg.MaxDownloadCount {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Exceded max download count, limit is %d", fh.Cfg.MaxDownloadCount)
		return
	}

	// ExpiresAt handling
	// Must be days

	maxUploadDays := r.Header.Get("Max-Upload-Days")
	maxUploadDaysInt, err := strconv.Atoi(maxUploadDays)

	if err != nil {
		maxUploadDaysInt = fh.Cfg.MaxUploadedDays
	}

	if maxUploadDaysInt > fh.Cfg.MaxUploadedDays {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Exceded max stored days, limit is %d", fh.Cfg.MaxUploadedDays)
		return
	}

	mr, err := r.MultipartReader()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	for {
		part, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Println(err)
			return
		}

		if part.FormName() != "file" {
			continue
		}

		log.Println("Filename: ", part.FileName())
		log.Println("Content-Type: ", part.Header.Get("Content-Type"))

		// Filename hash
		savedFile := genFileNameHash(part.FileName())

		log.Printf("generated filename: %s", savedFile)

		// declare paths

		uploadsPath := path.Join(fh.Cfg.UploadPath, "uploads")
		metaPath := path.Join(fh.Cfg.UploadPath, "uploads", "meta")

		// file absolute path
		savePath := path.Join(fh.Cfg.UploadPath, "uploads", savedFile)

		// create file and save data onto the file
		file, err := os.Create(savePath)
		if err != nil {
			http.Error(w, "could not save file", http.StatusInternalServerError)
			log.Println(err)
			return
		}

		// copy data into file, cleanup file if something goes wrong
		defer file.Close()

		size, err := io.Copy(file, part)
		if err != nil {
			http.Error(w, "could not save file", http.StatusInternalServerError)
			log.Println(err)
			log.Printf("could not copy data into file %s", savePath)

			if err := cleanupAndAddToOrphans(savePath); err != nil {
				log.Println(err)
			}
		}

		log.Println("File Size: ", size)

		// TODO: Encrypt the file
		// https://medium.com/@mertkimyonsen/encrypt-a-file-using-go-f1fe3bc7c635

		// Define the metadata

		maxUploadDaysIntInHours := maxUploadDaysInt * 24

		fileMeta := models.FileMeta{
			Path:             path.Join(uploadsPath, savedFile),
			MaxDownloadCount: maxDownloadCountInt,
			ExpiresAt:        time.Now().Add(time.Duration(maxUploadDaysIntInHours) * time.Hour),
			OriginalName:     part.FileName(),
			DeletionToken:    "delete123",
		}

		// generate the name which is simply the filename with a json extension
		metaFileName := path.Join(metaPath, savedFile+".json")

		log.Printf("saving metadata file %s", metaFileName)

		// create metadata file
		metaFile, err := os.Create(metaFileName)
		if err != nil {
			http.Error(w, "could not save file", http.StatusInternalServerError)

			log.Println(err)
			log.Printf("could not create metadata file %s, check if uploads/meta dir is created", metaFileName)
			log.Println("removing created file...")

			// In case of removal failure, add to orphans file list.
			//os.Remove(savePath) // USE THIS ONLY TO TEST

			if err := cleanupAndAddToOrphans(savePath); err != nil {
				log.Println(err)
				return
			}

			log.Println("successfully removed file")
			return
		}

		// try to dump the metadata into the created metadata file, cleanup if failed
		jsonFileMeta, err := json.MarshalIndent(fileMeta, "", "  ")
		if err != nil {
			http.Error(w, "could not save file", http.StatusInternalServerError)
			log.Println(err)
			log.Println("failed converting struct to json")

			// remove the created file
			if err := cleanupAndAddToOrphans(savePath); err != nil {
				log.Println(err)
			}

			// remove the metadata file

			if err := removeFile(metaFileName); err != nil {
				log.Printf("could not remove file %s", metaFileName)
				log.Println(err)
				return
			}

			return
		}

		if _, err := metaFile.Write(jsonFileMeta); err != nil {
			http.Error(w, "could not save file", http.StatusInternalServerError)
			log.Println(err)
			log.Println("could not add metadata into meta file")

			// remove the created file
			if err := cleanupAndAddToOrphans(savePath); err != nil {
				log.Println(err)
			}

			return
		}

		fmt.Fprintf(w, "Successfully uploaded %s (%d bytes) with hash %s", part.FileName(), size, savedFile)
		return
	}

	http.Error(w, "File not found", http.StatusBadRequest)

}

// Will delete the given file and if unavailable, it will create a file with
// the same name as the given file into the orphans dir.
// If the path /something/somewhere/uploads/xyz is given, then xyz will be deleted.
// If xyz could not be deleted, then a file with the name xyz will be created under /something/somewhere/uploads/orphans
func cleanupAndAddToOrphans(p string) error {
	fileName := path.Base(p)
	filePath := path.Dir(p)
	orphanFile := path.Join(filePath, "orphans", fileName)

	log.Println("starting cleanup process")

	if err := os.Remove(p); err != nil {
		log.Printf("could not remove file %s", p)
		log.Printf("creating orphan file %s", orphanFile)
		if _, err := os.Create(orphanFile); err != nil {
			log.Printf("failed to create %s, remove manually", orphanFile)
			return err
		}
		return err
	}
	log.Println("completed cleanup successfully")

	return nil
}

// Simply removes a file
func removeFile(p string) error {
	if err := os.Remove(p); err != nil {
		return err
	}

	return nil
}

// Generates a hash using the given string, the actual unix time and a random 3 sequence number
func genFileNameHash(fn string) string {
	unixTime := fmt.Sprint(time.Now().Unix())
	salt := randStr(3)
	algo := sha256.New()
	algo.Write([]byte(fn + unixTime + salt))
	return hex.EncodeToString(algo.Sum(nil))
}

// Generates a random stringth of the length n
func randStr(n int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	result := make([]byte, n)
	for i := range result {
		idx, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		result[i] = charset[idx.Int64()]
	}

	return string(result)
}
