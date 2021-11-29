package main

import (
	"context"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/abiosoft/ishell/v2"
	"github.com/alexedwards/scs/v2"
	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/r3labs/sse/v2"
	"golang.org/x/oauth2"
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
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = true

	mux := http.NewServeMux()
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	loginType := "implicit-login"
	dashboardType := map[string]string{
		"implicit-login":        "dashboard.html",
		"explicit-login":        "dashboard.html",
		"explicit-login-window": "dashboard-close-login.html",
	}

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		templates := template.Must(template.ParseGlob("templates/*.html"))

		email := session.GetString(r.Context(), "email")

		if email == "" {
			// no user in session

			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		// user exists; show the sauce!
		err := templates.ExecuteTemplate(w, dashboardType[loginType], map[string]string{
			"user": strings.Split(email, "@")[0],
		})
		if err != nil {
			log.Println(err)
		}
	})

	mux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		if loginType == "implicit-login" {
			http.Redirect(w, r, oauth2Config.AuthCodeURL("abcdefghijkl"), http.StatusFound)
			return
		}

		templates := template.Must(template.ParseGlob("templates/*.html"))
		err := templates.ExecuteTemplate(w, loginType+".html", map[string]string{
			"url": oauth2Config.AuthCodeURL("abcdefghijkl"),
		})
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

	events := sse.New()
	events.AutoReplay = false
	events.AutoStream = false
	events.CreateStream("messages")

	embed := http.NewServeMux()
	embed.Handle("/", http.FileServer(http.Dir("./static")))

	combined := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.Host, "dashboards.localtest.me") {
			embed.ServeHTTP(w, r)
			return
		}

		sessionH := session.LoadAndSave(mux)
		if r.URL.Path == "/events" {
			events.ServeHTTP(w, r)
			return
		}
		sessionH.ServeHTTP(w, r)
	})

	go func() {
		err := http.ListenAndServeTLS(":8443", "./cert.pem", "./cert-key.pem", combined)
		if err != nil {
			log.Fatal(err)
		}
	}()

	shell := ishell.New()
	shell.AddCmd(&ishell.Cmd{
		Name: "same-site",
		Help: "set cookie samesite",
		Func: func(c *ishell.Context) {

			options := []string{"lax", "strict", "none"}
			choice := c.MultiChoice(options, "Set Cookie SameSite")
			c.Println()

			switch choice {
			case 0:
				session.Cookie.SameSite = http.SameSiteLaxMode
			case 1:
				session.Cookie.SameSite = http.SameSiteStrictMode
			case 2:
				session.Cookie.SameSite = http.SameSiteNoneMode
			}

			err := session.Iterate(context.Background(), func(c context.Context) error {
				return session.Destroy(c)
			})
			if err != nil {
				log.Println(err)
			}

			events.Publish("messages", &sse.Event{
				Data: []byte("samesite=" + options[choice]),
			})
		},
	})
	shell.AddCmd(&ishell.Cmd{
		Name: "login-type",
		Help: "set login type",
		Func: func(c *ishell.Context) {

			options := []string{"implicit-login", "explicit-login", "explicit-login-window"}
			choice := c.MultiChoice(options, "Set Login Type")
			c.Println()

			loginType = options[choice]

			err := session.Iterate(context.Background(), func(c context.Context) error {
				return session.Destroy(c)
			})
			if err != nil {
				log.Println(err)
			}

			events.Publish("messages", &sse.Event{
				Data: []byte("login-type=" + loginType),
			})
		},
	})

	shell.Run()
}
