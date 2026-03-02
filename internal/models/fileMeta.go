package models

import "time"

type FileMeta struct {
	Path             string    `json:"path"`
	MaxDownloadCount int       `json:"maxDownloadCount"`
	ExpiresAt        time.Time `json:"expiresAt"`
	OriginalName     string    `json:"originalName"`
	DeletionToken    string    `json:deletionToken`
}
