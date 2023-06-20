package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httplog"
	chiprometheus "github.com/nathan-jones/chi-prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Audience struct {
	Number int    `json:"number"`
	Status string `json:"status"`
}
type User struct {
	Number  int    `json:"number"`
	Name    string `json:"name"`
	Balance int    `json:"balance"`
}
type CheckUserRequest struct {
	Balance int `json:"balance"`
}
type ReserveAudienceRequest struct {
	Status string `json:"status"`
}

func GetActiveUsers() ([]User, error) {
	resp, err := http.Get("http://host.docker.internal:8081/users?balance=0")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, err
	}

	var response []User
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return nil, err
	}

	return response, nil
}
func ActivateUser(number int) (*User, error) {
	activationRequest := CheckUserRequest{
		Balance: 10,
	}

	jsonBytes, err := json.Marshal(activationRequest)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post("http://host.docker.internal:8081/users/"+strconv.Itoa(number), "application/json", bytes.NewReader(jsonBytes))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	} else if resp.StatusCode != http.StatusOK {
		return nil, err
	}

	var user User
	err = json.NewDecoder(resp.Body).Decode(&user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func main() {

	logger := httplog.NewLogger("service-two", httplog.Options{
		JSON: true,
	})

	r := chi.NewRouter()
	r.Use(chiprometheus.NewMiddleware("service-two"))
	r.Use(middleware.RealIP)
	r.Use(httplog.RequestLogger(logger))
	r.Use(middleware.Recoverer)

	r.Handle("/metrics", promhttp.Handler())

	r.Get("/active-users", func(w http.ResponseWriter, r *http.Request) {
		users, err := GetActiveUsers()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(users)
	})

	r.Post("/activate-user/{number}", func(w http.ResponseWriter, r *http.Request) {
		number, err := strconv.Atoi(chi.URLParam(r, "number"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		user, err := ActivateUser(number)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if user == nil {
			http.NotFound(w, r)
			return
		}

		json.NewEncoder(w).Encode(user)
	})

	http.ListenAndServe(":8082", r)
}
