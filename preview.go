package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func CmdPreview(port int, markdownFile string) {
	var clients sync.Map

	// Determine the parent directory of the markdown file
	watchDir := filepath.Dir(markdownFile)
	filename := filepath.Base(markdownFile)

	log.Println("Monitoring file", filename, "in directory", watchDir)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if websocket.IsWebSocketUpgrade(r) {
			// Handle WebSocket connection
			conn, err := upgrader.Upgrade(w, r, nil)
			if err != nil {
				log.Println("WebSocket upgrade error:", err)
				return
			}
			log.Println("WebSocket connection opened")
			defer conn.Close()

			clients.Store(conn, true)

			for {
				_, _, err := conn.ReadMessage()
				if err != nil {
					log.Println("WebSocket read error:", err)
					break
				}
			}

			// Remove client from active connections on disconnection
			clients.Delete(conn)
			log.Println("WebSocket connection closed")
			return
		}

		// Serve the markdown file as HTML
		fileContent, err := os.ReadFile(markdownFile)
		if err != nil {
			http.Error(w, "File not found", http.StatusNotFound)
			return
		}

		_, htmlContent := ConvertMarkdownToHtml(fileContent, NewMdOpts(
			EnableMathjax(),
			EnablePreviewMode(),
			StandaloneDocument()))
		w.Header().Set("Content-Type", "text/html")
		w.Write(htmlContent)
	})

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	if err := watcher.Add(watchDir); err != nil {
		log.Fatal("Failed to add directory to watcher:", err)
	}

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Has(fsnotify.Write) && filepath.Base(event.Name) == filename {
					log.Println("Write event detected for", event.Name)
					fileContent, err := os.ReadFile(event.Name)
					if err != nil {
						log.Println("Error reading updated file:", err)
						continue
					}

					_, htmlContent := ConvertMarkdownToHtml(fileContent, NewMdOpts())
					log.Println("Converted updated content to HTML for file:", event.Name)

					clients.Range(func(key, value any) bool {
						conn := key.(*websocket.Conn)
						log.Println("Sending updated content to client for file:", event.Name)
						err := conn.WriteMessage(websocket.TextMessage, htmlContent)
						if err != nil {
							log.Println("Error sending message to client:", err)
							clients.Delete(conn)
						}
						return true
					})
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("Watcher error:", err)
			}
		}
	}()

	log.Printf("Starting server on http://localhost:%d", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}
