package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
)

type PDAService struct {
}

var availablePDAs []PDAProcessor
var sessionMap map[string]*PDAProcessor

func (pdaService *PDAService) getAllAvailablePDAs() []PDAProcessor {
	return availablePDAs
}

func (pdaService *PDAService) createNewPDA(id int, pdaProcessor PDAProcessor) (PDAProcessor, error) {
	log.Println("Creating new PDA with id: ", id)
	pdaProcessor.ID = id
	isValid, err := validatePDADetails(id, pdaProcessor)
	if !isValid && err != nil {
		return pdaProcessor, err
	}

	//create new json file for newly incoming specification
	isCreated, err := createJsonFile(id, pdaProcessor)
	if !isCreated && err != nil {
		return pdaProcessor, err
	}

	//goroutine to add to available pdas
	go addToAvailablePDA(pdaProcessor)

	log.Println("Successfully created PDA with id: ", id)
	// return created PDA object
	return pdaProcessor, nil
}

func (pdaService *PDAService) initService() {
	// initialize session map
	sessionMap = make(map[string]*PDAProcessor)

	// read existing PDAs from the file system and load them in available PDA list
	pdaService.loadExistingPDAs()
}

func (pdaService *PDAService) loadExistingPDAs() {
	files, err := ioutil.ReadDir(PDA_FILES_BASE_FOLDER)
	if err != nil {
		log.Fatal(err)
	}
	for _, f := range files {
		pdaProcessor := openFileByName(f.Name())
		if pdaProcessor != nil {
			availablePDAs = append(availablePDAs, *pdaProcessor)
		}
	}
}

func (pdaService *PDAService) createSession(pdaId int) (string, error) {
	index := pdaIndexInAvailablePDAsById(pdaId)
	if index == -1 {
		log.Println("PDA", pdaId, "does not exist")
		return "", errors.New("PDA with specified id does not exist")
	}

	sessionId := "session_" + strconv.Itoa(rand.Intn(100))
	_, containsKey := sessionMap[sessionId]
	for containsKey {
		sessionId = "session_" + strconv.Itoa(rand.Intn(100))
		_, containsKey = sessionMap[sessionId]
	}

	filename := PDA_FILE_NAME_PREFIX + strconv.Itoa(pdaId) + PDA_FILE_NAME_POSTFIX
	pdaProcessor := openFileByName(filename)
	if pdaProcessor != nil {
		sessionMap[sessionId] = pdaProcessor
		return sessionId, nil
	}
	return "", errors.New("couldn't create session")
}

/**
Method to reset PDA in given session id
*/
func (pdaService *PDAService) resetPDA(sessionId string, pdaId int) error {
	// get pda for session id
	pdaProcessor, hasSession := sessionMap[sessionId]
	if !hasSession {
		return errors.New("invalid session")
	}

	if pdaId != pdaProcessor.ID {
		return errors.New("invalid pda id for given session")
	}

	// reset pda
	pdaProcessor.reset(true)
	return nil
}

/**
Method to check if PDA is in accepted state
*/
func (pdaService *PDAService) isAccepted(sessionId string, pdaId int) (bool, error) {
	// get pda for session id
	pdaProcessor, hasSession := sessionMap[sessionId]
	if !hasSession {
		return false, errors.New("invalid session")
	}

	if pdaId != pdaProcessor.ID {
		return false, errors.New("invalid pda id for given session")
	}

	// return is accepted boolean
	isAccepted := pdaProcessor.is_accepted()
	return isAccepted, nil
}

/**
Method to peek "k" states from the stack
*/
func (pdaService *PDAService) peek(sessionId string, pdaId int, k int) ([]string, error) {
	// get pda for session id
	pdaProcessor, hasSession := sessionMap[sessionId]
	if !hasSession {
		return nil, errors.New("invalid session")
	}

	if pdaId != pdaProcessor.ID {
		return nil, errors.New("invalid pda id for given session")
	}

	//  return array of states
	states := pdaProcessor.peek(k)

	// return empty list if stack is null
	if states == nil {
		states = make([]string, 0)
	}

	return states, nil
}

func (pdaService *PDAService) stackLength(sessionId string, pdaId int) (int, error) {
	// get pda for session id
	pdaProcessor, hasSession := sessionMap[sessionId]
	if !hasSession {
		return 0, errors.New("invalid session")
	}

	if pdaId != pdaProcessor.ID {
		return 0, errors.New("invalid pda id for given session")
	}

	length := 0
	for _, s := range pdaProcessor.Stack {
		if s != pdaProcessor.Eos {
			length = length + 1
		}
	}

	// return stack length
	return length, nil
}

