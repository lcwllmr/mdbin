package main

import (
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

func CmdServe(port int, htmlDir string) {
	log.Printf("Starting server...\n")
	htmlDir = initHTMLDir(htmlDir)

	http.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		uploadHandler(htmlDir, w, r)
	})
	http.Handle("/view/", http.StripPrefix("/view/", http.FileServer(http.Dir(htmlDir))))

	log.Println("Binding to port", port)
	if err := http.ListenAndServe(":"+strconv.Itoa(port), nil); err != nil {
		log.Fatalf("Failed to launch server: %v", err)
	}
	log.Println("Server started successfully")
}

func initHTMLDir(htmlDir string) string {
	if htmlDir == "" {
		log.Println("No HTML store directory specified. Using a temporary directory.")
		tmpDir, err := os.MkdirTemp("", "*-mdbin")
		if err != nil {
			log.Fatalf("Failed to create temporary directory: %v", err)
		}
		htmlDir = tmpDir
		log.Printf("Using temporary directory for HTML store: %s\n", htmlDir)
	} else {
		absolutePath, err := filepath.Abs(htmlDir)
		if err != nil {
			log.Fatalf("Failed to get absolute path of store directory: %v", err)
		}
		htmlDir = absolutePath

		if _, err := os.Stat(htmlDir); os.IsNotExist(err) {
			log.Printf("Directory %s does not exist. Creating it...\n", htmlDir)
			if err := os.MkdirAll(htmlDir, 0755); err != nil {
				log.Fatalf("Failed to create HTML store at %s: %v", htmlDir, err)
			}
		} else {
			fi, err := os.Stat(htmlDir)
			switch {
			case err != nil:
				log.Fatalf("Failed to check HTML store directory %s: %v", htmlDir, err)
			case fi.IsDir():
				if err := os.Chmod(htmlDir, 0755); err != nil {
					log.Fatalf("Failed to set permissions for HTML store directory %s: %v", htmlDir, err)
				}
			default:
				log.Fatalf("%s is not a directory", htmlDir)
			}
		}
		log.Printf("HTML store is at: %s\n", htmlDir)
	}

	return htmlDir
}

func uploadHandler(htmlDir string, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	mdContent, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}

	if len(mdContent) == 0 {
		http.Error(w, "Empty content", http.StatusBadRequest)
		return
	}

	id, htmlDoc := ConvertMarkdownToHtml(mdContent, NewMdOpts(
		EnableMathjax(),
		StandaloneDocument()))

	filePath := filepath.Join(htmlDir, id+".html")
	if err := os.WriteFile(filePath, htmlDoc, 0644); err != nil {
		log.Printf("Failed to save file %v, error: %v", id, err)
		http.Error(w, "Upload failed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(id))

	log.Printf("Successfully received new file with ID %v", id)
}
