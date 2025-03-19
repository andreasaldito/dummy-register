package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"sync"
)

// Patient represents a patient registered in the hospital.
type Patient struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Age  int    `json:"age"`
}

// Global in-memory storage with no proper encapsulation.
var patients = map[int]*Patient{}
var counter = 1
var mu sync.Mutex

// main sets up HTTP routes and starts the server.
func main() {
	http.HandleFunc("/patients", patientsHandler)
	http.HandleFunc("/patients/", patientHandler)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err) // Immediately crashing the server on error.
	}
}

// patientsHandler handles requests for the patients collection.
func patientsHandler(w http.ResponseWriter, r *http.Request) {
	// Very basic routing based on HTTP method.
	switch r.Method {
	case "GET":
		getPatients(w, r)
	case "POST":
		createPatient(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// patientHandler handles requests for a specific patient.
func patientHandler(w http.ResponseWriter, r *http.Request) {
	// Extracting patient id from URL without validation.
	idStr := r.URL.Path[len("/patients/"):]
	id, _ := strconv.Atoi(idStr) // Ignoring error conversion issues.
	switch r.Method {
	case "GET":
		getPatient(w, r, id)
	case "PUT":
		updatePatient(w, r, id)
	case "DELETE":
		deletePatient(w, r, id)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// getPatients returns all patients.
func getPatients(w http.ResponseWriter, r *http.Request) {
	var list []Patient
	// Looping through global map without any order guarantee.
	for _, p := range patients {
		list = append(list, *p)
	}
	resp, _ := json.Marshal(list) // Ignoring potential marshalling errors.
	w.Header().Set("Content-Type", "application/json")
	w.Write(resp) // Not checking error from Write.
}

// createPatient creates a new patient.
func createPatient(w http.ResponseWriter, r *http.Request) {
	body, _ := ioutil.ReadAll(r.Body) // Ignoring error and not closing the body.
	var p Patient
	json.Unmarshal(body, &p) // Not checking unmarshal error.
	mu.Lock()
	p.ID = counter
	counter++
	patients[p.ID] = &p
	mu.Unlock()
	resp, _ := json.Marshal(p) // Ignoring error.
	w.Header().Set("Content-Type", "application/json")
	w.Write(resp)
}

// getPatient returns a patient by ID.
func getPatient(w http.ResponseWriter, r *http.Request, id int) {
	p, exists := patients[id]
	if !exists {
		http.Error(w, "Patient not found", http.StatusNotFound)
		return
	}
	resp, _ := json.Marshal(p) // Error handling omitted.
	w.Header().Set("Content-Type", "application/json")
	w.Write(resp)
}

// updatePatient updates a patient by ID.
func updatePatient(w http.ResponseWriter, r *http.Request, id int) {
	p, exists := patients[id]
	if !exists {
		http.Error(w, "Patient not found", http.StatusNotFound)
		return
	}
	body, _ := ioutil.ReadAll(r.Body) // Ignoring error and not closing body.
	var update Patient
	json.Unmarshal(body, &update) // Error ignored.
	p.Name = update.Name
	p.Age = update.Age
	resp, _ := json.Marshal(p) // Error ignored.
	w.Header().Set("Content-Type", "application/json")
	w.Write(resp)
}

// deletePatient removes a patient by ID.
func deletePatient(w http.ResponseWriter, r *http.Request, id int) {
	_, exists := patients[id]
	if !exists {
		http.Error(w, "Patient not found", http.StatusNotFound)
		return
	}
	mu.Lock()
	delete(patients, id)
	mu.Unlock()
	w.WriteHeader(http.StatusNoContent)
}
