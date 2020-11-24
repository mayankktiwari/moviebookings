package main

import (
    "encoding/json"
    "fmt"
    "strconv"
    "time"
    "strings"

    "github.com/hyperledger/fabric/common/util"
    "github.com/hyperledger/fabric/core/chaincode/shim"
    pb "github.com/hyperledger/fabric/protos/peer"
)


var logger = shim.NewLogger("Chaincode for Movie Bookings")

// Chaincode is the definition of the Booking Chaincode structure.
type BookingChaincode struct {
}

type BookingDetails struct {
    BookedByUser string `json:"bookedByUser"`
    MovieName string `json:"movieName"`
    TimeSlot string `json:"timeSlot"`
    ReqNmbrOfTickets int `json:"reqNmbrOfTickets"`
    BookingId string `json:"bookingId"`
    ReceiptId string `json:"receiptId"`
    BeverageFlag string `json:"beverageFlag"`
    BookingTime time.Time `json:"bookingTime"`
}

type movie struct {
    MovieName string `json:"movieName"`
    AvailalbeTimeSlots string `json:"availalbeTimeSlots"`
    TotalTickets int `json:"totalTickets"`
    RemainingTickets int `json:"remainingTickets"`
    HouseFullFlag string `json:"houseFullFlag"`
    ModificationTime time.Time `json:"modificationTime"`
}

// ===================================================================================
// Main
// ===================================================================================
func main() {
    err := shim.Start(new(BookingChaincode))
    if err != nil {
        fmt.Printf("Error starting BookingChaincode chaincode: %s", err)
    }
}

// Init initializes chaincode
// ===========================
func(t * BookingChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
    return shim.Success(nil)
}

