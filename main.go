package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// Handles file uploads
func uploadFileHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	r.ParseMultipartForm(10 << 20) // 10MB limit for now
	file, handler, err := r.FormFile("myFile")
	if err != nil {
		http.Error(w, "Error retrieving file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	os.MkdirAll("./uploads", os.ModePerm)
	dstPath := filepath.Join("uploads", handler.Filename)
	
	dst, err := os.Create(dstPath)
	if err != nil {
		http.Error(w, "Error saving file", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	io.Copy(dst, file)
	
	fmt.Fprintf(w, "/files/%s", handler.Filename)
}

// Handles text/link sharing
func shareTextHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	r.ParseForm()
	textContent := r.FormValue("textContent")

	if textContent == "" {
		http.Error(w, "Text is empty", http.StatusBadRequest)
		return
	}

	os.MkdirAll("./uploads", os.ModePerm)
	
	// Create a unique filename for the text snippet
	fileName := fmt.Sprintf("snippet_%d.txt", time.Now().Unix())
	dstPath := filepath.Join("uploads", fileName)

	// Save the text to the file
	err := os.WriteFile(dstPath, []byte(textContent), 0644)
	if err != nil {
		http.Error(w, "Error saving text", http.StatusInternalServerError)
		return
	}

	// Return the link to the text file
	fmt.Fprintf(w, "/files/%s", fileName)
}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})

	http.HandleFunc("/upload", uploadFileHandler)
	http.HandleFunc("/share-text", shareTextHandler)
	
	http.Handle("/files/", http.StripPrefix("/files/", http.FileServer(http.Dir("./uploads"))))

	fmt.Println("Server started on http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
