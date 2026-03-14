package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

func uploadFileHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Only allow POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 2. Limit the upload size (e.g., 10 MB limit)
	r.ParseMultipartForm(10 << 20)

	// 3. Retrieve the file from form data
	file, handler, err := r.FormFile("myFile")
	if err != nil {
		fmt.Println("Error Retrieving the File:", err)
		http.Error(w, "Error retrieving file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// 4. Create an "uploads" directory if it doesn't exist
	os.MkdirAll("./uploads", os.ModePerm)

	// 5. Create a new file in the uploads directory
	dstPath := filepath.Join("uploads", handler.Filename)
	dst, err := os.Create(dstPath)
	if err != nil {
		http.Error(w, "Error saving file", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	// 6. Copy the uploaded file data to the new file
	if _, err := io.Copy(dst, file); err != nil {
		http.Error(w, "Error saving file", http.StatusInternalServerError)
		return
	}

	// 7. Return the download URL to the frontend
	fileURL := fmt.Sprintf("/files/%s", handler.Filename)
	fmt.Fprintf(w, fileURL)
}

func main() {
	// Serve the frontend HTML
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})

	// Handle the file upload requests
	http.HandleFunc("/upload", uploadFileHandler)

	// Serve the uploaded files so they can be downloaded
	http.Handle("/files/", http.StripPrefix("/files/", http.FileServer(http.Dir("./uploads"))))

	fmt.Println("Server started on http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
