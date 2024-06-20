package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
)

// Middleware untuk mengizinkan CORS
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

// Middleware untuk otorisasi (saat ini tidak ada logika khusus)
func authorize(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: Tambahkan logika autentikasi di sini jika diperlukan
		next.ServeHTTP(w, r)
	}
}

func addNahkoda(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Metode permintaan tidak valid", http.StatusMethodNotAllowed)
		return
	}

	var nahkoda struct {
		Nama         string `json:"nama"`
		NomorHP      string `json:"nomor_hp"`
		JenisKelamin string `json:"jenis_kelamin"`
	}

	if err := json.NewDecoder(r.Body).Decode(&nahkoda); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	db, err := sql.Open("mysql", "root:@tcp(127.0.0.1:3306)/service_nahkoda")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	_, err = db.Exec("INSERT INTO nahkodas (nama, nomor_hp, jenis_kelamin) VALUES (?, ?, ?)",
		nahkoda.Nama, nahkoda.NomorHP, nahkoda.JenisKelamin)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func editNahkoda(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Metode permintaan tidak valid", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.URL.Path[len("/edit-nahkoda/"):]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "ID nahkoda tidak valid", http.StatusBadRequest)
		return
	}

	var nahkoda struct {
		Nama         string `json:"nama"`
		NomorHP      string `json:"nomor_hp"`
		JenisKelamin string `json:"jenis_kelamin"`
	}

	if err := json.NewDecoder(r.Body).Decode(&nahkoda); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	db, err := sql.Open("mysql", "root:@tcp(127.0.0.1:3306)/service_nahkoda")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	_, err = db.Exec("UPDATE nahkodas SET nama=?, nomor_hp=?, jenis_kelamin=? WHERE id=?",
		nahkoda.Nama, nahkoda.NomorHP, nahkoda.JenisKelamin, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func deleteNahkoda(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Metode permintaan tidak valid", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.URL.Path[len("/delete-nahkoda/"):]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "ID nahkoda tidak valid", http.StatusBadRequest)
		return
	}

	db, err := sql.Open("mysql", "root:@tcp(127.0.0.1:3306)/service_nahkoda")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	_, err = db.Exec("DELETE FROM nahkodas WHERE id=?", id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func getAllNahkoda(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Metode permintaan tidak valid", http.StatusMethodNotAllowed)
		return
	}

	db, err := sql.Open("mysql", "root:@tcp(127.0.0.1:3306)/service_nahkoda")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	rows, err := db.Query("SELECT id, nama, nomor_hp, jenis_kelamin FROM nahkodas")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var nahkodas []map[string]interface{}
	for rows.Next() {
		var id int
		var nama, nomor_hp, jenis_kelamin string
		if err := rows.Scan(&id, &nama, &nomor_hp, &jenis_kelamin); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		nahkoda := map[string]interface{}{
			"id":            id,
			"nama":          nama,
			"nomor_hp":      nomor_hp,
			"jenis_kelamin": jenis_kelamin,
		}
		nahkodas = append(nahkodas, nahkoda)
	}

	if err := json.NewEncoder(w).Encode(nahkodas); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func main() {
	http.HandleFunc("/add-nahkoda", authorize(addNahkoda))
	http.HandleFunc("/edit-nahkoda/", authorize(editNahkoda))
	http.HandleFunc("/delete-nahkoda/", authorize(deleteNahkoda))
	http.HandleFunc("/get-all-nahkoda", authorize(getAllNahkoda))

	http.ListenAndServe(":9008", enableCORS(http.DefaultServeMux))
}
