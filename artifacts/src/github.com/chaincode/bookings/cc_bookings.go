package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/hyperledger/fabric/common/util"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

var logger = shim.NewLogger("Chaincode for Movie Bookings")

// BookingChaincode - definition of the Booking Chaincode structure.
type BookingChaincode struct {
}

var strFlag = "ExchangeFlagData"

type BookingDetails struct {
	BookedByUser     string    `json:"bookedByUser"`
	MovieName        string    `json:"movieName"`
	TimeSlot         string    `json:"timeSlot"`
	ReqNmbrOfTickets int       `json:"reqNmbrOfTickets"`
	BookingId        string    `json:"bookingId"`
	SeatDetails      []SeatDetails    `json:"seatDetails"`
    BookingTime string `json:"bookingTime"`
}

type SeatDetails struct {
    SeatNumber    string     `json:"seatNumber"`
	ReceiptNumber string    `json:"receiptNumber"`
    BeverageFlag  string    `json:"beverageFlag"`
    WaterToSodaExchangeFlag string `json:"waterToSodaExchangeFlag"`
}

type DatewiseBeverageExchangeDetails struct {
    Date string `json:"date"`
    DailyQuota string `json:"dailyQuota"`
}

type movie struct {
	MovieName          string    `json:"movieName"`
	AvailalbeTimeSlots string    `json:"availalbeTimeSlots"`
	TotalTickets       int       `json:"totalTickets"`
	RemainingTickets   int       `json:"remainingTickets"`
	HouseFullFlag      string    `json:"houseFullFlag"`
	ModificationTime   time.Time `json:"modificationTime"`
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
func (t *BookingChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
    
    // Initializing the Water to Soda exchange flag
    current_time := time.Now()
    date := string(current_time.Format("2006-January-02"))
    dailyQuota := strconv.Itoa(200)
    datewiseBeverageExchangeDetails := DatewiseBeverageExchangeDetails{Date: date, DailyQuota: dailyQuota}
    datewiseBeverageExchangeBytes, _ := json.Marshal(datewiseBeverageExchangeDetails)
    stub.PutState(strFlag, datewiseBeverageExchangeBytes)
    
	return shim.Success(nil)
}

// Invoke - Entry point for Invocations
// ========================================
func (t *BookingChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {

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
func (t *BookingChaincode) initBookingDetails(stub shim.ChaincodeStubInterface, args []string) pb.Response {

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
	bookingId := bookedByUser + "_" + strconv.FormatInt(time.Now().Unix(), 10)

	logger.Info("Booking Details: ", bookedByUser, movieName, timeSlot, reqNmbrOfTickets)

	// ---- CALLING MOVIES CHAINCODE TO CHECK AVAILABILITY ---- //
	chainCodeArgs := util.ToChaincodeArgs("getMoviesByName", movieName)
	response := stub.InvokeChaincode("cc_movies", chainCodeArgs, "mychannel")
	var m movie
	json.Unmarshal(response.Payload, &m)
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

            // Creating list of SeatNumber, Receipts and Beverage Flag
			seatDetailsList := []SeatDetails{}
            i := 0
            var waterToSodaExchangeFlag string
			for i < reqNmbrOfTickets {
                seatNumber := strconv.Itoa(i)
                currTime := time.Now()
                currDateStr := string(currTime.Format("2006-January-02"))
				receiptNumber := strconv.FormatInt(currTime.Unix(), 10)
                beverageFlag := "True"

                // Fetching data for Soda/Water exchange
                remainingValueForDateBytes, _ := stub.GetState(strFlag)
                var data DatewiseBeverageExchangeDetails
                json.Unmarshal(remainingValueForDateBytes, &data)
                date := data.Date
                dailyQuota, _ := strconv.Atoi(data.DailyQuota)

                if dailyQuota > 0 && date == currDateStr {
                    waterToSodaExchangeFlag = "True"
                    // dailyQuotaNewVal := dailyQuota - 1
                    dailyQuotaNewVal := dailyQuota - reqNmbrOfTickets
                    exchangeCountRemaining := strconv.Itoa(dailyQuotaNewVal)
                    newData := DatewiseBeverageExchangeDetails{Date: date, DailyQuota: exchangeCountRemaining}
                    newDataBytes, _ := json.Marshal(newData)
                    stub.PutState(strFlag, newDataBytes)
                } else {
                    waterToSodaExchangeFlag = "False"
                }

                // Putting data into DatewiseBeverageExchangeDetails for next date
                if dailyQuota == 0 && date == currDateStr {
                    nextDay := currTime.AddDate(0, 0, 1)
                    date := string(nextDay.Format("2006-January-02"))
                    dailyQuota := strconv.Itoa(200)
                    datewiseBeverageExchangeDetails := DatewiseBeverageExchangeDetails{Date: date, DailyQuota: dailyQuota}
                    datewiseBeverageExchangeBytes, _ := json.Marshal(datewiseBeverageExchangeDetails)
                    stub.PutState(strFlag, datewiseBeverageExchangeBytes)
                }

				fmt.Println("Receipt ID: ", receiptNumber)
				fmt.Println("Seat Number: ", seatNumber)
				seatDetailsObj := SeatDetails{SeatNumber: seatNumber, ReceiptNumber: receiptNumber, BeverageFlag: beverageFlag, WaterToSodaExchangeFlag: waterToSodaExchangeFlag}
				seatDetailsList = append(seatDetailsList, seatDetailsObj)
				i = i + 1
            }
            
            currTime := time.Now()
            bookingTime := currTime.Format(time.RFC3339Nano)

			BookingDetailsObj := BookingDetails{
				BookedByUser:     bookedByUser,
				MovieName:        movieName,
				TimeSlot:         timeSlot,
				ReqNmbrOfTickets: reqNmbrOfTickets,
				BookingId:        bookingId,
				SeatDetails:      seatDetailsList,
				BookingTime:      bookingTime }

			bookingDetailsAsBytes, _ := json.Marshal(BookingDetailsObj)
			err = stub.PutState(bookedByUser, bookingDetailsAsBytes)
			if err != nil {
				return shim.Error(err.Error())
			}

			// Updating the Movie data -- calling initMovieDetails after show book for user
			chainCodeArgs := util.ToChaincodeArgs("initMovieDetails", resMovieName, resTimeSlots, resTotalTicketsStr, remainingTicketsStr, resHouseFullFlag)
			response := stub.InvokeChaincode("cc_movies", chainCodeArgs, "mychannel")

			if response.Status != shim.OK {
				return shim.Error(response.Message)
			}

			eventMessage := "{ \"message\" : \"Movie show booked succcessfully\", \"Booking ID\" : \"" + bookingId + "\", \"code\" : \"200\"}"
			err = stub.SetEvent("evtsender", []byte(eventMessage))
			if err != nil {
				return shim.Error(err.Error())
			}

            msg := "Show booked successfully. Booking ID: " + bookingId
            logger.Info(msg)
			return shim.Success([]byte(msg))

		} else if resRemainingTickets < reqNmbrOfTickets { // ELSE - Check if the requested number of seats are available for booking or not

			var remTickets int
			if remainingTicketsInt != 0 && remainingTicketsInt < 0 {
				remTickets = remainingTicketsInt * -1
			} else {
				remTickets = remainingTicketsInt
			}

			remTicketsStr := strconv.Itoa(remTickets)

			eventMessage := "{ \"Available Tickets\" : \"" + remTicketsStr + "\", \"message\" : \"Only limited seats are available.\", \"code\" : \"200\"}"
			err = stub.SetEvent("evtsender", []byte(eventMessage))
			if err != nil {
				return shim.Error(err.Error())
			}

            msg := "Only limited seats are available. Remaning Seat: " + remTicketsStr
            logger.Info(msg)
			return shim.Success([]byte(msg))

		} else if remainingTicketsInt == 0 { // ELSE - when there is zero seats available.

			houseFullFlag := "True"

			chainCodeArgs := util.ToChaincodeArgs("initMovieDetails", resMovieName, resTimeSlots, resTotalTicketsStr, remainingTicketsStr, houseFullFlag)
			response := stub.InvokeChaincode("cc_movies", chainCodeArgs, "mychannel")

			if response.Status != shim.OK {
				return shim.Error(response.Message)
			}

			eventMessage := "{ \"message\" : \"Movie show is not available for Booking\", \"code\" : \"200\"}"
			err = stub.SetEvent("evtsender", []byte(eventMessage))
			if err != nil {
				return shim.Error(err.Error())
			}

            msg := "Selected time slot for " + movieName + " is housefull already."
            logger.Info(msg)
			return shim.Success([]byte(msg))
		}
	} else {
        msg := "Requested movie is not available for booking."
        logger.Info(msg)
		return shim.Success([]byte(msg))
	}

	fmt.Println("- end Movie Booking request")

	return shim.Success(nil)
}

func (t *BookingChaincode) getShowDetailsByTimeSlot(stub shim.ChaincodeStubInterface, args []string) pb.Response {

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
