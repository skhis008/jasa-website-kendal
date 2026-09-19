// main.go
//
// Backend sederhana untuk Kendal Jasa Website, dibuat dengan Go standard library saja
// (net/http) supaya tidak perlu mengunduh dependency eksternal untuk menjalankannya.
// Tugas utamanya: menerima data dari form kontak di frontend Nuxt dan menyimpannya
// (di sini disimpan ke file log sederhana; di produksi bisa diganti ke database/email).
package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

// ContactRequest merepresentasikan body JSON yang dikirim dari form kontak.
type ContactRequest struct {
	Name    string `json:"name"`
	Email   string `json:"email"`
	Message string `json:"message"`
}

// apiResponse adalah bentuk standar respons JSON dari API ini.
type apiResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/health", handleHealth)
	mux.HandleFunc("/api/contact", handleContact)

	// corsMiddleware membungkus semua route supaya frontend (yang berjalan di origin/port berbeda,
	// misalnya http://localhost:3000) diizinkan mengakses API ini.
	handler := corsMiddleware(mux)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server backend berjalan di http://localhost:%s\n", port)
	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatalf("Server gagal berjalan: %v", err)
	}
}

// corsMiddleware menambahkan header CORS dasar.
// Untuk produksi, ganti "*" dengan domain frontend Anda yang sebenarnya.
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Ambil origin domain dari request header
		origin := r.Header.Get("Origin")

		// Daftar domain yang diizinkan (ganti dengan domain Vercel asli Anda nantinya)
		allowedOrigins := map[string]bool{
			"http://localhost:3000":          true, // lokal Nuxt
			"https://nama-app-anda.vercel.app": true, // domain Vercel produksi
		}

		// Jika origin terdaftar, izinkan. Jika tidak ada di daftar, fallback ke domain utama
		if allowedOrigins[origin] {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "https://nama-app-anda.vercel.app")
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		// Preflight request OPTIONS
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// handleHealth adalah endpoint sederhana untuk memastikan server hidup, berguna saat deployment.
func handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, apiResponse{Success: true, Message: "Server berjalan normal"})
}

// handleContact menerima data form kontak, memvalidasinya, lalu menyimpannya ke log.
func handleContact(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, apiResponse{Success: false, Message: "Metode tidak diizinkan, gunakan POST"})
		return
	}

	var req ContactRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, apiResponse{Success: false, Message: "Format data tidak valid"})
		return
	}

	// Validasi dasar di sisi server (jangan hanya mengandalkan validasi di frontend).
	req.Name = strings.TrimSpace(req.Name)
	req.Email = strings.TrimSpace(req.Email)
	req.Message = strings.TrimSpace(req.Message)

	if req.Name == "" || req.Email == "" || req.Message == "" {
		writeJSON(w, http.StatusBadRequest, apiResponse{Success: false, Message: "Nama, email, dan pesan wajib diisi"})
		return
	}
	if !strings.Contains(req.Email, "@") {
		writeJSON(w, http.StatusBadRequest, apiResponse{Success: false, Message: "Format email tidak valid"})
		return
	}

	// TODO produksi: kirim email notifikasi (misalnya lewat SMTP/SendGrid)
	// atau simpan ke database (PostgreSQL/MySQL) alih-alih hanya mencetak ke log.
	log.Printf("[PESAN BARU] %s | %s | waktu: %s\nPesan: %s\n",
		req.Name, req.Email, time.Now().Format(time.RFC3339), req.Message)

	writeJSON(w, http.StatusOK, apiResponse{
		Success: true,
		Message: "Pesan berhasil dikirim, terima kasih!",
	})
}

// writeJSON adalah helper agar tidak menulis ulang boilerplate encoding JSON di setiap handler.
func writeJSON(w http.ResponseWriter, statusCode int, payload apiResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("Gagal menulis respons JSON: %v", err)
	}
}
