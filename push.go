package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
)

func CmdPush(server string, file string) {
	fileContent, err := os.ReadFile(file)
	if err != nil {
		fmt.Printf("Failed to read file: %v\n", err)
		os.Exit(1)
	}

	uploadURL := fmt.Sprintf("%s/upload", server)
	req, err := http.NewRequest(http.MethodPut, uploadURL, bytes.NewReader(fileContent))
	if err != nil {
		fmt.Printf("Failed to create request at %s: %v\n", uploadURL, err)
		os.Exit(1)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Failed to send request to %s: %v\n", uploadURL, err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Failed to read response: %v\n", err)
		os.Exit(1)
	}

	view_url := fmt.Sprintf("%s/view/%s.html", server, responseBody)
	fmt.Printf("Upload successful. View at %s\n", view_url)
}
