package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
	"strconv"
)

var pdaService *PDAService

/*
Method to return all the available pda in the system
*/
func pdaList(w http.ResponseWriter, r *http.Request) {
	respondWithJSON(w, http.StatusOK, pdaService.getAllAvailablePDAs())
}

/*
Method to create new pda in the system
*/
func createPDA(w http.ResponseWriter, r *http.Request) {
	//Accept PDA with ID and specification
	id, err := strconv.Atoi(parseRequestVariable(r, "id"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "ID is not an integer")
		return
	}

	// unmarshal body
	var pdaProcessor PDAProcessor
	_ = json.NewDecoder(r.Body).Decode(&pdaProcessor)
	pdaProcessor.ID = id

	// create pda processor
	createdPDA, err1 := pdaService.createNewPDA(id, pdaProcessor)
	if err1 != nil {
		respondWithError(w, http.StatusBadRequest, err1.Error())
		return
	}

	// return created pda processor
	respondWithJSON(w, http.StatusOK, createdPDA)
}

func resetPDA(w http.ResponseWriter, r *http.Request) {
	sessionId, pdaId, err := parseSessionIdAndPdaId(r)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	// reset pda
	err = pdaService.resetPDA(sessionId, pdaId)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	respondWithJSON(w, http.StatusOK, map[string]bool{"reset": true})
}

func isAccepted(w http.ResponseWriter, r *http.Request) {
	sessionId, pdaId, err := parseSessionIdAndPdaId(r)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	isAccepted, err := pdaService.isAccepted(sessionId, pdaId)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	respondWithJSON(w, http.StatusOK, map[string]bool{"is_accepted": isAccepted})
}

func peek(w http.ResponseWriter, r *http.Request) {
	sessionId, pdaId, err := parseSessionIdAndPdaId(r)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	k, err1 := strconv.Atoi(parseRequestVariable(r, "k"))
	if err1 != nil {
		respondWithError(w, http.StatusBadRequest, "k is not an integer")
		return
	}

	states, err2 := pdaService.peek(sessionId, pdaId, k)
	if err2 != nil {
		respondWithError(w, http.StatusBadRequest, err2.Error())
		return
	}
	respondWithJSON(w, http.StatusOK, states)
}

func stackLength(w http.ResponseWriter, r *http.Request) {
	sessionId, pdaId, err := parseSessionIdAndPdaId(r)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	length, err1 := pdaService.stackLength(sessionId, pdaId)
	if err1 != nil {
		respondWithError(w, http.StatusBadRequest, err1.Error())
		return
	}
	respondWithJSON(w, http.StatusOK, map[string]int{"stack_length": length})
}

func currentState(w http.ResponseWriter, r *http.Request) {
	sessionId, pdaId, err := parseSessionIdAndPdaId(r)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	state, err1 := pdaService.currentState(sessionId, pdaId)
	if err1 != nil {
		respondWithError(w, http.StatusBadRequest, err1.Error())
		return
	}
	respondWithJSON(w, http.StatusOK, map[string]string{"current_state": state})
}

func closePDA(w http.ResponseWriter, r *http.Request) {
	sessionId, pdaId, err := parseSessionIdAndPdaId(r)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	err = pdaService.closePDA(sessionId, pdaId)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	respondWithJSON(w, http.StatusOK, map[string]bool{"closed": true})
}

func deletePDA(w http.ResponseWriter, r *http.Request) {
	pdaId, err := strconv.Atoi(parseRequestVariable(r, "id"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "id is not an integer")
		return
	}

	err = pdaService.deletePDA(pdaId)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	respondWithJSON(w, http.StatusOK, map[string]bool{"deleted": true})
}

func presentToken(w http.ResponseWriter, r *http.Request) {
	// get session id and PDA id from header and path variables respectively
	sessionId, pdaId, err := parseSessionIdAndPdaId(r)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	// get position from the path variables
	position, err := strconv.Atoi(parseRequestVariable(r, "position"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "position is not an integer")
		return
	}

	// get token from the path variables
	token := parseRequestVariable(r, "token")

	isConsumed, err := pdaService.presentToken(sessionId, pdaId, token, position)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]bool{"is_consumed": isConsumed})
}

