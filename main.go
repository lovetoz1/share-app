package main

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"os"
	"path/filepath"
)

// generateCode creates a random 6-digit code
func generateCode() string {
	const charset = "0123456789"
	b := make([]byte, 6)
	for i := range b {
		idx, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		b[i] = charset[idx.Int64()]
	}
	return string(b)
}

// uploadFileHandler handles file uploads and returns a 6-digit code
func uploadFileHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Direct call to FormFile is more efficient for single file uploads
	file, handler, err := r.FormFile("myFile")
	if err != nil {
		http.Error(w, "Error retrieving file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	code := generateCode()
	dirPath := filepath.Join("uploads", code)
	os.MkdirAll(dirPath, os.ModePerm)
	
	dstPath := filepath.Join(dirPath, handler.Filename)
	dst, err := os.Create(dstPath)
	if err != nil {
		http.Error(w, "Error saving file", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	io.Copy(dst, file)
	
	fmt.Printf("Upload complete: Code %s, File: %s\n", code, handler.Filename)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"code": code})
}

// shareTextHandler handles text/link sharing and returns a 6-digit code
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

	code := generateCode()
	dirPath := filepath.Join("uploads", code)
	os.MkdirAll(dirPath, os.ModePerm)
	
	dstPath := filepath.Join(dirPath, "snippet.txt")
	err := os.WriteFile(dstPath, []byte(textContent), 0644)
	if err != nil {
		http.Error(w, "Error saving text", http.StatusInternalServerError)
		return
	}

	fmt.Printf("Text share complete: Code %s\n", code)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"code": code})
}

// retrieveHandler finds a file associated with a code and returns its metadata
func retrieveHandler(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if len(code) != 6 {
		http.Error(w, "Invalid code", http.StatusBadRequest)
		return
	}

	dirPath := filepath.Join("uploads", code)
	files, err := os.ReadDir(dirPath)
	if err != nil || len(files) == 0 {
		http.Error(w, "Code not found", http.StatusNotFound)
		return
	}

	fileName := files[0].Name()
	fileUrl := fmt.Sprintf("/files/%s/%s", code, fileName)
	
	itemType := "file"
	if fileName == "snippet.txt" {
		itemType = "text"
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"fileName": fileName,
		"fileUrl":  fileUrl,
		"type":     itemType,
	})
}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})

	http.HandleFunc("/upload", uploadFileHandler)
	http.HandleFunc("/share-text", shareTextHandler)
	http.HandleFunc("/retrieve", retrieveHandler)
	
	http.Handle("/files/", http.StripPrefix("/files/", http.FileServer(http.Dir("./uploads"))))

	fmt.Println("Server started on http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
