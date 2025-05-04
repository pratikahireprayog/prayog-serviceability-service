package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

func main() {
	// Define a simple handler function
	http.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "pong")
	})

	// Define a handler for serviceability checking
	http.HandleFunc("/api/v1/serviceability/check/", func(w http.ResponseWriter, r *http.Request) {
		postalCode := r.URL.Path[len("/api/v1/serviceability/check/"):]
		response := fmt.Sprintf(`{"postal_code": "%s", "is_serviceable": true}`, postalCode)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, response)
	})

	// Define a handler for bulk serviceability checking
	http.HandleFunc("/api/v1/serviceability/bulk-check", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var request struct {
			PostalCodes []string `json:"postal_codes"`
		}

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Error reading request body", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		if err := json.Unmarshal(body, &request); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		results := make(map[string]bool)
		for _, code := range request.PostalCodes {
			// Simple mock implementation - all postal codes are serviceable except "00000"
			results[code] = code != "00000"
		}

		response := map[string]interface{}{
			"results": results,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	// Start the server
	fmt.Println("Starting server at port 8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
