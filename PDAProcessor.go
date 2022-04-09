package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
)

type PDAProcessor struct {
	ID                        int        `json:"ID"`
	Name                      string     `json:"name"`
	States                    []string   `json:"states"`
	InputAlphabet             []string   `json:"input_alphabet"`
	StackAlphabet             []string   `json:"stack_alphabet"`
	AcceptingStates           []string   `json:"accepting_states"`
	StartState                string     `json:"start_state"`
	Transitions               [][]string `json:"transitions"`
	Eos                       string     `json:"eos"`
	Stack                     []string   `json:"-"`
	CurrentState              string     `json:"-"`
	TransitionsTaken          []string   `json:"-"`
	CurrentStackTop           string     `json:"-"`
	PdaClock                  int        `json:"-"`
	PendingTokenQueue         []string   `json:"-"`
	LastConsumedPosition      int        `json:"-"`
	PDAFailedInLastEvaluation bool       `json:"-"`
	EOSPresentedAtPosition    int        `json:"-"`
}

/**
load and parse/process spec as the JSON specification string of a PDA. Return True on success.
*/
func (pdaProcessor *PDAProcessor) open(specFilePath string) (bool, error) {
	fmt.Println("\n***************** Open PDA ************************")
	pdaProcessor.PdaClock = 1
	fmt.Println("Opening specs file: " + specFilePath)

	file, err := ioutil.ReadFile(specFilePath)
	pdaProcessor.PdaClock++

	if err != nil {
		return false, fmt.Errorf("couldn't open specification file: "+specFilePath, err)
	}

	err = json.Unmarshal(file, &pdaProcessor)

	if err != nil {
		return false, fmt.Errorf("couldn't unmashal specification from file to PDAProcessor object, %v", err)
	}

	// initialize all the struct parameters
	pdaProcessor.reset(false)
	pdaProcessor.PdaClock++

	fmt.Println("Successfully loaded PDA, file: " + specFilePath)

	return true, nil
}

/**
initialize the PDA to be at its start state with an empty stack.
*/
func (pdaProcessor *PDAProcessor) reset(isReset bool) {
	if isReset {
		fmt.Println("\n***************** Reset PDA", pdaProcessor.ID, "************************")
	}
	pdaProcessor.CurrentState = pdaProcessor.StartState
	pdaProcessor.CurrentStackTop = ""
	pdaProcessor.Stack = []string{}
	pdaProcessor.TransitionsTaken = []string{}
	pdaProcessor.PendingTokenQueue = make([]string, PDA_PENDING_QUEUE_LENGTH)
	pdaProcessor.EOSPresentedAtPosition = -1
	pdaProcessor.LastConsumedPosition = -1
	pdaProcessor.PDAFailedInLastEvaluation = false

	// add current state as first state in transition taken
	addTransitionIfRequired(pdaProcessor, pdaProcessor.CurrentState)

	if isReset {
		fmt.Println("PDA Reset complete. fields reset, \n" +
			" - current state\n" +
			" - stack\n" +
			" - pending token queue\n" +
			" - transitions taken\n" +
			" - last evaluation status")
	}
}

/**
present token as the current input token to the PDA.The PDA consumes the token,
takes appropriate transition(s), and returns the #transitions taken due to this put() call.
*/
func (pdaProcessor *PDAProcessor) evaluateInput(tokenStream string) ([]string, error) {
	if tokenStream == "" {
		// on empty input DO NOT do anything as this, let start state be final state
	} else {
		// append end of string to input token

		pdaProcessor.put(0, "")
		errorInEvaluation := false
		inputToken := strings.Fields(tokenStream)
		for index, c := range inputToken {
			err := validateInput(pdaProcessor, c)
			if err != nil {
				return nil, err
			}
			transitionTaken := pdaProcessor.put(index+1, c)

			if len(transitionTaken) == 0 {
				fmt.Printf("PDA failed to make transition for input %q at index: %d \n", c, index+1)
				errorInEvaluation = true
				break
			} else {
				addTransitionIfRequired(pdaProcessor, transitionTaken)
				fmt.Println("Current PDA Clock tick value is:", pdaProcessor.PdaClock)
			}

		}

		if !errorInEvaluation && len(pdaProcessor.Stack) == 1 {
			if pdaProcessor.eos() {
				fmt.Println("End of input has reached!")
			}
			transitionTaken := pdaProcessor.put(len(tokenStream), "")
			addTransitionIfRequired(pdaProcessor, transitionTaken)
		}
	}

	return pdaProcessor.TransitionsTaken, nil
}

