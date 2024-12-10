package main

import (
	"encoding/json"
	"net/http"
	"sync"
)

type CartItem struct {
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
}

type Order struct {
	ID       string     `json:"id"`
	UserID   string     `json:"user_id"`
	Items    []CartItem `json:"items"`
	Total    float64    `json:"total"`
	Status   string     `json:"status"`
}

var (
	carts   = make(map[string][]CartItem)
	orders  = make(map[string]Order)
	mutex   = sync.RWMutex{}
	counter = 1
)

func main() {
	http.HandleFunc("/cart", handleCart)
	http.HandleFunc("/checkout", handleCheckout)

	http.ListenAndServe(":8081", nil)
}

func handleCart(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		mutex.RLock()
		cart := carts[userID]
		mutex.RUnlock()
		json.NewEncoder(w).Encode(cart)

	case http.MethodPost:
		var item CartItem
		if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		mutex.Lock()
		carts[userID] = append(carts[userID], item)
		mutex.Unlock()

		w.WriteHeader(http.StatusCreated)
	}
}

func handleCheckout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var order Order
	if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	mutex.Lock()
	order.ID = string(counter)
	counter++
	order.Status = "Confirmed"
	orders[order.ID] = order
	mutex.Unlock()

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(order)
}