func (pdaService *PDAService) currentState(sessionId string, pdaId int) (string, error) {
	// get pda for session id
	pdaProcessor, hasSession := sessionMap[sessionId]
	if !hasSession {
		return "", errors.New("invalid session")
	}

	if pdaId != pdaProcessor.ID {
		return "", errors.New("invalid pda id for given session")
	}

	// return current state
	return pdaProcessor.CurrentState, nil
}

func (pdaService *PDAService) closePDA(sessionId string, pdaId int) error {
	// get pda for session id
	pdaProcessor, hasSession := sessionMap[sessionId]
	if !hasSession {
		return errors.New("invalid session")
	}

	if pdaId != pdaProcessor.ID {
		return errors.New("invalid pda id for given session")
	}

	// call close
	pdaProcessor.close()
	return nil
}

func (pdaService *PDAService) deletePDA(pdaId int) error {
	log.Println("\n***************** Delete PDA", pdaId, "************************")
	index := pdaIndexInAvailablePDAsById(pdaId)
	if index == -1 {
		log.Println("PDA", pdaId, "does not exist")
		return errors.New("PDA with specified id does not exist")
	}

	// invalidate all the sessions using this PDA
	log.Println("removing all the sessions with specified PDA")
	count := 0
	for key, value := range sessionMap {
		if value.ID == pdaId {
			delete(sessionMap, key)
			count++
		}
	}
	log.Println("removed", count, "sessions using specified PDA")

	// remove PDA processor object from available PDAs
	log.Println("removing PDA from available index list")
	availablePDAs = append(availablePDAs[:index], availablePDAs[index+1:]...)

	// delete PDA spec file from file system
	log.Println("removing PDA specification file from permanent storage")
	filename := PDA_FILE_NAME_PREFIX + strconv.Itoa(pdaId) + PDA_FILE_NAME_POSTFIX

	err := os.Remove(filepath.Join(PDA_FILES_BASE_FOLDER, filepath.Base(filename)))
	if err != nil {
		log.Println("failed to delete file from the file system")
		return errors.New("failed to delete file from the file system")
	}
	log.Println("PDA specification file is removed")

	return nil
}

func (pdaService *PDAService) presentToken(sessionId string, pdaId int, token string, position int) (bool, error) {
	// get pda for session id
	pdaProcessor, hasSession := sessionMap[sessionId]
	if !hasSession {
		return false, errors.New("invalid session")
	}

	if pdaId != pdaProcessor.ID {
		return false, errors.New("invalid pda id for given session")
	}

	// check if token is valid input alphabet
	found := false
	for _, inputAlphabet := range pdaProcessor.InputAlphabet {
		if inputAlphabet == token {
			found = true
			break
		}
	}
	if !found {
		return false, errors.New("input token is not supported by PDA")
	}

	// present token to PDA
	transitionTaken, err := pdaProcessor.pushToQueue(position, token)
	if err != nil {
		log.Println(err.Error())
		return false, err
	}

	return len(transitionTaken) != 0, nil
}

func (pdaService *PDAService) getPendingQueue(sessionId string, pdaId int) ([]string, error) {
	// get pda for session id
	pdaProcessor, hasSession := sessionMap[sessionId]
	if !hasSession {
		return nil, errors.New("invalid session")
	}

	if pdaId != pdaProcessor.ID {
		return nil, errors.New("invalid pda id for given session")
	}

	// return stack length
	return pdaProcessor.queued_tokens(), nil
}

func (pdaService *PDAService) getSessionPDA(sessionId string, pdaId int) (interface{}, error) {
	// get pda for session id
	pdaProcessor, hasSession := sessionMap[sessionId]
	if !hasSession {
		return pdaProcessor, errors.New("invalid session")
	}
	if pdaId != pdaProcessor.ID {
		return pdaProcessor, errors.New("invalid pda id for given session")
	}

	m := make(map[string]interface{})

	m["session_id"] = sessionId
	m["pda_id"] = pdaProcessor.ID
	m["pda_name"] = pdaProcessor.Name
	m["pda_stack"] = pdaProcessor.Stack

	return m, nil
}

func (pdaService *PDAService) presentEOS(sessionId string, pdaId int, position int) error {
	// get pda for session id
	pdaProcessor, hasSession := sessionMap[sessionId]
	if !hasSession {
		return errors.New("invalid session")
	}

	if pdaId != pdaProcessor.ID {
		return errors.New("invalid pda id for given session")
	}

	// present token to PDA
	err := pdaProcessor.presentEOS(position)
	if err != nil {
		log.Println(err.Error())
		return err
	}
	return nil
}

