package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

func main() {
	r := mux.NewRouter()

	// Routes
	r.HandleFunc("/get-all-kapal", authorize(getAllKapal)).Methods("GET")
	r.HandleFunc("/get-kapal-by-id", authorize(getKapalByID)).Methods("GET")
	r.HandleFunc("/create-kapal", authorize(createKapal)).Methods("POST")
	r.HandleFunc("/update-kapal", authorize(updateKapal)).Methods("PUT")
	r.HandleFunc("/delete-kapal/{id}", authorize(deleteKapal)).Methods("DELETE")
	r.HandleFunc("/get-kapals-by-pemilik-kapal-id", authorize(getKapalsByPemilikKapalId)).Methods("GET")

	// CORS configuration
	headers := handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization"})
	methods := handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"})
	origins := handlers.AllowedOrigins([]string{"*"}) // You can specify the allowed origins instead of "*"

	log.Println("Server running on port 9010")
	log.Fatal(http.ListenAndServe(":9010", handlers.CORS(headers, methods, origins)(r)))
}

// Middleware
func authorize(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Authorization logic here (if any)
		next.ServeHTTP(w, r)
	}
}

// Handlers
func getAllKapal(w http.ResponseWriter, r *http.Request) {
	db, err := sql.Open("mysql", "root:@tcp(127.0.0.1:3306)/service_kapal")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	rows, err := db.Query("SELECT id, nama, deskripsi, pemilik_kapal_id FROM kapals")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var kapals []map[string]string
	for rows.Next() {
		var id, nama, deskripsi, pemilik_kapal_id string
		if err := rows.Scan(&id, &nama, &deskripsi, &pemilik_kapal_id); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		kapal := map[string]string{
			"id":               id,
			"nama":             nama,
			"deskripsi":        deskripsi,
			"pemilik_kapal_id": pemilik_kapal_id,
		}
		kapals = append(kapals, kapal)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(kapals)
}

func getKapalByID(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "Parameter ID is required", http.StatusBadRequest)
		return
	}

	db, err := sql.Open("mysql", "root:@tcp(127.0.0.1:3306)/service_kapal")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	row := db.QueryRow("SELECT id, nama, deskripsi, pemilik_kapal_id FROM kapals WHERE id = ?", id)

	var kapalID, nama, deskripsi, pemilikKapalID string
	err = row.Scan(&kapalID, &nama, &deskripsi, &pemilikKapalID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Kapal not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	kapal := map[string]string{
		"id":               kapalID,
		"nama":             nama,
		"deskripsi":        deskripsi,
		"pemilik_kapal_id": pemilikKapalID,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(kapal)
}

func createKapal(w http.ResponseWriter, r *http.Request) {
	var kapal map[string]string
	err := json.NewDecoder(r.Body).Decode(&kapal)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	db, err := sql.Open("mysql", "root:@tcp(127.0.0.1:3306)/service_kapal")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	result, err := db.Exec("INSERT INTO kapals (nama, deskripsi, pemilik_kapal_id) VALUES (?, ?, ?)", kapal["nama"], kapal["deskripsi"], kapal["pemilik_kapal_id"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	id, _ := result.LastInsertId()
	w.Header().Set("Content-Type", "application/json")
	response := map[string]int64{"id": id}
	json.NewEncoder(w).Encode(response)
}

func updateKapal(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ID        string `json:"id"`
		Nama      string `json:"nama"`
		Deskripsi string `json:"deskripsi"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	db, err := sql.Open("mysql", "root:@tcp(127.0.0.1:3306)/service_kapal")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	query := "UPDATE kapals SET nama = ?, deskripsi = ? WHERE id = ?"
	stmt, err := db.Prepare(query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(input.Nama, input.Deskripsi, input.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Kapal updated successfully"))
}

func getKapalsByPemilikKapalId(w http.ResponseWriter, r *http.Request) {
	pemilikKapalID := r.URL.Query().Get("pemilik_kapal_id")
	if pemilikKapalID == "" {
		http.Error(w, "Parameter pemilik_kapal_id is required", http.StatusBadRequest)
		return
	}

	db, err := sql.Open("mysql", "root:@tcp(127.0.0.1:3306)/service_kapal")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	rows, err := db.Query("SELECT id, nama, deskripsi, pemilik_kapal_id FROM kapals WHERE pemilik_kapal_id = ?", pemilikKapalID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var kapals []map[string]string
	for rows.Next() {
		var kapalID, nama, deskripsi, pemilikKapalID string
		if err := rows.Scan(&kapalID, &nama, &deskripsi, &pemilikKapalID); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		kapal := map[string]string{
			"id":               kapalID,
			"nama":             nama,
			"deskripsi":        deskripsi,
			"pemilik_kapal_id": pemilikKapalID,
		}
		kapals = append(kapals, kapal)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(kapals)
}

func deleteKapal(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	db, err := sql.Open("mysql", "root:@tcp(127.0.0.1:3306)/service_kapal")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	result, err := db.Exec("DELETE FROM kapals WHERE id = ?", id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if rowsAffected == 0 {
		http.Error(w, "Kapal not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Kapal deleted successfully"))
}
