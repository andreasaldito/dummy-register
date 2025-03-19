package main

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"
	"regexp"
	"strconv"
	"sync"
)

// Patient represents a patient registered in the hospital.
type Patient struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Age      int    `json:"age"`
	Password string `json:"password"`
}

// Global in-memory storage with no proper encapsulation.
var patients = map[int]*Patient{}
var counter = 1
var mu sync.Mutex

func main() {
	// Original routes
	http.HandleFunc("/patients", patientsHandler)
	http.HandleFunc("/patients/", patientHandler)
	http.HandleFunc("/insecure", insecureHandler)

	// Duplicated routes (same logic, different endpoints)
	http.HandleFunc("/patients-dup", patientsHandlerDup)
	http.HandleFunc("/patients-dup/", patientHandlerDup)
	http.HandleFunc("/insecure-dup", insecureHandlerDup)

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// ORIGINAL HANDLERS
// ─────────────────────────────────────────────────────────────────────────────

// patientsHandler handles requests for the patients collection.
func patientsHandler(w http.ResponseWriter, r *http.Request) {
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
	idStr := r.URL.Path[len("/patients/"):]
	id, _ := strconv.Atoi(idStr) // ignoring error

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
	for _, p := range patients {
		list = append(list, *p)
	}
	resp, _ := json.Marshal(list)
	w.Header().Set("Content-Type", "application/json")
	w.Write(resp)
}

// createPatient creates a new patient.
func createPatient(w http.ResponseWriter, r *http.Request) {
	body, _ := ioutil.ReadAll(r.Body) // ignoring error
	var p Patient
	json.Unmarshal(body, &p) // ignoring error

	// BAD PRACTICE: Regex is compiled on every call
	if !validatePatientName(p.Name) {
		http.Error(w, "Invalid patient name", http.StatusBadRequest)
		return
	}

	// BAD PRACTICE: Storing password using MD5 (insecure)
	hash := md5.Sum([]byte(p.Password))
	p.Password = hex.EncodeToString(hash[:])

	mu.Lock()
	p.ID = counter
	counter++
	patients[p.ID] = &p
	mu.Unlock()

	resp, _ := json.Marshal(p)
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
	resp, _ := json.Marshal(p)
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
	body, _ := ioutil.ReadAll(r.Body) // ignoring error
	var update Patient
	json.Unmarshal(body, &update) // ignoring error
	p.Name = update.Name
	p.Age = update.Age

	// Re-hash the password if it was updated (still insecure)
	if update.Password != "" {
		hash := md5.Sum([]byte(update.Password))
		p.Password = hex.EncodeToString(hash[:])
	}

	resp, _ := json.Marshal(p)
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

// validatePatientName uses a regex to validate a patient's name.
// Intentionally re-compiling the regex each time.
func validatePatientName(name string) bool {
	re, _ := regexp.Compile("^[A-Z][a-z]+(?:\\s[A-Z][a-z]+)*$")
	return re.MatchString(name)
}

// insecureHandler is a demonstration endpoint that performs a command injection
// by passing user input directly to 'exec.Command'. This is highly insecure
// and should be flagged by any decent security scanner.
func insecureHandler(w http.ResponseWriter, r *http.Request) {
	cmdStr := r.URL.Query().Get("cmd")
	if cmdStr == "" {
		http.Error(w, "No command provided", http.StatusBadRequest)
		return
	}

	cmd := exec.Command("sh", "-c", cmdStr)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	cmd.Run() // ignoring error

	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(out.String()))
}

// ─────────────────────────────────────────────────────────────────────────────
// DUPLICATED HANDLERS (CODE DUPLICATION ~100%)
// ─────────────────────────────────────────────────────────────────────────────

// patientsHandlerDup handles requests for the duplicated patients collection.
func patientsHandlerDup(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		getPatientsDup(w, r)
	case "POST":
		createPatientDup(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// patientHandlerDup handles requests for a specific duplicated patient.
func patientHandlerDup(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Path[len("/patients-dup/"):]
	id, _ := strconv.Atoi(idStr) // ignoring error

	switch r.Method {
	case "GET":
		getPatientDup(w, r, id)
	case "PUT":
		updatePatientDup(w, r, id)
	case "DELETE":
		deletePatientDup(w, r, id)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// getPatientsDup returns all duplicated patients.
func getPatientsDup(w http.ResponseWriter, r *http.Request) {
	var list []Patient
	for _, p := range patients {
		list = append(list, *p)
	}
	resp, _ := json.Marshal(list)
	w.Header().Set("Content-Type", "application/json")
	w.Write(resp)
}

// createPatientDup creates a new duplicated patient.
func createPatientDup(w http.ResponseWriter, r *http.Request) {
	body, _ := ioutil.ReadAll(r.Body) // ignoring error
	var p Patient
	json.Unmarshal(body, &p) // ignoring error

	// BAD PRACTICE: Regex is compiled on every call
	if !validatePatientName(p.Name) {
		http.Error(w, "Invalid patient name", http.StatusBadRequest)
		return
	}

	// BAD PRACTICE: Storing password using MD5 (insecure)
	hash := md5.Sum([]byte(p.Password))
	p.Password = hex.EncodeToString(hash[:])

	mu.Lock()
	p.ID = counter
	counter++
	patients[p.ID] = &p
	mu.Unlock()

	resp, _ := json.Marshal(p)
	w.Header().Set("Content-Type", "application/json")
	w.Write(resp)
}

// getPatientDup returns a duplicated patient by ID.
func getPatientDup(w http.ResponseWriter, r *http.Request, id int) {
	p, exists := patients[id]
	if !exists {
		http.Error(w, "Patient not found (dup)", http.StatusNotFound)
		return
	}
	resp, _ := json.Marshal(p)
	w.Header().Set("Content-Type", "application/json")
	w.Write(resp)
}

// updatePatientDup updates a duplicated patient by ID.
func updatePatientDup(w http.ResponseWriter, r *http.Request, id int) {
	p, exists := patients[id]
	if !exists {
		http.Error(w, "Patient not found (dup)", http.StatusNotFound)
		return
	}
	body, _ := ioutil.ReadAll(r.Body) // ignoring error
	var update Patient
	json.Unmarshal(body, &update) // ignoring error
	p.Name = update.Name
	p.Age = update.Age

	if update.Password != "" {
		hash := md5.Sum([]byte(update.Password))
		p.Password = hex.EncodeToString(hash[:])
	}

	resp, _ := json.Marshal(p)
	w.Header().Set("Content-Type", "application/json")
	w.Write(resp)
}

// deletePatientDup removes a duplicated patient by ID.
func deletePatientDup(w http.ResponseWriter, r *http.Request, id int) {
	_, exists := patients[id]
	if !exists {
		http.Error(w, "Patient not found (dup)", http.StatusNotFound)
		return
	}
	mu.Lock()
	delete(patients, id)
	mu.Unlock()
	w.WriteHeader(http.StatusNoContent)
}

// insecureHandlerDup is a duplicated demonstration endpoint that performs
// a command injection by passing user input directly to 'exec.Command'.
func insecureHandlerDup(w http.ResponseWriter, r *http.Request) {
	cmdStr := r.URL.Query().Get("cmd")
	if cmdStr == "" {
		http.Error(w, "No command provided (dup)", http.StatusBadRequest)
		return
	}

	cmd := exec.Command("sh", "-c", cmdStr)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	cmd.Run() // ignoring error

	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(out.String()))
}
