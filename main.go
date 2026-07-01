package main

import (
	"context"
	"embed"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"text/template"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/joho/godotenv"
	"github.com/mrchip53/simoni.wang/database"
)

var (
	//go:embed static/*
	staticFS embed.FS

	//go:embed templates
	templateFS embed.FS
)

type CreateLinkRequest struct {
	URL string `json:"url"`
}

func main() {
	godotenv.Load()

	ctx := context.Background()

	tmpl, err := template.ParseFS(templateFS, "templates/*.html")
	if err != nil {
		log.Fatal("Unable to parse templates:", err)
	}

	conn, err := pgx.Connect(ctx, os.Getenv("POSTGRES_URL"))
	if err != nil {
		log.Fatal("Unable to connect to database:", err)
	}
	defer conn.Close(ctx)

	queries := database.New(conn)

	mux := http.NewServeMux()

	// Serve static files
	mux.Handle("/static/", http.FileServer(http.FS(staticFS)))

	// Landing page
	mux.HandleFunc("/{$}", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Maybe I will put something here in the future."))
	})

	// Create link API
	mux.HandleFunc("POST /api/create", func(w http.ResponseWriter, r *http.Request) {
		body := CreateLinkRequest{}
		err := json.NewDecoder(r.Body).Decode(&body)
		if err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if body.URL == "" {
			http.Error(w, "URL is required", http.StatusBadRequest)
			return
		}

		res, err := queries.CreateLink(r.Context(), body.URL)
		if err != nil {
			http.Error(w, "Failed to create link", http.StatusInternalServerError)
			return
		}

		response := map[string]string{
			"id": encodeUUIDBytesToBase62(res.ID.Bytes),
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	// Link resolution
	mux.HandleFunc("/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		uuidBytes := decodeBase62ToUUID(id)

		res, err := queries.GetLink(r.Context(), pgtype.UUID{Bytes: uuidBytes, Valid: true})
		if err != nil {
			http.Error(w, "Link not found", http.StatusNotFound)
			return
		}

		err = tmpl.ExecuteTemplate(w, "link.html", map[string]any{
			"Url": res.Url,
		})
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	})

	log.Fatal(http.ListenAndServe(":8080", mux))
}
