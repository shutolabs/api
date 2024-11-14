package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
)

func FetchImage(path string) ([]byte, error) {
	_, runningLocally := os.LookupEnv("RUNNING_LOCALLY")
	remote := "server:"
	if runningLocally {
		remote = "test:/"
	}
	cmd := exec.Command("rclone", "cat", remote+path)
	var imgBuffer bytes.Buffer
	cmd.Stdout = &imgBuffer
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to fetch image with rclone: %v", err)
	}
	return imgBuffer.Bytes(), nil
}

type RcloneFile struct {
	Path     string `json:"Path"`
	Name     string `json:"Name"`
	Size     int64  `json:"Size"`
	MimeType string `json:"MimeType"`
	ModTime  string `json:"ModTime"`
	IsDir    bool   `json:"IsDir"`
}

func ListPath(path string) ([]RcloneFile, error) {
	_, runningLocally := os.LookupEnv("RUNNING_LOCALLY")
	remote := "server:"
	if runningLocally {
		remote = "test:/"
	}
	log.Println(remote + path)
	output, err := exec.Command("rclone", "lsjson", remote+path).Output()
	if err != nil {
		return nil, fmt.Errorf("Error executing rclone lsjson: %v", err)
	}

	var files []RcloneFile
	if err := json.Unmarshal(output, &files); err != nil {
		return nil, fmt.Errorf("Error parsing JSON output: %v", err)
	}
	return files, nil
}
