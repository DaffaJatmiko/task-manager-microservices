package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
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

	mux.Route("/tasks", func(r chi.Router) {
		r.Get("/", app.GetTasks)             // GET /tasks
		r.Post("/", app.CreateTask)          // POST /tasks
		r.Get("/userId", app.GetTask)          // GET /tasks/{id}
		r.Put("/update", app.UpdateTask)       // PUT /tasks/{id}
		r.Delete("/delete", app.DeleteTask)    // DELETE /tasks/{id}
	})

return mux
}