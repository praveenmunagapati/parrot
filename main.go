package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/anthonynsimon/parrot/api"
	"github.com/anthonynsimon/parrot/api/auth"
	"github.com/anthonynsimon/parrot/datastore"
	"github.com/anthonynsimon/parrot/logger"
	"github.com/joho/godotenv"
	"github.com/pressly/chi"
	"github.com/pressly/chi/middleware"
)

func init() {
	// Config log
	log.SetFormatter(&log.TextFormatter{})
}

func main() {
	// init environment variables
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}

	// init and ping datastore
	dbName := os.Getenv("DB")
	dbURL := os.Getenv("DB_URL")

	ds, err := datastore.NewDatastore(dbName, dbURL)
	if err != nil {
		log.Fatal(err)
	}

	defer ds.Close()
	if err = ds.Ping(); err != nil {
		log.Fatal(fmt.Sprintf("failed to ping datastore.\nerr: %s", err))
	}

	// init routers and middleware
	mainRouter := chi.NewRouter()
	mainRouter.Use(logger.Request)
	mainRouter.Use(middleware.StripSlashes)

	ap := auth.Provider{
		Name:       string([]byte(os.Getenv("DOMAIN"))),
		SigningKey: []byte(os.Getenv("API_SIGNING_KEY"))}

	apiRouter := api.NewRouter(ds, ap)
	mainRouter.Mount("/api", apiRouter)

	// config server
	addr := "localhost:8080"

	// init server
	s := &http.Server{
		Addr:           addr,
		Handler:        mainRouter,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	fmt.Println(fmt.Sprintf("Listening on %s", addr))

	log.Fatal(s.ListenAndServe())
}
