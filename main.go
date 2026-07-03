package main

import (
	"context"
	"embed"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"text/template"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
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
	URL  string  `json:"url"`
	Slug *string `json:"slug,omitempty"`
}

type CreateLinkResponse struct {
	ID   string  `json:"id"`
	Slug *string `json:"slug,omitempty"`
}

func main() {
	godotenv.Load()

	ctx := context.Background()

	tmpl, err := template.ParseFS(templateFS, "templates/*.html")
	if err != nil {
		log.Fatal("Unable to parse templates:", err)
	}

	pool, err := pgxpool.New(ctx, os.Getenv("POSTGRES_URL"))
	if err != nil {
		log.Fatal("Unable to create connection pool:", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatal("Unable to ping database pool:", err)
	}

	queries := database.New(pool)

	mux := http.NewServeMux()

	// Serve static files
	mux.Handle("/static/", http.FileServer(http.FS(staticFS)))

	// Landing page
	mux.HandleFunc("/{$}", func(w http.ResponseWriter, r *http.Request) {
		links, err := queries.GetLastNLinks(r.Context(), 10)
		if err != nil {
			http.Error(w, "Failed to fetch links", http.StatusInternalServerError)
			return
		}

		encodedLinks := make([]map[string]any, len(links))
		for i, link := range links {
			timeStr := link.CreatedAt.Time.Format("2006-01-02 15:04:05")
			encodedLinks[i] = map[string]any{
				"ID":        encodeUUIDBytesToBase62(link.ID.Bytes),
				"Url":       link.Url,
				"Slug":      PgTextToPointerString(link.Slug),
				"CreatedAt": timeStr,
			}
		}

		err = tmpl.ExecuteTemplate(w, "index.html", encodedLinks)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	})

	// Delete link API
	mux.HandleFunc("POST /api/delete/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		uuidBytes := decodeBase62ToUUID(id)

		err := queries.DeleteLink(r.Context(), pgtype.UUID{Bytes: uuidBytes, Valid: true})
		if err != nil {
			http.Error(w, "Failed to delete link", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)
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

		res, err := queries.CreateLink(r.Context(), database.CreateLinkParams{
			Url:  body.URL,
			Slug: PointerStringToPgText(body.Slug),
		})
		if err != nil {
			http.Error(w, "Failed to create link", http.StatusInternalServerError)
			return
		}

		response := CreateLinkResponse{
			ID:   encodeUUIDBytesToBase62(res.ID.Bytes),
			Slug: PgTextToPointerString(res.Slug),
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	// Link resolution
	mux.HandleFunc("/s/{slug}", func(w http.ResponseWriter, r *http.Request) {
		slug := r.PathValue("slug")

		if slug == "" {
			http.Error(w, "Slug is required", http.StatusBadRequest)
			return
		}

		res, err := queries.GetLinkBySlug(r.Context(), PointerStringToPgText(&slug))
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