/**
announce the end of input token-stream to the PDA.
*/
func (pdaProcessor *PDAProcessor) eos() bool {
	if pdaProcessor.Stack[len(pdaProcessor.Stack)-1] == "$" {
		return true
	}
	return false
}

/**
return True if PDA is currently at an accepting state with empty stack; False otherwise.
*/
func (pdaProcessor *PDAProcessor) is_accepted() bool {
	fmt.Println("\n***************** Is Accepted by PDA", pdaProcessor.ID, "************************")
	fmt.Println("Current state:", pdaProcessor.CurrentState)
	// check if current state exists in accepting states array
	found := findInArray(pdaProcessor.AcceptingStates, pdaProcessor.CurrentState)
	pdaProcessor.PdaClock++
	stackLength := 0
	for _, a := range pdaProcessor.Stack {
		if a != pdaProcessor.Eos {
			stackLength++
		}
	}

	return found && stackLength == 0
}

/**
return up to k stack tokens from the top of the stack (default k = 1) without modifying the stack.
*/
func (pdaProcessor *PDAProcessor) peek(k int) []string {
	fmt.Println("\n***************** Peek Stack States from PDA", pdaProcessor.ID, "************************")
	fmt.Printf("peeking top %+v states from current stack: %+v \n", k, pdaProcessor.Stack)
	if k <= 0 {
		k = 1
	}
	pdaProcessor.PdaClock++
	if len(pdaProcessor.Stack) <= k {
		return pdaProcessor.Stack
	}

	return pdaProcessor.Stack[len(pdaProcessor.Stack)-k:]
}

/**
return the current state of the PDA’s control.
*/
func (pdaProcessor *PDAProcessor) current_state() string {
	fmt.Println("\n***************** Get Current State PDA", pdaProcessor.ID, "************************")
	fmt.Printf("Current State: %q\n", pdaProcessor.CurrentState)

	return pdaProcessor.CurrentState
}

/**
garbage-collect/return any (re-usable) resources used by the PDA.
*/
func (pdaProcessor *PDAProcessor) close() {
	fmt.Println("\n****************** Close PDA", pdaProcessor.ID, "************************")
}

func (pdaProcessor *PDAProcessor) pushToQueue(position int, token string) (string, error) {
	fmt.Println("\n***************** New Token Presented to PDA", pdaProcessor.ID, "************************")
	if pdaProcessor.PDAFailedInLastEvaluation {
		return "", errors.New("PDA Evaluation failed while consuming last valid token, Please reset PDA before using it")
	}
	if pdaProcessor.EOSPresentedAtPosition != -1 && pdaProcessor.EOSPresentedAtPosition < position {
		return "", errors.New("EOS is already presented before this position so PDA can not accept more tokens")
	}
	if position < pdaProcessor.LastConsumedPosition {
		return "", errors.New("PDA already consumed all the tokens up to position " + strconv.Itoa(pdaProcessor.LastConsumedPosition-1))
	}

	// if first token is presented to tFhe pda then make initial transition
	if position == 0 && pdaProcessor.LastConsumedPosition == -1 {
		fmt.Println("Initializing empty state")
		pdaProcessor.put(0, "")
	}

	var transitionTaken = ""
	var err error = nil
	fmt.Printf("New token %q presented at position %d\n", token, position)
	if position == 0 || position == pdaProcessor.LastConsumedPosition {
		transitionTaken, err = consumeToken(pdaProcessor, position, token, false)
		if err != nil {
			return transitionTaken, err
		}
	} else {
		if len(pdaProcessor.PendingTokenQueue[position]) == 0 {
			fmt.Printf("Token %q is added to pending queue at position %d\n", token, position)

			// push to pending queue at given position
			pdaProcessor.PendingTokenQueue[position] = token
		} else {
			fmt.Printf("Token already existing for this position in pending queue\n")
		}
		printLog(pdaProcessor)
	}

	return transitionTaken, nil
}

