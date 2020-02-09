package main

import (
	"bytes"
	"context"
	"github.com/alecthomas/repr"
	"log"
	"net/http"
	"os"
	"time"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/auth"

	"cloud.google.com/go/datastore"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

type User struct {
	Age         int8
	Description string
	K           *datastore.Key `datastore:"__key__"`
}

func getIdTokenFromBody(r *http.Request) (string, error) {
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r.Body); err != nil {
		return "", err
	}
	return buf.String(), nil
}


// MustAuthMiddleware ensures that the user is authenticated using the
// session token, if not it redirects to login
func MustAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		fa := ctx.Value("firebase_auth").(*auth.Client)

		cookie, err := r.Cookie("session")
		if err != nil {
			if err != http.ErrNoCookie {
				http.Error(w, "Internal Problem", http.StatusInternalServerError)
				return
			}
			// TODO(rHermes): Add a "destination url" here to redirect after login
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		// decoded, err := fa.VerifySessionCookieAndCheckRevoked(ctx, cookie.Value)
		decoded, err := fa.VerifySessionCookie(ctx, cookie.Value)
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}


		// Here we apply the
		ctx = context.WithValue(ctx,"sess_decoded", decoded)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}


func sessionLoginHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	fa := ctx.Value("firebase_auth").(*auth.Client)

	idToken, err := getIdTokenFromBody(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	decoded, err := fa.VerifyIDToken(ctx, idToken)
	if err != nil {
		http.Error(w, "Invalid ID token", http.StatusUnauthorized)
		return
	}

	if time.Now().Unix()-decoded.AuthTime > 5*60 {
		http.Error(w, "Recent sign-in required", http.StatusUnauthorized)
		return
	}

	expiresIn := time.Hour * 24 * 5
	cookie, err := fa.SessionCookie(ctx, idToken, expiresIn)
	if err != nil {
		http.Error(w, "Failed to create a session cookie", http.StatusInternalServerError)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    cookie,
		MaxAge:   int(expiresIn.Seconds()),
		HttpOnly: true,
		Secure:   true,
	})

	//pr := repr.New(w)
	//pr.Println(decoded)
}

func loginPageHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	t := GetTemplateFromCtx(ctx, LoginTemplateName)

	bctx := TemplateBaseCtx{
		Firebase: ctx.Value("firebase_cfg").(FirebaseConfig),
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, bctx); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	buf.WriteTo(w)
}

func logoutPageHandler(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name: "session",
		Value: "",
		MaxAge: 0,
	})
	http.Redirect(w, r, "/", http.StatusFound)
}

func profilePageHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	decoded := ctx.Value("sess_decoded").(*auth.Token)
	repr.Println(decoded)

	t := GetTemplateFromCtx(ctx, ProfileIndexTemplateName)

	bctx := TemplateBaseCtx{
		Firebase: ctx.Value("firebase_cfg").(FirebaseConfig),
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, bctx); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	buf.WriteTo(w)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	dc := ctx.Value("datastore_client").(*datastore.Client)
	dc = dc

	t := GetTemplateFromCtx(ctx, IndexTemplateName)

	bctx := TemplateBaseCtx{
		Firebase: ctx.Value("firebase_cfg").(FirebaseConfig),
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, bctx); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	buf.WriteTo(w)
}

func main() {
	ctx := context.Background()

	// We need to get a firebase app asswell
	app, err := firebase.NewApp(ctx, nil)
	if err != nil {
		log.Fatalf("We where not able to get a firebase app: %s\n", err.Error())
	}
	auth, err := app.Auth(ctx)
	if err != nil {
		log.Fatalf("Couldn't get authentication: %s\n", err.Error())
	}

	// Create firebase config
	fireCfg := FirebaseConfig{
		ApiKey:            os.Getenv("FIREBASE_API_KEY"),
		AuthDomain:        os.Getenv("FIREBASE_AUTH_DOMAIN"),
		DatabaseURL:       os.Getenv("FIREBASE_DB_URL"),
		ProjectID:         os.Getenv("FIREBASE_PROJECT_ID"),
		StorageBucket:     os.Getenv("FIREBASE_STORAGE_BUCKET"),
		MessagingSenderID: os.Getenv("FIREBASE_MESSAGING_SENDER_ID"),
		AppID:             os.Getenv("FIREBASE_APP_ID"),
		MeasurementID:     os.Getenv("FIREBASE_MEASUREMENT_ID"),
	}

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
	r.Use(middleware.WithValue("firebase_auth", auth))
	r.Use(middleware.WithValue("firebase_cfg", fireCfg))
	r.Use(TemplateMiddleware)

	r.Get("/", indexHandler)
	r.Get("/login", loginPageHandler)
	r.Get("/logout", logoutPageHandler)
	r.Post("/sessionLogin", sessionLoginHandler)

	r.Route("/profile", func(r chi.Router) {
		r.Use(MustAuthMiddleware)

		r.Get("/", profilePageHandler)
	})

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
