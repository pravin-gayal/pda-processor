package main

import (
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
)

type PDAReplicaService struct {
}

type ReplicaGroup struct {
	Gid              int          `json:"gid"`
	GroupName        string       `json:"group_name"`
	PdaGroupMembers  []string     `json:"pda_members"`
	PdaCode          int          `json:"pda_code"`
	PdaSpecification PDAProcessor `json:"pda_specification"`
}

var availableReplicaGroups []ReplicaGroup

func (replicaService *PDAReplicaService) getAvailableReplicaGroups() []ReplicaGroup {
	return availableReplicaGroups
}

func (replicaService *PDAReplicaService) createReplicaGroup(gid int, replicaGroup ReplicaGroup) (ReplicaGroup, error) {
	fmt.Println("Creating new Replica Group with id: ", gid)
	if len(replicaGroup.PdaGroupMembers) == 0 {
		return replicaGroup, errors.New("Group members cannot be empty")
	}
	if len(replicaGroup.GroupName) == 0 {
		return replicaGroup, errors.New("Group name cannot be empty")
	}

	existingRG := findReplicaGroupById(gid)
	if existingRG.Gid != 0 {
		return replicaGroup, errors.New("gid already used")
	}

	// save specification on base server
	fmt.Println("Creating new PDA")
	_, _ = pdaService.createNewPDA(replicaGroup.PdaCode, replicaGroup.PdaSpecification)
	fmt.Println("Created new PDA")
	// call pda members to load saved PDA
	for _, memberUrl := range replicaGroup.PdaGroupMembers {
		makeHttpGetCall(memberUrl + "/pdas/" + strconv.Itoa(replicaGroup.PdaCode) + "/load")
	}
	availableReplicaGroups = append(availableReplicaGroups, replicaGroup)
	fmt.Println("Successfully created replica group with id: ", gid)
	return replicaGroup, nil
}

func (replicaService *PDAReplicaService) resetReplicaGroup(gid int) error {
	fmt.Println("Resetting Replica Group")
	existingRG := findReplicaGroupById(gid)
	if existingRG.Gid == 0 {
		return errors.New("invalid gid")
	}
	// call pda members to load saved PDA
	for _, memberUrl := range existingRG.PdaGroupMembers {
		makeHttpPutCall(memberUrl + "/pdas/" + strconv.Itoa(existingRG.PdaCode) + "/reset")
	}
	fmt.Println("Successfully reset Replica Group")
	return nil
}

func (replicaService *PDAReplicaService) getMembersFromReplicaGroup(gid int) ([]string, error) {
	fmt.Println("Getting members from a Replica Group")
	existingRG := findReplicaGroupById(gid)
	if existingRG.Gid == 0 {
		return existingRG.PdaGroupMembers, errors.New("invalid group id")
	}
	return existingRG.PdaGroupMembers, nil
}

func (replicaService *PDAReplicaService) connectToReplicaGroup(gid int) (string, error) {
	fmt.Println("Connecting to a Replica Group")
	existingRG := findReplicaGroupById(gid)
	if existingRG.Gid == 0 {
		return "", errors.New("invalid group id")
	}
	randomMemberIndex := rand.Intn(len(existingRG.PdaGroupMembers))
	return existingRG.PdaGroupMembers[randomMemberIndex], nil
}

func (replicaService *PDAReplicaService) deleteReplicaGroup(gid int) error {
	fmt.Println("Deleting Replica Group")
	index := -1
	for i, replicaGroup := range availableReplicaGroups {
		if replicaGroup.Gid == gid {
			index = i
			break
		}
	}
	if index == -1 {
		fmt.Println("Replica group", gid, "does not exist")
		return errors.New("Replica group with specified id does not exist")
	}

	availableReplicaGroups = append(availableReplicaGroups[:index], availableReplicaGroups[index+1:]...)
	return nil
}

func (replicaService *PDAReplicaService) closeReplicaGroup(gid int) error {
	fmt.Println("Closing Replica Group")
	existingRG := findReplicaGroupById(gid)
	if existingRG.Gid == 0 {
		return errors.New("invalid gid")
	}
	// call pda members to load saved PDA
	for _, memberUrl := range existingRG.PdaGroupMembers {
		makeHttpPutCall(memberUrl + "/pdas/" + strconv.Itoa(existingRG.PdaCode) + "/close")
	}
	fmt.Println("Successfully closed Replica Group")
	return nil
}

func (replicaService *PDAReplicaService) addPDAToReplicaGroup(pdaId int, replicaGroup ReplicaGroup) error {
	existingRG := findReplicaGroupById(replicaGroup.Gid)
	if existingRG.Gid == 0 {
		return errors.New("invalid gid")
	}



	return nil
}

func findReplicaGroupById(gid int) ReplicaGroup {
	for _, replicaGroup := range availableReplicaGroups {
		if replicaGroup.Gid == gid {
			return replicaGroup
		}
	}
	// return empty struct
	return ReplicaGroup{}
}

func makeHttpGetCall(url string) {
	fmt.Println("Making call to", url)
	resp, err := http.Get(url)

	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
}

func makeHttpPutCall(url string) {
	fmt.Println("Making call to", url)
	resp, err := http.Get(url)

	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
}
