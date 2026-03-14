package main

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"
)

var db *sql.DB

func main() {
	// Connect to PostgreSQL via Coolify Environment Variable
	dbURL := os.Getenv("DATABASE_URL")
	var err error
	db, err = sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal(err)
	}

	// Auto-create the table if it doesn't exist
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS shares (
		code VARCHAR(6) PRIMARY KEY,
		content TEXT NOT NULL
	)`)
	if err != nil {
		log.Fatal(err)
	}

	// Serve the UI
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})

	// Handle saving data and generating code
	http.HandleFunc("/share", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" { return }
		content := r.FormValue("content")
		
		// Generate random 6-digit code
		rand.Seed(time.Now().UnixNano())
		code := fmt.Sprintf("%06d", rand.Intn(1000000))

		// Save to DB
		_, err := db.Exec("INSERT INTO shares (code, content) VALUES ($1, $2)", code, content)
		if err != nil {
			http.Error(w, "Database error", 500)
			return
		}
		fmt.Fprintf(w, "<html><body><h2 style='text-align:center; font-family:sans-serif; margin-top:50px;'>Success! Your code is: <b style='color:green; font-size: 30px;'>%s</b></h2></body></html>", code)
	})

	// Handle retrieving data
	http.HandleFunc("/get", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		var content string
		err := db.QueryRow("SELECT content FROM shares WHERE code = $1", code).Scan(&content)
		if err != nil {
			fmt.Fprintf(w, "<html><body><h2 style='text-align:center; font-family:sans-serif; color:red; margin-top:50px;'>Invalid or expired code.</h2></body></html>")
			return
		}
		fmt.Fprintf(w, "<html><body><h2 style='text-align:center; font-family:sans-serif; margin-top:50px;'>Message: <br><br> <span style='font-weight:normal;'>%s</span></h2></body></html>", content)
	})

	fmt.Println("Server is running on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
