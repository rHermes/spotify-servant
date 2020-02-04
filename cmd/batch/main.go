package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"cloud.google.com/go/datastore"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

type User struct {
	Description string
}

// cronOnly defines a handler that only allows google cron jobs
func cronOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Appengine-Cron") != "true" {
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	dc := r.Context().Value("datastore_client").(*datastore.Client)
	dc = dc
	fmt.Fprint(w, "Hello, World, from batch with love!")
}

func main() {
	ctx := context.Background()

	// This is because there is a bug in the
	// https://github.com/googleapis/google-cloud-go/issues/1751
	projectID := datastore.DetectProjectID
	if os.Getenv("RUN_WITH_DEVAPPSERVER") == "1" {
		projectID = "asdsad"
	}

	client, err := datastore.NewClient(ctx, projectID)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	r := chi.NewRouter()
	r.Use(middleware.WithValue("datastore_client", client))
	r.Use(cronOnly)

	r.Get("/", indexHandler)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("Defaulting to port %s", port)
	}

	log.Printf("Listening on port %s", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatal(err)
	}

}