func consumeToken(pdaProcessor *PDAProcessor, position int, token string, processingPendingQueue bool) (string, error) {
	var transitionTaken = ""

	// consume token directly
	fmt.Printf("Consuming token %q at position %d\n", token, position)
	transitionTaken = pdaProcessor.put(position+1, token)

	if len(transitionTaken) == 0 {
		pdaProcessor.PDAFailedInLastEvaluation = true
		return "", errors.New(fmt.Sprintf("PDA failed to make transition for input %q at position: %d", token, position))
	} else {
		addTransitionIfRequired(pdaProcessor, transitionTaken)

		// process if current toke was at EOS position
		if pdaProcessor.EOSPresentedAtPosition != -1 && pdaProcessor.EOSPresentedAtPosition == position {
			// reached EOS
			transitionTaken, err := reachedEOS(pdaProcessor)
			if err != nil {
				return transitionTaken, err
			}
		} else if !processingPendingQueue {
			// as this call is for actual token presented from user and it is accepted as well,
			// process subsequent items from pending queue if any
			fmt.Println("Processing subsequent tokens from pending queue")
			processedAtLeastOne := false
			for index, t := range pdaProcessor.PendingTokenQueue {
				if index > position {
					if len(t) != 0 {
						_, err := consumeToken(pdaProcessor, index, t, true)
						if err != nil {
							return transitionTaken, err
						}
						// clear token consumed
						pdaProcessor.PendingTokenQueue[index] = ""
						processedAtLeastOne = true
					} else {
						break
					}
				}
			}

			if !processedAtLeastOne {
				fmt.Println("No subsequent tokens to consume")
			}
		}
	}

	return transitionTaken, nil
}

func reachedEOS(pdaProcessor *PDAProcessor) (string, error) {
	if !pdaProcessor.eos() {
		pdaProcessor.PDAFailedInLastEvaluation = true
		return "", errors.New(fmt.Sprintf("PDA reached presented EOS at position %d but stack is not empty so PDA failed to make final transition\n", pdaProcessor.EOSPresentedAtPosition))
	}

	// make final pop on eos
	transitionTaken := pdaProcessor.put(pdaProcessor.LastConsumedPosition+1, "")
	if len(transitionTaken) == 0 {
		pdaProcessor.PDAFailedInLastEvaluation = true
		return "", errors.New("PDA failed to make final transition on EOS")
	} else {
		addTransitionIfRequired(pdaProcessor, transitionTaken)
		return transitionTaken, nil
	}
}

func printLog(pdaProcessor *PDAProcessor) {
	fmt.Println("--- Status ---")
	fmt.Printf("Current State: %q\n", pdaProcessor.CurrentState)
	fmt.Printf("Stack: %+v\n", pdaProcessor.Stack)
	if pdaProcessor.LastConsumedPosition == -1 {
		fmt.Println("Last consumed position: no token consumed yet")
	} else {
		fmt.Printf("Last consumed position: %d\n", pdaProcessor.LastConsumedPosition-1)
	}
	fmt.Printf("Pending Queue: %+v\n", truncateEmptyTokens(pdaProcessor.PendingTokenQueue))
	fmt.Printf("Transitions taken so far: %+v\n", pdaProcessor.TransitionsTaken)
	fmt.Println("-------------------------")
	fmt.Println()
}

func (pdaProcessor *PDAProcessor) put(position int, token string) string {
	if len(pdaProcessor.Stack) >= 1 {
		pdaProcessor.CurrentStackTop = pdaProcessor.Stack[len(pdaProcessor.Stack)-1]
		pdaProcessor.PdaClock++
	}

	var transitionTaken string
	for _, transition := range pdaProcessor.Transitions {
		// transition is PDA transition array defined as [current_state, current_input, current_stack_top, next_state, to_be_stack_top]
		state := transition[0]
		input := transition[1]
		stackTop := transition[2]
		nextState := transition[3]
		elementToPush := transition[4]

		fmt.Printf("Evaluating: %q, %q, %q ----> %q, %q, %q, %q, %q, stack: %+v\n", pdaProcessor.CurrentState, token, pdaProcessor.CurrentStackTop, state, input, stackTop, nextState, elementToPush, pdaProcessor.Stack)

		if state == pdaProcessor.CurrentState {
			if input == token {
				if pdaProcessor.CurrentStackTop == stackTop {
					pushOrPopIfRequired(pdaProcessor, elementToPush, nextState, pdaProcessor.CurrentStackTop)
					pdaProcessor.PdaClock++
					pdaProcessor.CurrentState = nextState
					pdaProcessor.LastConsumedPosition = position
					transitionTaken = nextState
					break
				} else if stackTop == "" {
					pushOrPopIfRequired(pdaProcessor, elementToPush, nextState, pdaProcessor.CurrentStackTop)
					pdaProcessor.PdaClock++
					pdaProcessor.CurrentState = nextState
					pdaProcessor.LastConsumedPosition = position
					transitionTaken = nextState
					break
				}
			} else if token == "" && pdaProcessor.CurrentStackTop == pdaProcessor.Eos && stackTop == pdaProcessor.Eos {
				pushOrPopIfRequired(pdaProcessor, elementToPush, nextState, pdaProcessor.CurrentStackTop)
				pdaProcessor.PdaClock++
				pdaProcessor.CurrentState = nextState
				pdaProcessor.LastConsumedPosition = position
				transitionTaken = nextState
				break
			}
		}
	}
	printLog(pdaProcessor)
	return transitionTaken
}

