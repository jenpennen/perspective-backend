package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type handler struct {
	pool *pgxpool.Pool
	
}
func newHandler(pool *pgxpool.Pool) *handler {
	return &handler{pool : pool}
}

func (h *handler) health(w http.ResponseWriter, _ *http.Request) {
	res := struct {
		Status string `json:"status"`
	}{
		Status: "healthy",
	}

	bytes, err := json.Marshal(res)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(bytes)
}

func (h *handler)build() http.Handler {
	router := chi.NewRouter()
	router.HandleFunc("/health", h.health)
	router.Get("/users", h.getUsersByFirstName)
	return router

} 

func (h *handler) getUsersByFirstName(w http.ResponseWriter, r *http.Request) {
	params := struct {
		Value string `json:"value"`
		// Value string `json:"value"`
	}{
	}
	fmt.Println("here1")
	err := json.NewDecoder(r.Body).Decode(&params)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	fmt.Println("here2")
	users, err := getUsersByFirstName(h.pool, params.Value)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	
	res := struct {
		Users []User `json:"users"`
	}{Users:users,}
	bytes, err := json.Marshal(res)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(bytes)
}


