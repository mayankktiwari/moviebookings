#!/bin/bash
#
# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

jq --version >/dev/null 2>&1
if [ $? -ne 0 ]; then
    echo "Please Install 'jq' https://stedolan.github.io/jq/ to execute this script"
    echo
    exit 1
fi

starttime=$(date +%s)

LANGUAGE="golang"
CC_BOOKING_SRC_PATH="github.com/chaincode/bookings"
CC_MOVIES_SRC_PATH="github.com/chaincode/movies"

echo
echo "POST request Enroll on Org1  ..."
echo
# Setting Org1
ORG1_TOKEN=$(
    curl -s -X POST \
    http://localhost:4000/users \
    -H "content-type: application/x-www-form-urlencoded" \
    -d 'username=Jim&orgName=Org1'
)
echo
echo $ORG1_TOKEN
ORG1_TOKEN=$(echo $ORG1_TOKEN | jq ".token" | sed "s/\"//g")
echo
echo "ORG1 token is $ORG1_TOKEN"
echo
echo
ORG2_TOKEN=$(
    curl -s -X POST \
    http://localhost:4000/users \
    -H "content-type: application/x-www-form-urlencoded" \
    -d 'username=Barry&orgName=Org2'
)
echo $ORG2_TOKEN
ORG2_TOKEN=$(echo $ORG2_TOKEN | jq ".token" | sed "s/\"//g")
echo
echo "ORG2 token is $ORG2_TOKEN"
echo
echo
echo " --- INVOKE - CREATE DUMMY ENTRIES IN DB --- "
TRX_ID=$(
curl -s -X POST \
http://localhost:4000/channels/mychannel/chaincodes/cc_movies \
-H "authorization: Bearer $ORG1_TOKEN" \
-H "content-type: application/json" \
-d "{
        \"peers\": [\"peer0.org1.example.com\",\"peer1.org1.example.com\"],
        \"fcn\":\"createDummyEntries\",
        \"args\":[]
}"
)
echo
echo "Transaction ID is $TRX_ID"
echo
echo
echo " --- INVOKE MOVIE CHAINCODE - ORG1 --- "
TRX_ID=$(
    curl -s -X POST \
    http://localhost:4000/channels/mychannel/chaincodes/cc_movies \
    -H "authorization: Bearer $ORG1_TOKEN" \
    -H "content-type: application/json" \
    -d '{
            "peers": ["peer0.org1.example.com","peer1.org1.example.com"],
            "fcn":"initMovieDetails",
            "args":["Inception", "09am - 12pm", "100", "100", "False"]
}'
)
echo "Transaction ID is $TRX_ID"
echo
echo " --- INVOKE MOVIE CHAINCODE - ORG1 --- "
TRX_ID=$(
    curl -s -X POST \
    http://localhost:4000/channels/mychannel/chaincodes/cc_movies \
    -H "authorization: Bearer $ORG1_TOKEN" \
    -H "content-type: application/json" \
    -d '{
            "peers": ["peer0.org1.example.com","peer1.org1.example.com"],
            "fcn":"initMovieDetails",
            "args":["The Shawshank Redemption", "6pm-9pm", "100", "3", "False"]
}'
)
echo "Transaction ID is $TRX_ID"
echo
echo
echo " --- INVOKE MOVIE CHAINCODE - ORG2 --- "
TRX_ID=$(
    curl -s -X POST \
    http://localhost:4000/channels/mychannel/chaincodes/cc_movies \
    -H "authorization: Bearer $ORG2_TOKEN" \
    -H "content-type: application/json" \
    -d '{
            "peers": ["peer0.org2.example.com","peer1.org2.example.com"],
            "fcn":"initMovieDetails",
            "args":["The Godfather", "09am - 12pm", "100", "0", "True"]
}'
)
echo "Transaction ID is $TRX_ID"
echo
echo
echo " --- INVOKE BOOKING CHAINCODE - Simple Book Tickets --- "
TRX_ID=$(
    curl -s -X POST \
    http://localhost:4000/channels/mychannel/chaincodes/cc_bookings \
    -H "authorization: Bearer $ORG1_TOKEN" \
    -H "content-type: application/json" \
    -d '{
                "peers": ["peer0.org1.example.com","peer1.org1.example.com"],
                "fcn":"initBookingDetails",
                "args":["Steve Smith", "Inception", "09am - 12pm", "6"]
}'
)
echo "Transaction ID is $TRX_ID"
echo
echo
echo " --- INVOKE BOOKING CHAINCODE - When To-Be-Booked tickets are greater then Remaning tickets --- "
TRX_ID=$(
    curl -s -X POST \
    http://localhost:4000/channels/mychannel/chaincodes/cc_bookings \
    -H "authorization: Bearer $ORG1_TOKEN" \
    -H "content-type: application/json" \
    -d '{
                "peers": ["peer0.org1.example.com","peer1.org1.example.com"],
                "fcn":"initBookingDetails",
                "args":["Rahul Dravid", "The Grudge", "6pm-9pm", "6"]
}'
)
echo "Transaction ID is $TRX_ID"
echo
echo
echo
echo " --- INVOKE BOOKING CHAINCODE - Book Another Ticket for Another movie --- "
TRX_ID=$(
    curl -s -X POST \
    http://localhost:4000/channels/mychannel/chaincodes/cc_bookings \
    -H "authorization: Bearer $ORG2_TOKEN" \
    -H "content-type: application/json" \
    -d '{
            "peers": ["peer0.org2.example.com","peer1.org2.example.com"],
            "fcn":"initBookingDetails",
            "args":["David Warner", "The Dark Knight", "6pm-9pm", "9"]
}'
)
echo "Transaction ID is $TRX_ID"