func (pdaProcessor *PDAProcessor) presentEOS(position int) error {
	fmt.Println("\n***************** Present EOS ************************")
	if pdaProcessor.PDAFailedInLastEvaluation {
		return errors.New("PDA Evaluation failed while consuming last valid token, Please reset PDA before using it")
	}

	printLog(pdaProcessor)
	if position < pdaProcessor.LastConsumedPosition-1 {
		return errors.New("PDA already consumed all the tokens up to position this position")
	}
	// set token present
	pdaProcessor.EOSPresentedAtPosition = position

	if position == pdaProcessor.LastConsumedPosition-1 {
		// reached EOS
		_, err := reachedEOS(pdaProcessor)
		if err != nil {
			return err
		}
	}

	return nil
}

func (pdaProcessor *PDAProcessor) queued_tokens() []string {
	fmt.Println("\n***************** Queued Token in PDA", pdaProcessor.ID, "************************")
	queue := truncateEmptyTokens(pdaProcessor.PendingTokenQueue)
	fmt.Printf("Queue: %+v\n", queue)

	return queue
}

func truncateEmptyTokens(tokensArray []string) []string {
	var a []string
	// find first non empty token index
	for _, t := range tokensArray {
		if len(t) != 0 {
			a = append(a, t)
		}
	}

	return a
}

func validateInput(pdaProcessor *PDAProcessor, token string) error {
	// validate if all token characters are valid input alphabets
	if len(token) > 0 {
		found := findInArray(pdaProcessor.InputAlphabet, token)
		pdaProcessor.PdaClock++

		if !found {
			return fmt.Errorf("Input token contains unsupported character/string %q", token)
		}
	}
	return nil
}

func pushToStack(pdaProcessor *PDAProcessor, token string) {
	pdaProcessor.Stack = append(pdaProcessor.Stack, token)
	pdaProcessor.PdaClock++
}

func popFromStack(pdaProcessor *PDAProcessor) string {
	var s string
	if len(pdaProcessor.Stack) >= 1 {
		s = pdaProcessor.Stack[len(pdaProcessor.Stack)-1]
		pdaProcessor.Stack = pdaProcessor.Stack[:len(pdaProcessor.Stack)-1]
		pdaProcessor.PdaClock++
	}
	return s
}

func addTransitionIfRequired(pdaProcessor *PDAProcessor, state string) {
	if len(state) > 0 {
		pdaProcessor.TransitionsTaken = append(pdaProcessor.TransitionsTaken, state)
		pdaProcessor.PdaClock++
	}
}

func pushOrPopIfRequired(pdaProcessor *PDAProcessor, elementToPush string, nextState string, currentStackTop string) bool {
	if elementToPush == "" {
		fmt.Printf("Found valid transition, %q => %q, popping: %q\n", pdaProcessor.CurrentState, nextState, currentStackTop)
		popFromStack(pdaProcessor)
		return true
	} else {
		fmt.Printf("Found valid transition, %q => %q, pushing: %q\n", pdaProcessor.CurrentState, nextState, elementToPush)
		pushToStack(pdaProcessor, elementToPush)
		return true
	}
}

func findInArray(arr []string, value string) bool {
	for _, a := range arr {
		if strings.Compare(a, value) == 0 {
			return true
		}
	}
	return false
}
