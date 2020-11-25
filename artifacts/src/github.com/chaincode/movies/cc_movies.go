package main

import (
    "encoding/json"
    "fmt"
    "time"
    "strconv"

    "github.com/hyperledger/fabric/core/chaincode/shim"
    pb "github.com/hyperledger/fabric/protos/peer"
)
var logger = shim.NewLogger("Movie-Chaincode to Store Movies")

// MovieChaincode is the definition of the chaincode structure.
type MovieChaincode struct {}

type MovieDetails struct {
    MovieName string `json:"movieName"`
    AvailalbeTimeSlots string `json:"availalbeTimeSlots"`
    TotalTickets int `json:"totalTickets"`
    RemainingTickets int `json:"remainingTickets"`
    HouseFullFlag string `json:"houseFullFlag"`
    ModificationTime time.Time `json:"modificationTime"`
}

// --- Calling MAIN ---
func main() {
    err := shim.Start(new(MovieChaincode))
    if err != nil {
        fmt.Printf("Error starting MovieChaincode chaincode: %s", err)
    }
}

// Init initializes chaincode
func(t * MovieChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
    return shim.Success(nil)
}

// Invoke - Entry point for Invocations
func(t * MovieChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
    function, args := stub.GetFunctionAndParameters()
    fmt.Println("invoke is running " + function)
    // Handle different functions
    if function == "initMovieDetails" { //creates a new entry for Movie
        return t.initMovieDetails(stub, args)
    } else if function == "getMoviesByName" { // Get the Details according to the TimeSlot
        return t.getMoviesByName(stub, args)
    } else if function == "createDummyEntries" { // To create dummy data in DB
        return t.createDummyEntries(stub)
    }
    fmt.Println("invoke did not find func: " + function) //error
    return shim.Error("Received unknown function invocation")
}

func(t * MovieChaincode) createDummyEntries(stub shim.ChaincodeStubInterface) pb.Response {

	movieDetailsList := []MovieDetails{
        MovieDetails{MovieName: "The Grudge", AvailalbeTimeSlots: "9am-12pm", TotalTickets: 100, RemainingTickets: 100, HouseFullFlag: "False", ModificationTime: time.Now()},
        MovieDetails{MovieName: "The Grudge", AvailalbeTimeSlots: "12pm-3pm", TotalTickets: 100, RemainingTickets: 100, HouseFullFlag: "False", ModificationTime: time.Now()},
        MovieDetails{MovieName: "The Grudge", AvailalbeTimeSlots: "6pm-9pm", TotalTickets: 100, RemainingTickets: 3, HouseFullFlag: "False", ModificationTime: time.Now()},
        MovieDetails{MovieName: "The Godfather", AvailalbeTimeSlots: "9am-12pm", TotalTickets: 100, RemainingTickets: 0, HouseFullFlag: "True", ModificationTime: time.Now()},
        MovieDetails{MovieName: "The Godfather", AvailalbeTimeSlots: "12pm-3pm", TotalTickets: 100, RemainingTickets: 100, HouseFullFlag: "False", ModificationTime: time.Now()},
        MovieDetails{MovieName: "The Dark Knight", AvailalbeTimeSlots: "6pm-9pm", TotalTickets: 100, RemainingTickets: 100, HouseFullFlag: "False", ModificationTime: time.Now()} }

	i := 0
	for i < len(movieDetailsList) {
        tsId := strconv.FormatInt(time.Now().Unix(), 10)
		fmt.Println("i is ", i)
		moviesAsBytes, _ := json.Marshal(movieDetailsList[i])
		stub.PutState(tsId, moviesAsBytes)
		fmt.Println("Added", movieDetailsList[i])
		i = i + 1
	}

	return shim.Success(nil)
}

// initMovieDetails - Creating record Movie name, time slots and total ticket for a show
func(t * MovieChaincode) initMovieDetails(stub shim.ChaincodeStubInterface, args[] string) pb.Response {

	logger.Info("########### START - initMovieDetails ###########")
	
    var err error
    if len(args) != 5 {
        return shim.Error("Incorrect number of arguments. Expecting 5")
    }

    // Initializing the primary parameters for Movies
    movieName := args[0]
    availalbeTimeSlots := args[1]
    totalTickets, err := strconv.Atoi(args[2])
    if err != nil {
        return shim.Error("Expecting integer value for Remaining Tickets")
    }
    remainingTickets, err := strconv.Atoi(args[3])
    if err != nil {
        return shim.Error("Expecting integer value for Remaining Tickets")
    }
    houseFullFlag := args[4]

	logger.Info("Details about Movie: \n", movieName, availalbeTimeSlots, totalTickets, remainingTickets)

    // ==== Create  ====
    MoviesList := &MovieDetails {
        MovieName: movieName,
        AvailalbeTimeSlots: availalbeTimeSlots,
        TotalTickets: totalTickets,
        RemainingTickets: remainingTickets,
        HouseFullFlag: houseFullFlag,
        ModificationTime: time.Now() }

    moviesListAsBytes, err := json.Marshal(MoviesList)
    if err != nil {
        return shim.Error(err.Error())
    }

    // Write the state to the ledger
    err = stub.PutState(movieName, moviesListAsBytes)
    if err != nil {
        return shim.Error(err.Error())
    }

    // Create Index
    indexName := "indexMovieAndTime"
    movieTimeIndexKey, err := stub.CreateCompositeKey(indexName, []string {MoviesList.MovieName, MoviesList.AvailalbeTimeSlots})
    if err != nil {
        return shim.Error(err.Error())
    }

    value := []byte{0x00}
    stub.PutState(movieTimeIndexKey, value)
    eventMessage := "{ \"Movie\" : \"" + movieName + "\", \"message\" : \"Movie record created succcessfully\", \"code\" : \"200\"}"
    err = stub.SetEvent("evtsender", [] byte(eventMessage))
    if err != nil {
        return shim.Error(err.Error())
    }

    fmt.Println("- end Movie record creation request")
	logger.Info("Movie record created successfully")
    return shim.Success(nil)

}

func(t * MovieChaincode) getMoviesByName(stub shim.ChaincodeStubInterface, args[] string) pb.Response {
    var movieName, jsonResp string
    var err error
    if len(args) != 1 {
        return shim.Error("Incorrect number of arguments. Expecting Time Slot to fetch the details")
    }

	movieName = args[0]
    valAsbytes, err := stub.GetState(movieName) //get the Incident details from chaincode state
    if err != nil {
        jsonResp = "{\"Error\":\"Failed to get state for given TimeSlot" + movieName + "\"}"
        return shim.Error(jsonResp)
    } else if valAsbytes == nil {
        jsonResp = "{\"Error\":\"No Movie show is running for the requested time slot: " + movieName + "\"}"
        return shim.Error(jsonResp)
    }

    return shim.Success(valAsbytes)
}
