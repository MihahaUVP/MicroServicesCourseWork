package main

import (
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

var audiences = []*Audience{
	{Number: 1317, Status: "free"},
	{Number: 1321, Status: "free"},
	{Number: 1318, Status: "free"},
	{Number: 1319, Status: "free"},
	{Number: 1306, Status: "buzy"},
	{Number: 1313, Status: "buzy"},
}
var users = []*User{
	{Number: 100, Name: "Misha", Balance: 10},
	{Number: 101, Name: "Ashim", Balance: 10},
	{Number: 102, Name: "Masha", Balance: 12},
	{Number: 103, Name: "Sasha", Balance: -10},
	{Number: 104, Name: "Pasha", Balance: -3},
	{Number: 105, Name: "Glasha", Balance: 10},
}

func GetUsers(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("balance")
	if status != "" {
		balance, err := strconv.Atoi(status)
		if err == nil {
			filteredUsers := []User{}
			for _, user := range users {
				if user.Balance >= balance {
					filteredUsers = append(filteredUsers, *user)
				}
			}
			json.NewEncoder(w).Encode(filteredUsers)
		}
	} else {
		json.NewEncoder(w).Encode(users)
	}
}
func UpdateUserBalance(w http.ResponseWriter, r *http.Request) {
	number, _ := strconv.Atoi(chi.URLParam(r, "number"))

	var updatedBalance struct {
		Balance int `json:"balance"`
	}

	err := json.NewDecoder(r.Body).Decode(&updatedBalance)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	for i, user := range users {
		if number == user.Number {
			users[i].Balance = updatedBalance.Balance
			json.NewEncoder(w).Encode(users[i])
			return
		}
	}

	http.NotFound(w, r)
}

func main() {

	logger := httplog.NewLogger("service-one", httplog.Options{
		JSON: true,
	})

	r := chi.NewRouter()
	r.Use(chiprometheus.NewMiddleware("service-one"))
	r.Use(middleware.RealIP)
	r.Use(httplog.RequestLogger(logger))
	r.Use(middleware.Recoverer)

	r.Handle("/metrics", promhttp.Handler())
	r.Get("/users", GetUsers)
	//r.Get("/audiences", GetAudiences)
	//r.Post("/audiences/{number}", UpdateAudienceStatus)
	r.Post("/users/{number}", UpdateUserBalance)
	http.ListenAndServe(":8081", r)
}