func queuedTokens(w http.ResponseWriter, r *http.Request) {
	sessionId, pdaId, err := parseSessionIdAndPdaId(r)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	pendingQueue, err1 := pdaService.getPendingQueue(sessionId, pdaId)
	if err1 != nil {
		respondWithError(w, http.StatusBadRequest, err1.Error())
		return
	}
	if pendingQueue == nil {
		pendingQueue = make([]string, 0)
	}
	respondWithJSON(w, http.StatusOK, pendingQueue)
}

func presentEOS(w http.ResponseWriter, r *http.Request) {
	// get session id and PDA id from header and path variables respectively
	sessionId, pdaId, err := parseSessionIdAndPdaId(r)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	// get position from the path variables
	position, err := strconv.Atoi(parseRequestVariable(r, "position"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "position is not an integer")
		return
	}

	err = pdaService.presentEOS(sessionId, pdaId, position)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	respondWithJSON(w, http.StatusOK, map[string]bool{"eos_declared": true})
}

func snapshot(w http.ResponseWriter, r *http.Request) {
	sessionId, pdaId, err := parseSessionIdAndPdaId(r)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	k, err1 := strconv.Atoi(parseRequestVariable(r, "k"))
	if err1 != nil {
		respondWithError(w, http.StatusBadRequest, "k is not an integer")
		return
	}

	snapshot, err2 := pdaService.snapshot(sessionId, pdaId, k)
	if err2 != nil {
		respondWithError(w, http.StatusBadRequest, err2.Error())
		return
	}
	respondWithJSON(w, http.StatusOK, snapshot)
}

func getC3State(w http.ResponseWriter, r *http.Request) {
	sessionId, pdaId, err := parseSessionIdAndPdaId(r)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	session, err2 := pdaService.getSessionPDA(sessionId, pdaId)
	if err2 != nil {
		respondWithError(w, http.StatusBadRequest, err2.Error())
		return
	}
	respondWithJSON(w, http.StatusOK, session)
}

