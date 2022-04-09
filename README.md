# Mobility and Replication for PDA Processors

## Description

A PushDown Automaton (PDA) is a simple computing device capable of performing quite interesting and powerful computations (eg parsing programming languages). Computer Scientists design PDAs to recognize whether a token-stream is in some desired language (set of token-streams).

In this project, I have implemented a PDA processor in Go.

A PDA is a Finite State Automaton (FSA) endowed with an unlimited-size stack. The stack essentially extends the finite amount of memory available to the FSA in the form of finite set of possible states for its state control (domain of values for the automaton’s single register).

PDA processor accepts a specification in JSON file and input token stream to evaluate. Both of these filepath are provided as an arguments to main PDA driver

#### What is PDA Specification?
PDA specification is a JSON file containing a JSON string to define the PDA name, number of states, valid input and stack alphabets, starting state and list of valid transitions

Sample PDA specification for a regular expression/language accepting the input in the form of 0^n1^n
```
{ 
  "name": "0^n1^n", 
  "states": [ "q1", "q2", "q3", "q4" ], 
  "input_alphabet": [ "0", "1" ], 
  "stack_alphabet": [ "0", "1" ], 
  "accepting_states": [ "q1", "q4" ], 
  "start_state": "q1", 
  "transitions": [ 
     [ "q1", null, null, "q2", "$" ], 
     [ "q2", "0", null, "q2", "0" ], 
     [ "q2", "1", "0", "q3", null ], 
     [ "q3", "1", "0", "q3", null ], 
     [ "q3", null, "$", "q4", null ] 
   ], 
   "eos": "$" 
} 
```
#### What is PDA Input Token Stream?
An input token stream is a sequence of tokens presented to the PDA for processing. e.g. in above PDA specification, a token stream could be 00001111

#### How does the PDA processes Input Token Stream?
The PDA processes an input sequence of tokens (token-stream) as follows. The PDA is presented with a single token at a time with any position. The PDA, upon presented with the current (next) input token, inspects whether all the tokens before this position are already consumed, if it is so then it will immediately consume the current token and make the appropriate transition. If the position of the currently presented token is not the one PDA is expecting to consume then it will push it pending tokens queue for processing it later.

#### How to Run PDA?
The bash script ```run-rest-server.sh``` is used to build the go project and to deploy RESTful PDA server/s at specified port number.

##### Start replica Server at default port 8801
```➜  pda-processor$ ./run-rest-server.sh ```

##### Start PDA Server at Specific Port
```➜  pda-processor$ ./run-rest-server.sh 1010```

## PDA Enhancements 

PDA Processor is enhanced to support Replication of PDA processors and client mobility with monotonic-write client consistency. Application is using session-id based approach to maintain client consistency across PDA servers and uniquely identifying the instance of PDA processor. 

Here are the available RESTful APIs, Base URL: http://localhost:8801

| HTTP Method  | URL                          | HTTP Headers        | HTTP Request Body                              | Function                                                              |
| -------------|------------------------------|---------------------|------------------------------------------------| ----------------------------------------------------------------------|
| GET          | base/pdas                    | none                | none                                           | List of names of PDAs available at the server |
| PUT          | base/pdas/id                 | none                | PDA Specification                              | Create at the server a PDA with the given id and the specification provided in the body of the request; calls `open()` method of PDA processor|
| PUT          | base/pdas/id/reset           | session-id required | none                                           | Call `reset()` method |
| PUT          | base/pdas/id/token/position  | session-id required | none                                           | Present a token at the given position |
| POST         | base/pdas/id/eos/position    | session-id required | none                                           | Call `eos()` with no tokens after (excluding) position <br/><br/> **Note**: This was supposed to be PUT method but the URL was conflicting with `base/pdas/id/token/position` as `token` can be any text. Updated method to POST  |
| GET          | base/pdas/id/is_accepted     | session-id required | none                                           | Call and return the value of `is_accepted()` |
| GET          | base/pdas/id/stack/top/k     | session-id required | none                                           | Call and return the value of `peek(k)` |
| GET          | base/pdas/id/stack/len       | session-id required | none                                           | Return the number of tokens currently in the stack |
| GET          | base/pdas/id/state           | session-id required | none                                           | Call and return the value of `current_state()` |
| GET          | base/pdas/id/tokens          | session-id required | none                                           | Call and return the value of `queued_tokens()` |
| GET          | base/pdas/id/snapshot/k      | session-id required | none                                           | Return a JSON message (array) three components: `the current_state()`, `queued_tokens()`, and `peek(k)` |
| PUT          | base/pdas/id/close           | session-id required | none                                           | Call `close()` |
| DELETE       | base/pdas/id/delete          | none                | none                                           | Delete the PDA with name from the server |
| GET          | base/pdas/id/createSession   | none                | none                                           | This is one additional API which is used to create a session for an user to interact with dedicated PDA instance. It  returns session id which is expected in HTTP header for all the above REST API calls to access dedicated PDA instance |
| GET          | base/replica_pdas            | none                | none                                           | Return list of ids of replica groups currently defined |
| PUT          | base/replica_pdas/gid        | none                | Replica Group structure with PDA Specification | Define a new replica group with the given member PDA addresses sharing the specification given in pda_code; create/replace the group members (as needed) |
| GET          | base/replica_pdas/gid/members| none                | none                                           | Return a JSON array with the addresses of its members |
| GET          | base/replica_pdas/gid/connect| none                | none                                           | Return the address of a random member that a client could connect to, This uses Golang's ```rand.Intn(...)``` method to generate random integer between 0 and no. of PDA members in the replica group |
| PUT          | base/replica_pdas/gid/close  | none                | none                                           | Close the members PDAs. This call close method on PDA members |
| DELETE       | base/replica_pdas/gid/delete | none                | none                                           | Delete replica group and all its members |
| PUT          | base/replica_pdas/gid/join   | none                | Replica Group details                          | The PDA id joins the replica group with the given address |


### PDA Implementation  
The PDA supports concurrent client sessions by maintaining session id per client per PDA. Client needs to create a session by calling `/pdas/{id}/createSession` API which returns a session id. This session id is expected in HTTP header to access client specific PDA instance.
This is included in demo screenshots where 2 independent sessions are created for same PDA from 2 different browsers/clients. 

The PDA implementation is written in below files.
#### PDA Server Implementation files
1. PDARestController.go
2. PDAProcessor.go
3. PDAService.go
4. PDAConstants.go

#### Replica Server Implementation files
1. PDAReplicaRestController.go
2. PDAReplicaService.go

