package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
)

type ShortLink struct {
	UID string `json:"uid"`
	Url string `json:"url"`
}

// Initialize Database
func initializedb() error {
	db, err := sql.Open("sqlite3", "database.db")
	if err != nil {
		fmt.Println(err)
	}

	createTable := `
	CREATE TABLE IF NOT EXISTS links (
		uid CHAR PRIMARY KEY UNIQUE NOT NULL,
		url TEXT NOT NULL
	);`

	_, err = db.Exec(createTable)
	if err != nil {
		fmt.Println(err)
	}

	return nil
}

// Create uid that not in database
func createUid(n int) string {
	rand.Seed(time.Now().UnixNano())

	alphabet := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	var uid string
	count := 1

	db, err := sql.Open("sqlite3", "database.db")
	if err != nil {
		fmt.Println(err)
		return ""
	}
	defer db.Close()

	for count != 0 {
		uid = ""
		for i := 0; i < n; i++ {
			letter := alphabet[rand.Intn(len(alphabet))]
			uid += string(letter)
		}

		query := "SELECT COUNT(*) FROM links WHERE uid = ?"
		err := db.QueryRow(query, uid).Scan(&count)
		if err != nil {
			fmt.Println(err)
			return ""
		}
	}

	return uid
}

// assign url to uid
func CreateShortUrl(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		fmt.Println(err)
		return
	}
	url := r.Form.Get("url")

	uid := createUid(6)

	db, err := sql.Open("sqlite3", "database.db")
	if err != nil {
		fmt.Println(err)
	}
	defer db.Close()

	query := "INSERT INTO links (url, uid) VALUES (?, ?)"
	_, err = db.Exec(query, url, uid)
	if err != nil {
		fmt.Println(err)
	}

	shortlink := ShortLink{
		UID: uid,
		Url: url,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(shortlink)
}

// Get url by uid
func GetUrl(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	uid := vars["uid"]

	db, err := sql.Open("sqlite3", "database.db")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer db.Close()

	query := "SELECT url FROM links WHERE uid =?"
	rows, err := db.Query(query, uid)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer rows.Close()

	var url string
	for rows.Next() {
		err := rows.Scan(&url)
		if err != nil {
			fmt.Println(err)
			return
		}
	}

	shortlink := ShortLink{
		UID: uid,
		Url: url,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(shortlink)

}

func RedirectUrl(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uid := vars["uid"]

	db, err := sql.Open("sqlite3", "database.db")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer db.Close()
	query := "SELECT url FROM links WHERE uid =?"
	rows, err := db.Query(query, uid)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer rows.Close()
	var url string
	for rows.Next() {
		err := rows.Scan(&url)
		if err != nil {
			fmt.Println(err)
			return
		}
	}

	http.Redirect(w, r, url, http.StatusSeeOther)
}

func main() {
	err := initializedb()
	if err != nil {
		fmt.Println(err)
		return
	}

	router := mux.NewRouter()

	router.HandleFunc("/api/links/{uid}", GetUrl).Methods("GET")
	router.HandleFunc("/api/create", CreateShortUrl).Methods("POST")
	router.HandleFunc("/{uid}", RedirectUrl).Methods("GET")
	http.ListenAndServe(":8080", router)

}
