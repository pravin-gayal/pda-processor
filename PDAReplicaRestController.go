package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
)

var pdaReplicaService *PDAReplicaService

func getAvailableReplicaGroups(w http.ResponseWriter, r *http.Request) {
	array := pdaReplicaService.getAvailableReplicaGroups()
	if array == nil {
		array = []ReplicaGroup{}
	}

	respondWithJSON(w, http.StatusOK, array)
}

func createReplicaGroup(w http.ResponseWriter, r *http.Request) {
	gid, err := parseGID(r)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	// unmarshal body
	var replicaGroup ReplicaGroup
	_ = json.NewDecoder(r.Body).Decode(&replicaGroup)
	replicaGroup.Gid = gid

	// create replica group
	createdRG, err1 := pdaReplicaService.createReplicaGroup(gid, replicaGroup)
	if err1 != nil {
		respondWithError(w, http.StatusBadRequest, err1.Error())
		return
	}

	// return created pda processor
	respondWithJSON(w, http.StatusOK, createdRG)
}

func resetReplicaGroup(w http.ResponseWriter, r *http.Request) {
	gid, err := parseGID(r)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	err1 := pdaReplicaService.resetReplicaGroup(gid)
	if err1 != nil {
		respondWithError(w, http.StatusBadRequest, err1.Error())
		return
	}
	respondWithJSON(w, http.StatusOK, map[string]bool{"reset": true})
}

func getMembersFromReplicaGroup(w http.ResponseWriter, r *http.Request) {
	gid, err := parseGID(r)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	members, err1 := pdaReplicaService.getMembersFromReplicaGroup(gid)
	if err1 != nil {
		respondWithError(w, http.StatusBadRequest, err1.Error())
		return
	}
	respondWithJSON(w, http.StatusOK, members)

}

func connectToReplicaGroup(w http.ResponseWriter, r *http.Request) {
	gid, err := parseGID(r)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	member, err1 := pdaReplicaService.connectToReplicaGroup(gid)
	if err1 != nil {
		respondWithError(w, http.StatusBadRequest, err1.Error())
		return
	}
	respondWithJSON(w, http.StatusOK, map[string]string{"connected_to_pda": member})
}

func deleteReplicaGroup(w http.ResponseWriter, r *http.Request) {
	gid, err := parseGID(r)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	err1 := pdaReplicaService.deleteReplicaGroup(gid)
	if err1 != nil {
		respondWithError(w, http.StatusBadRequest, err1.Error())
		return
	}
	respondWithJSON(w, http.StatusOK, map[string]bool{"deleted": true})

}

func closeReplicaGroup(w http.ResponseWriter, r *http.Request) {
	gid, err := parseGID(r)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	err1 := pdaReplicaService.closeReplicaGroup(gid)
	if err1 != nil {
		respondWithError(w, http.StatusBadRequest, err1.Error())
		return
	}
	respondWithJSON(w, http.StatusOK, map[string]bool{"closed": true})
}

func addPDAToReplicaGroup(w http.ResponseWriter, r *http.Request) {
	pdaId, err := strconv.Atoi(parseRequestVariable(r, "id"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "ID is not an integer")
		return
	}

	var replicaGroup ReplicaGroup
	_ = json.NewDecoder(r.Body).Decode(&replicaGroup)

	err1 := pdaReplicaService.addPDAToReplicaGroup(pdaId, replicaGroup)
	if err1 != nil {
		respondWithError(w, http.StatusBadRequest, err1.Error())
		return
	}

	// return created pda processor
	respondWithJSON(w, http.StatusOK, map[string]bool{"added": true})
}

func parseGID(r *http.Request) (int, error) {
	gid, err := strconv.Atoi(parseRequestVariable(r, "gid"))
	if err != nil {
		return gid, errors.New("GID is not an integer")
	}
	if gid < 0 {
		return gid, errors.New("GID should be a positive integer")
	}

	return gid, nil
}
