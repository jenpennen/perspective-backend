package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
)

type apiserver struct {
	pool *pgxpool.Pool
	
}

func (a *apiserver) build() http.Handler {
	router:=chi.NewRouter()
	router.Use(middleware.Logger)
	api:= newHandler(a.pool).build()
	router.Mount("/api",api)
	return router
}

func newAPIServer(pool *pgxpool.Pool) (*apiserver) {
	return &apiserver{pool: pool}
}

