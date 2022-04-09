package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

func main1() {

	// get first argument as PDA specification file path
	if len(os.Args) <= 1 {
		log.Fatal("PDA specification file path is missing")
	}
	// get first argument as file path
	pdaSpecsFilePath := os.Args[1]

	// load PDA with specification from file on given paths
	pdaProcessor := &PDAProcessor{}
	opened, err := pdaProcessor.open(pdaSpecsFilePath)
	if err != nil {
		log.Fatal(err)
	} else if !opened {
		log.Fatal("PDA specification file was not opened")
	}

	// initialize input token stream
	inputString := ""

	// check for optional input token stream
	if len(os.Args) == 3 {
		inputFilePath := os.Args[2]
		// read input stream from the file
		str, _ := readInputFromFile(inputFilePath)
		inputString = str
	} else {
		fmt.Print("Input token stream file path is not specified, \nPlease enter input token stream: ")
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		inputString = scanner.Text()
	}

	// feed in inout token to PDA
	transitionsTaken, err := pdaProcessor.evaluateInput(inputString)
	if err != nil {
		log.Fatal(err)
	}

	// check if PDA accepted the input token
	acceptedToken := pdaProcessor.is_accepted()
	printFinalStatus(pdaProcessor, inputString, acceptedToken, transitionsTaken)

	pdaProcessor.reset(true)
	pdaProcessor.close()

	fmt.Println("End of main")
	fmt.Println("PDA has finished executing the given input token stream")
	fmt.Println("***************************************************************")
}

/**
read input token stream from specified file path
*/
func readInputFromFile(inputFilePath string) (string, error) {
	filebuffer, err := ioutil.ReadFile(inputFilePath)
	if err != nil {
		log.Fatal(err)
	}
	return string(filebuffer), nil
}

func printFinalStatus(pdaProcessor *PDAProcessor, inputString string, acceptedToken bool, transitionsTaken []string) {
	fmt.Println("\n************************** Result **************************")
	if acceptedToken {
		fmt.Printf("PDA %q successfully accepted the input: %q \n", pdaProcessor.Name, inputString)
	} else {
		fmt.Printf("PDA %q rejected the input: %q \n", pdaProcessor.Name, inputString)
	}

	fmt.Printf("Final state: %q, Accepting states: %+v, Stack: %+v\n", pdaProcessor.CurrentState, pdaProcessor.AcceptingStates, pdaProcessor.Stack)

	if len(transitionsTaken) <= 0 {
		fmt.Println("No transition taken")
	} else if len(transitionsTaken) == 1 {
		fmt.Println("Transition taken:", transitionsTaken[0])
	} else {
		fmt.Println("Transitions taken: ")
		var previousState = ""
		for i, currentState := range transitionsTaken {
			if i == 0 {
				previousState = currentState
			} else {
				fmt.Printf("%q => %q, \n", previousState, currentState)
				previousState = currentState
			}
		}
		fmt.Println()
	}

	//fmt.Print("*************************************************************\n\n")
}
