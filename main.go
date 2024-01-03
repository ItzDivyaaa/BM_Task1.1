package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
)

// User represents a user record
type User struct {
	ID         string      `json:"id"`
	SecretCode string      `json:"secretCode"`
	Name       string      `json:"name"`
	Email      string      `json:"email"`
	Complaints []Complaint `json:"complaints"`
}

type Complaint struct {
	ID         string `json:"id"`
	Title      string `json:"title"`
	Summary    string `json:"summary"`
	Severity   int    `json:"severity"`
	Resolved   bool   `json:"resolved"`
	SecretCode string
}

var mu sync.Mutex

var users = make(map[string]User)

var complaints = make(map[string]Complaint)

func main() {
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/register", registerHandler)
	http.HandleFunc("/submitComplaint", submitComplaintHandler)
	http.HandleFunc("/getAllComplaintsForUser", getAllComplaintsForUserHandler)
	http.HandleFunc("/getAllComplaintsForAdmin", getAllComplaintsForAdminHandler)
	http.HandleFunc("/viewComplaint", viewComplaintHandler)
	http.HandleFunc("/resolveComplaint", resolveComplaintHandler)

	fmt.Println("Server is running on :8080...")
	http.ListenAndServe(":8080", nil)
}

func writeError(w http.ResponseWriter, errMsg string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	errorMessage := map[string]string{"error": errMsg}
	json.NewEncoder(w).Encode(errorMessage)
}

func generateUniqueID() string {
	return fmt.Sprintf("%d", len(complaints)+1)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	var credentials struct {
		SecretCode string `json:"secretCode"`
	}

	if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil {
		writeError(w, err.Error(), http.StatusBadRequest)
		return
	}

	user, exists := users[credentials.SecretCode]
	if !exists {
		writeError(w, "User not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(user)
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	var newUser User

	if err := json.NewDecoder(r.Body).Decode(&newUser); err != nil {
		writeError(w, err.Error(), http.StatusBadRequest)
		return
	}

	if _, exists := users[newUser.SecretCode]; exists {
		writeError(w, "Secret code already in use", http.StatusBadRequest)
		return
	}

	newUser.ID = generateUniqueID()

	newUser.Complaints = []Complaint{}

	users[newUser.SecretCode] = newUser

	json.NewEncoder(w).Encode(newUser)
}

func submitComplaintHandler(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	var newComplaint Complaint

	if err := json.NewDecoder(r.Body).Decode(&newComplaint); err != nil {
		writeError(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Check if the user exists
	user, exists := users[newComplaint.SecretCode]
	if !exists {
		writeError(w, "User not found", http.StatusNotFound)
		return
	}

	newComplaint.ID = generateUniqueID()

	complaints[newComplaint.ID] = newComplaint

	user.Complaints = append(user.Complaints, newComplaint)
	users[newComplaint.SecretCode] = user

	w.WriteHeader(http.StatusCreated)
}

func getAllComplaintsForUserHandler(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	var user struct {
		SecretCode string `json:"secretCode"`
	}

	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		writeError(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Check if the user exists
	userDetails, exists := users[user.SecretCode]
	if !exists {
		writeError(w, "User not found", http.StatusNotFound)
		return
	}

	// Return all complaints for the user
	json.NewEncoder(w).Encode(userDetails.Complaints)
}

func getAllComplaintsForAdminHandler(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	var adminCredentials struct {
		SecretCode string `json:"secretCode"`
	}

	if err := json.NewDecoder(r.Body).Decode(&adminCredentials); err != nil {
		writeError(w, err.Error(), http.StatusBadRequest)
		return
	}

	if adminCredentials.SecretCode != "admin" {
		writeError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Return all complaints for administrators
	var allComplaints []Complaint
	for _, user := range users {
		allComplaints = append(allComplaints, user.Complaints...)
	}

	json.NewEncoder(w).Encode(allComplaints)
}

func viewComplaintHandler(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	var complaint struct {
		ID string `json:"id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&complaint); err != nil {
		writeError(w, err.Error(), http.StatusBadRequest)
		return
	}

	complaintDetails, exists := complaints[complaint.ID]
	if !exists {
		writeError(w, "Complaint not found", http.StatusNotFound)
		return
	}

	_, exists = users[complaintDetails.SecretCode]
	if !exists {
		writeError(w, "User not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(complaintDetails)
}

func resolveComplaintHandler(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	var complaint struct {
		ID string `json:"id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&complaint); err != nil {
		writeError(w, err.Error(), http.StatusBadRequest)
		return
	}

	var adminCredentials struct {
		SecretCode string `json:"secretCode"`
	}

	if err := json.NewDecoder(r.Body).Decode(&adminCredentials); err != nil {
		writeError(w, err.Error(), http.StatusBadRequest)
		return
	}

	if adminCredentials.SecretCode != "admin" {
		writeError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Check if the complaint exists
	complaintDetails, exists := complaints[complaint.ID]
	if !exists {
		writeError(w, "Complaint not found", http.StatusNotFound)
		return
	}

	complaintDetails.Resolved = true
	complaints[complaint.ID] = complaintDetails

	w.WriteHeader(http.StatusNoContent)
}
