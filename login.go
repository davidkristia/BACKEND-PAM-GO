package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
)

func checkCredentials(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		// Handle preflight request
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Buka koneksi ke database
	db, err := sql.Open("mysql", "root:@tcp(127.0.0.1:3306)/service_user")
	if err != nil {
		// Jika terjadi kesalahan saat membuka koneksi ke database
		http.Error(w, "Service sedang bermasalah", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Parse form data
	err = r.ParseForm()
	if err != nil {
		http.Error(w, "Failed to parse form data", http.StatusInternalServerError)
		return
	}

	// Ambil data username, password dari form
	email := r.Form.Get("email")
	password := r.Form.Get("password")

	// Periksa apakah email dan password cocok dalam database
	var id int      // ID pengguna
	var role string // Peran pengguna
	var username string
	row := db.QueryRow("SELECT id, role, username FROM users WHERE email = ? AND password = ?", email, password)
	err = row.Scan(&id, &role, &username)
	if err != nil {
		// Jika terjadi kesalahan saat mengeksekusi query atau tidak ada baris yang cocok
		response := map[string]interface{}{
			"status":  "failed",
			"message": "Email atau password salah",
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		json.NewEncoder(w).Encode(response)
		return // Menghentikan eksekusi fungsi setelah memberikan respons
	}

	// Jika cocok, kirim status kode 200 (OK) bersama dengan ID pengguna dan role dalam bentuk JSON
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	response := map[string]interface{}{
		"status":   "success",
		"id":       id,
		"role":     role,
		"username": username,
	}
	json.NewEncoder(w).Encode(response)
}

func main() {
	var mux = http.NewServeMux()

	// login
	mux.HandleFunc("/check-credentials", checkCredentials)

	fmt.Println("User server running on port: 8004")

	// Jalankan server HTTP pada port 8004
	http.ListenAndServe(":8004", mux)
}
