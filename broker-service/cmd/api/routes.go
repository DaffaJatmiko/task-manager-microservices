package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/go-chi/chi/v5/middleware"
)

func (app *Config) routes() http.Handler {
	mux := chi.NewRouter()

	//specify who is allowed to connect
	mux.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))

	mux.Use(middleware.Heartbeat("/ping"))

	mux.Post("/", app.Broker)
	mux.Post("/log-grpc", app.LogViaGRPC)
	mux.Post("/handle", app.HandleSubmission)

	mux.With(JWTMiddleware).Route("/handle-task", func(r  chi.Router){
		r.Post("/", app.HandleTaskService)
		r.Put("/", app.HandleTaskService)
		r.Delete("/", app.HandleTaskService)
		r.Get("/", app.HandleTaskService)
	}) 


	return mux
}