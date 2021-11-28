package main

import (
	"context"
	"github.com/alexedwards/scs/v2"
	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"
)

func main() {
	provider, err := oidc.NewProvider(context.Background(), os.Getenv("OAUTH2_SERVER"))
	if err != nil {
		log.Fatal(err)
	}
	oauth2Config := oauth2.Config{
		ClientID:     os.Getenv("OAUTH2_CLIENT_ID"),
		ClientSecret: os.Getenv("OAUTH2_CLIENT_SECRET"),
		Endpoint:     provider.Endpoint(),
		RedirectURL:  os.Getenv("OAUTH2_REDIRECT_URL"),
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}
	idTokenVerifier := provider.Verifier(&oidc.Config{ClientID: oauth2Config.ClientID})

	session := scs.New()
	session.Cookie = scs.SessionCookie{
		Name:     "session",
		HttpOnly: true,
		Path:     "/",
		Persist:  true,
		SameSite: http.SameSiteLaxMode,
		Secure:   true,
	}

	mux := http.NewServeMux()
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		templates := template.Must(template.ParseGlob("templates/*.html"))

		email := session.GetString(r.Context(), "email")

		var err error

		if email == "" {
			// no user in session

			err = templates.ExecuteTemplate(w, "login-redirect.html", map[string]string{
				"url": oauth2Config.AuthCodeURL("abcdefghijkl"),
			})
		} else {
			// user exists; show the sauce!

			err = templates.ExecuteTemplate(w, "dashboard.html", map[string]string{
				"user": strings.Split(email, "@")[0],
			})
		}

		if err != nil {
			log.Println(err)
		}
	})

	mux.HandleFunc("/oauth2/callback", func(w http.ResponseWriter, r *http.Request) {
		ctx := context.Background()
		oauth2Token, err := oauth2Config.Exchange(ctx, r.URL.Query().Get("code"))
		if err != nil {
			log.Println(err)
			http.Error(w, "token exchange failed", http.StatusInternalServerError)
			return
		}

		rawIDToken, ok := oauth2Token.Extra("id_token").(string)
		if !ok {
			http.Error(w, "missing id_token", http.StatusInternalServerError)
			return
		}

		idToken, err := idTokenVerifier.Verify(ctx, rawIDToken)
		if err != nil {
			log.Println(err)
			http.Error(w, "unverified id_token", http.StatusInternalServerError)
			return
		}

		var claims struct {
			Email string `json:"email"`
		}
		if err := idToken.Claims(&claims); err != nil {
			log.Println(err)
			http.Error(w, "failed to parse claims", http.StatusInternalServerError)
			return
		}

		session.Put(r.Context(), "email", claims.Email)
		http.Redirect(w, r, "/", http.StatusFound)
	})

	handlers := session.LoadAndSave(mux)

	err = http.ListenAndServeTLS(":8443", "./cert.pem", "./cert-key.pem", handlers)
	if err != nil {
		log.Fatal(err)
	}
}
