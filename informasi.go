package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
)

type Info struct {
	ID         int    `json:"id"`
	Penjelasan string `json:"penjelasan"`
}

func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			return
		}

		next.ServeHTTP(w, r)
	})
}

func HandleGetAllInfo(w http.ResponseWriter, r *http.Request) {
	db, err := sql.Open("sqlite3", "service_informasi.db")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	rows, err := db.Query("SELECT id, penjelasan FROM info")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var infos []Info
	for rows.Next() {
		var i Info
		if err := rows.Scan(&i.ID, &i.Penjelasan); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		infos = append(infos, i)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(infos)
}

func HandleSaveInfo(w http.ResponseWriter, r *http.Request) {
	var info Info
	err := json.NewDecoder(r.Body).Decode(&info)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	db, err := sql.Open("sqlite3", "service_informasi.db")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	_, err = db.Exec("INSERT INTO info (penjelasan) VALUES (?)", info.Penjelasan)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func HandleUpdateInfo(w http.ResponseWriter, r *http.Request) {
	var info Info
	err := json.NewDecoder(r.Body).Decode(&info)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	db, err := sql.Open("sqlite3", "service_informasi.db")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	_, err = db.Exec("UPDATE info SET penjelasan = ? WHERE id = ?", info.Penjelasan, info.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func main() {
	db, err := sql.Open("sqlite3", "service_informasi.db")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS info (id INTEGER PRIMARY KEY AUTOINCREMENT, penjelasan TEXT)")
	if err != nil {
		panic(err)
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/get-all-info", HandleGetAllInfo)
	mux.HandleFunc("/save-info", HandleSaveInfo)
	mux.HandleFunc("/update-info", HandleUpdateInfo)
	// Tambahkan handler untuk endpoint lain jika diperlukan

	fmt.Println("Server running on port : 9090")
	http.ListenAndServe(":9090", enableCORS(mux))
}
