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

func indexHandler(w http.ResponseWriter, r *http.Request) {
	dc := r.Context().Value("datastore_client").(*datastore.Client)
	dc = dc
	fmt.Fprint(w, "Hello, World!")
}

func main() {

	ctx := context.Background()

	client, err := datastore.NewClient(ctx, datastore.DetectProjectID)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	kind := "User"
	name := "sampleUser"
	userKey := datastore.NameKey(kind, name, nil)

	user := User{
		Description: "This is a dummy user",
	}

	if _, err := client.Put(ctx, userKey, &user); err != nil {
		log.Fatalf("Failed to save task: %v\n", err)
	}

	r := chi.NewRouter()
	r.Use(middleware.WithValue("datastore_client", client))

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