type result struct {
	CurrentState string   `json:"current_state"`
	Peek         []string `json:"peek"`
	QueuedTokens []string `json:"queued_tokens"`
}

func (pdaService *PDAService) snapshot(sessionId string, pdaId int, k int) (result, error) {
	result := result{}
	// get pda for session id
	pdaProcessor, hasSession := sessionMap[sessionId]
	if !hasSession {
		return result, errors.New("invalid session")
	}

	if pdaId != pdaProcessor.ID {
		return result, errors.New("invalid pda id for given session")
	}

	peek := pdaProcessor.peek(k)
	if peek == nil {
		peek = make([]string, 0)
	}

	queue := pdaProcessor.queued_tokens()
	if queue == nil {
		queue = make([]string, 0)
	}

	result.CurrentState = pdaProcessor.current_state()
	result.Peek = peek
	result.QueuedTokens = queue
	return result, nil
}

func (pdaService *PDAService) loadPdaIntoAvailablePdas(pdaId int) {
	filename := PDA_FILE_NAME_PREFIX + strconv.Itoa(pdaId) + PDA_FILE_NAME_POSTFIX
	pdaProcessor := openFileByName(filename)
	if pdaProcessor != nil {
		addToAvailablePDA(*pdaProcessor)
	}
}

func (pdaService *PDAService) getPDAById(pdaId int) (PDAProcessor, error) {
	var pdaProcessor PDAProcessor
	for _, pda := range availablePDAs {
		if pda.ID == pdaId {
			pdaProcessor = pda
			break
		}
	}

	if pdaProcessor.ID == 0 {
		return pdaProcessor, errors.New("Invalid pda id")
	}

	return pdaProcessor, nil
}

// ***************************************************************//
// ******************** Private Methods **************************//
// ***************************************************************//
func pdaIndexInAvailablePDAsById(pdaId int) int {
	for i, pda := range availablePDAs {
		if pda.ID == pdaId {
			return i
		}
	}
	return -1
}

func addToAvailablePDA(pdaProcessor PDAProcessor) {
	availablePDAs = append(availablePDAs, pdaProcessor)
}

func validatePDADetails(id int, pdaProcessor PDAProcessor) (bool, error) {
	// validate id
	if id <= 0 {
		return false, errors.New("ID should be a positive integer")
	}
	for _, pda := range availablePDAs {
		if pda.ID == id {
			return false, errors.New("ID already used")
		}
	}

	// validate pda details
	if pdaProcessor.Name == "" {
		return false, errors.New("PDA Name field cannot be empty")
	}
	if pdaProcessor.StartState == "" {
		return false, errors.New("PDA start state cannot be empty")
	}
	if pdaProcessor.Eos == "" {
		return false, errors.New("PDA end of stream field cannot be empty")
	}
	if len(pdaProcessor.States) == 0 {
		return false, errors.New("PDA states cannot be empty")
	}
	if len(pdaProcessor.InputAlphabet) == 0 {
		return false, errors.New("PDA input alphabets field cannot be empty")
	}
	if len(pdaProcessor.StackAlphabet) == 0 {
		return false, errors.New("PDA stack alphabets field cannot be empty")
	}
	if len(pdaProcessor.AcceptingStates) == 0 {
		return false, errors.New("PDA accepting states field cannot be empty")
	}
	return true, nil
}

func createJsonFile(pdaId int, pdaProcessor PDAProcessor) (bool, error) {
	dataBytes, err1 := json.MarshalIndent(pdaProcessor, "", "  ")
	if err1 != nil {
		return false, errors.New("could not marshal to a JSON file")
	}
	filename := PDA_FILE_NAME_PREFIX + strconv.Itoa(pdaId) + PDA_FILE_NAME_POSTFIX
	err3 := ioutil.WriteFile(filepath.Join(PDA_FILES_BASE_FOLDER, filepath.Base(filename)), dataBytes, 0644)
	if err3 != nil {
		log.Println(err3)
		return false, errors.New("could not write a json to the file")
	}
	return true, nil
}

func openFileByName(fileName string) *PDAProcessor {
	filename := filepath.Join(PDA_FILES_BASE_FOLDER, fileName)
	pdaProcessor := &PDAProcessor{}
	opened, err := pdaProcessor.open(filename)
	if err != nil {
		log.Println(err)
		return nil
	} else if !opened {
		log.Println("PDA specification file was not opened")
		return nil
	}
	return pdaProcessor
}