func createSession(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	pdaId, err := strconv.Atoi(params["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "ID is not an integer")
		return
	}

	sessionId, err := pdaService.createSession(pdaId)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	respondWithJSON(w, http.StatusOK, map[string]string{"sessionId": sessionId})
}

func loadPdaIntoAvailablePdas(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	pdaId, err := strconv.Atoi(params["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "ID is not an integer")
		return
	}
	pdaService.loadPdaIntoAvailablePdas(pdaId)
	respondWithJSON(w, http.StatusOK, map[string]string{})
}

func getPDAById(w http.ResponseWriter, r *http.Request) {
	//Accept PDA with ID and specification
	id, err := strconv.Atoi(parseRequestVariable(r, "id"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "ID is not an integer")
		return
	}

	pdaProcessor, err := pdaService.getPDAById(id)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	respondWithJSON(w, http.StatusOK, pdaProcessor)
}

func handleRequests(portNumber string) {
	fmt.Println("----------------------------------------------")
	if len(portNumber) == 0 {
		portNumber = "8801"
		fmt.Println("No port specified, using default port: " + portNumber)
	}
	fmt.Println("Starting PDA Server on port: " + portNumber)
	myRouter := mux.NewRouter().StrictSlash(true)
	myRouter.HandleFunc("/pdas", pdaList).Methods("GET")
	myRouter.HandleFunc("/pdas/{id}", createPDA).Methods("PUT")
	myRouter.HandleFunc("/pdas/{id}/reset", resetPDA).Methods("PUT")
	myRouter.HandleFunc("/pdas/{id}/{token}/{position}", presentToken).Methods("PUT")
	myRouter.HandleFunc("/pdas/{id}/eos/{position}", presentEOS).Methods("POST") // In requirement document this is PUT but it is conflicting with  "/pdas/{id}/{token}/{position}" so changed it to POST
	myRouter.HandleFunc("/pdas/{id}/is_accepted", isAccepted).Methods("GET")
	myRouter.HandleFunc("/pdas/{id}/stack/top/{k}", peek).Methods("GET")
	myRouter.HandleFunc("/pdas/{id}/stack/len", stackLength).Methods("GET")
	myRouter.HandleFunc("/pdas/{id}/state", currentState).Methods("GET")
	myRouter.HandleFunc("/pdas/{id}/close", closePDA).Methods("GET")
	myRouter.HandleFunc("/pdas/{id}/delete", deletePDA).Methods("DELETE")
	myRouter.HandleFunc("/pdas/{id}/tokens", queuedTokens).Methods("GET")
	myRouter.HandleFunc("/pdas/{id}/snapshot/{k}", snapshot).Methods("GET")
	myRouter.HandleFunc("/pdas/{id}/join", addPDAToReplicaGroup).Methods("PUT")
	myRouter.HandleFunc("/pdas/{id}/code", getPDAById).Methods("GET")
	myRouter.HandleFunc("/pdas/{id}/c3state", getC3State).Methods("GET")

	// additional utilities apis
	myRouter.HandleFunc("/pdas/{id}/createSession", createSession).Methods("GET")
	myRouter.HandleFunc("/pdas/{id}/load", loadPdaIntoAvailablePdas).Methods("GET")

	// replica group APIs
	myRouter.HandleFunc("/replica_pdas", getAvailableReplicaGroups).Methods("GET")
	myRouter.HandleFunc("/replica_pdas/{gid}", createReplicaGroup).Methods("PUT")
	myRouter.HandleFunc("/replica_pdas/{gid}/reset", resetReplicaGroup).Methods("PUT")
	myRouter.HandleFunc("/replica_pdas/{gid}/members", getMembersFromReplicaGroup).Methods("GET")
	myRouter.HandleFunc("/replica_pdas/{gid}/connect", connectToReplicaGroup).Methods("GET")
	myRouter.HandleFunc("/replica_pdas/{gid}/close", closeReplicaGroup).Methods("PUT")
	myRouter.HandleFunc("/replica_pdas/{gid}/delete", deleteReplicaGroup).Methods("DELETE")

	// TODO To be implemented
	// TODO myRouter.HandleFunc("/replica_pdas/{id}/code", getPDASpecs).Methods("GET")

	fmt.Println("Successfully started PDA Server on port: " + portNumber)
	fmt.Printf("Base URL: http://localhost:%s/\n", portNumber)
	fmt.Println("----------------------------------------------")
	log.Fatal(http.ListenAndServe(":"+portNumber,
		handlers.CORS(
			handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization", "session-id"}),
			handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "HEAD", "OPTIONS"}),
			handlers.AllowedOrigins([]string{"*"}))(myRouter)))
}

func parseSessionIdAndPdaId(r *http.Request) (string, int, error) {
	sessionId := r.Header.Get("session-id")
	if len(sessionId) == 0 {
		return sessionId, 0, errors.New("missing session id")
	}

	pdaId, err := strconv.Atoi(parseRequestVariable(r, "id"))
	if err != nil {
		return sessionId, pdaId, errors.New("id is not an integer")
	}
	if pdaId < 0 {
		return sessionId, pdaId, errors.New("id should be a positive integer")
	}

	return sessionId, pdaId, nil
}
func parseRequestVariable(r *http.Request, paramKey string) string {
	params := mux.Vars(r)
	return params[paramKey]
}
func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}
func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

var port string

func main() {
	if len(port) > 0 {
		if _, err := strconv.Atoi(port); err != nil {
			fmt.Print("Invalid Port: Port number should be integers\n\n")
			os.Exit(1)
		} else if len(port) < 4 {
			fmt.Print("Invalid Port: Port number should be at least 4 digit\n\n")
			os.Exit(1)
		}
	}

	// init service
	pdaService = &PDAService{}

	pdaService.initService()

	// register handler to router
	handleRequests(port)
}