// Invoke - Entry point for Invocations
// ========================================
func(t * BookingChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {

	function, args := stub.GetFunctionAndParameters()
	
    fmt.Println("invoke is running " + function)

    // Handle different functions
    if function == "initBookingDetails" { // Making Booking Details for Users
        return t.initBookingDetails(stub, args)
    } else if function == "getShowDetailsByTimeSlot" { // Get the Details according to the TimeSlot
        return t.getShowDetailsByTimeSlot(stub, args)
	}
	
    fmt.Println("invoke did not find func: " + function) //error
    return shim.Error("Received unknown function invocation")
}

// initBookingDetails - Creating record Movie name, time slots and total ticket for a show
func(t * BookingChaincode) initBookingDetails(stub shim.ChaincodeStubInterface, args[] string) pb.Response {

    logger.Info("########### START - initBookingDetails ###########")
	
	var err error
    if len(args) != 4 {
        return shim.Error("Incorrect number of arguments. Expecting 4")
	}
	
    // Params for Ticket Bookings
    bookedByUser := args[0]
    movieName := args[1]
    timeSlot := args[2]
    reqNmbrOfTickets, err := strconv.Atoi(args[3])
    if err != nil {
        return shim.Error("Expecting an integer value for Booking Number of Tickets")
    }
    currentTime := time.Now().Unix()
    bookingId := bookedByUser + "_" + strconv.Itoa(int(currentTime))

    logger.Info("Booking Details: ", bookedByUser, movieName, timeSlot, reqNmbrOfTickets)

    // ---- CALLING MOVIES CHAINCODE TO CHECK AVAILABILITY ---- //
    chainCodeArgs := util.ToChaincodeArgs("getMoviesByName", movieName)
    response := stub.InvokeChaincode("cc_movies", chainCodeArgs, "mychannel")
    var m movie
    json.Unmarshal(response.Payload, & m)
    // logger.Info("Chaincode Response: ", m)

    resMovieName := m.MovieName
    resTimeSlots := m.AvailalbeTimeSlots
    resTotalTicketsInt := m.TotalTickets
    resRemainingTickets := m.RemainingTickets
    resHouseFullFlag := m.HouseFullFlag

    logger.Info("Output of existing movie: ", resMovieName, resTimeSlots, resTotalTicketsInt, resRemainingTickets, resHouseFullFlag)

    // ---- Verify following before booking tickets for user
    // 1. Requested movie exists
    // 2. Booking available for the requested time slot
    // 3. Movie is not housefull yet
    if strings.ToUpper(resMovieName) == strings.ToUpper(movieName) && resTimeSlots == timeSlot && strings.ToUpper(resHouseFullFlag) == "FALSE" {

		remainingTicketsInt := resRemainingTickets - reqNmbrOfTickets
        remainingTicketsStr := strconv.Itoa(remainingTicketsInt)
        resTotalTicketsStr := strconv.Itoa(resTotalTicketsInt)

        // Check whether seats available and remaining seats are greater then booked seat, book the seats for user.
        if resRemainingTickets > 0 && resRemainingTickets > reqNmbrOfTickets {

            bookingDetailsList := []BookingDetails{}
            i := 0
            for i<reqNmbrOfTickets {

                currTime := time.Now().Unix()
                receiptId := strconv.Itoa(int(currTime))
                beverageFlag := "True"

                fmt.Println("Pringting Data for Receipt ID: ", receiptId)

                BookingDetailsObj := BookingDetails {
                    BookedByUser: bookedByUser,
                    MovieName: movieName,
                    TimeSlot: timeSlot,
                    ReqNmbrOfTickets: reqNmbrOfTickets,
                    BookingId: bookingId,
                    ReceiptId: receiptId,
                    BeverageFlag: beverageFlag,
                    BookingTime: time.Now() }

                    bookingDetailsList = append(bookingDetailsList, BookingDetailsObj)
                    i = i + 1
                }

                j := 0
                for j < len(bookingDetailsList) {
                    fmt.Println("j is ", j)
                    bookingDetailsAsBytes, _ := json.Marshal(bookingDetailsList[j])
                    stub.PutState(bookingDetailsList[j].BookedByUser+strconv.Itoa(i), bookingDetailsAsBytes)
                    fmt.Println("Added", bookingDetailsList[j])
                    j = j + 1
                }
			
            // Updating the Movie data -- calling initMovieDetails after show book for user
            chainCodeArgs := util.ToChaincodeArgs("initMovieDetails", resMovieName, resTimeSlots, resTotalTicketsStr, remainingTicketsStr, resHouseFullFlag)
        	response := stub.InvokeChaincode("cc_movies", chainCodeArgs, "mychannel")

            if response.Status != shim.OK {
                return shim.Error(response.Message)
            }

            eventMessage := "{ \"message\" : \"Movie show booked succcessfully\", \"Booking ID\" : \"" + bookingId + "\", \"code\" : \"200\"}"
            err = stub.SetEvent("evtsender", [] byte(eventMessage))
            if err != nil {
                return shim.Error(err.Error())
            }
            // ELSE - Check if the requested number of seats are available for booking or not
        } else if resRemainingTickets < reqNmbrOfTickets {

            var remTickets int
            if remainingTicketsInt != 0 && remainingTicketsInt < 0 {
                remTickets = remainingTicketsInt * -1
            } else {
                remTickets = remainingTicketsInt
            }
            remTicketsStr := strconv.Itoa(remTickets)
            eventMessage := "{ \"Available Tickets\" : \"" + remTicketsStr + "\", \"message\" : \"Only limited seats are available.\", \"code\" : \"200\"}"
            err = stub.SetEvent("evtsender", [] byte(eventMessage))
            if err != nil {
                return shim.Error(err.Error())
            }
        } else if remainingTicketsInt == 0 {
            houseFullFlag := "True"

                chainCodeArgs := util.ToChaincodeArgs("initMovieDetails", resMovieName, resTimeSlots, resTotalTicketsStr, remainingTicketsStr, houseFullFlag)
                response := stub.InvokeChaincode("cc_movies", chainCodeArgs, "mychannel")

            if response.Status != shim.OK {
                return shim.Error(response.Message)
            }

            eventMessage := "{ \"message\" : \"Movie show is not available for Booking\", \"code\" : \"200\"}"
                err = stub.SetEvent("evtsender", [] byte(eventMessage))
            if err != nil {
                return shim.Error(err.Error())
            }
        }
    } else {
        return shim.Error("Can not book selected movie!!")
    }

    fmt.Println("- end Movie Booking request")

    return shim.Success(nil)
}

func(t * BookingChaincode) getShowDetailsByTimeSlot(stub shim.ChaincodeStubInterface, args[] string) pb.Response {

    var timeSlot, jsonResp string
    var err error

    if len(args) != 1 {
        return shim.Error("Incorrect number of arguments. Expecting Time Slot to fetch the details")
    }
    timeSlot = args[0]

    valAsbytes, err := stub.GetState(timeSlot) //get the Incident details from chaincode state
    if err != nil {
        jsonResp = "{\"Error\":\"Failed to get state for given TimeSlot" + timeSlot + "\"}"
        return shim.Error(jsonResp)
    } else if valAsbytes == nil {
        jsonResp = "{\"Error\":\"No Movie show is running for the requested time slot: " + timeSlot + "\"}"
        return shim.Error(jsonResp)
    }

    return shim.Success(valAsbytes)
}