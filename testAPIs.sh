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
# Setting Org2
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
echo " --- CREATE CHANNEL --- "
curl -s -X POST \
http://localhost:4000/channels \
-H "authorization: Bearer $ORG1_TOKEN" \
-H "content-type: application/json" \
-d '{
        "channelName":"mychannel",
        "channelConfigPath":"../artifacts/channel/mychannel.tx"
}'
echo
echo
echo
sleep 5
echo " --- JOIN CHANNEL - ORG1 --- "
curl -s -X POST \
http://localhost:4000/channels/mychannel/peers \
-H "authorization: Bearer $ORG1_TOKEN" \
-H "content-type: application/json" \
-d '{
        "peers": ["peer0.org1.example.com","peer1.org1.example.com"]
}'
echo
echo
echo " --- JOIN CHANNEL - ORG2 --- "
curl -s -X POST \
http://localhost:4000/channels/mychannel/peers \
-H "authorization: Bearer $ORG2_TOKEN" \
-H "content-type: application/json" \
-d '{
        "peers": ["peer0.org2.example.com","peer1.org2.example.com"]
}'
echo
echo
echo
echo " --- INSTALL MOVIE CHAINCODE - ORG1 --- "
curl -s -X POST \
http://localhost:4000/chaincodes \
-H "authorization: Bearer $ORG1_TOKEN" \
-H "content-type: application/json" \
-d "{
        \"peers\": [\"peer0.org1.example.com\",\"peer1.org1.example.com\"],
        \"chaincodeName\":\"cc_movies\",
        \"chaincodePath\":\"$CC_MOVIES_SRC_PATH\",
        \"chaincodeType\": \"$LANGUAGE\",
        \"chaincodeVersion\":\"v0\"
}"
echo
echo
echo " --- INSTALL MOVIE CHAINCODE - ORG2 --- "
curl -s -X POST \
http://localhost:4000/chaincodes \
-H "authorization: Bearer $ORG2_TOKEN" \
-H "content-type: application/json" \
-d "{
        \"peers\": [\"peer0.org2.example.com\",\"peer1.org2.example.com\"],
        \"chaincodeName\":\"cc_movies\",
        \"chaincodePath\":\"$CC_MOVIES_SRC_PATH\",
        \"chaincodeType\": \"$LANGUAGE\",
        \"chaincodeVersion\":\"v0\"
}"
echo
echo
echo
echo " --- INSTANTIATE MOVIE CHAINCODE --- "
curl -s -X POST \
http://localhost:4000/channels/mychannel/chaincodes \
-H "authorization: Bearer $ORG1_TOKEN" \
-H "content-type: application/json" \
-d "{
        \"chaincodeName\":\"cc_movies\",
        \"chaincodeVersion\":\"v0\",
        \"chaincodeType\": \"$LANGUAGE\",
        \"args\":[]
}"
echo
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
            "args":["Shawshank Redemption", "09am - 12pm", "100", "100", "false"]
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
            "args":["Green Mile", "09am - 12pm", "100", "100", "false"]
}'
)
echo "Transaction ID is $TRX_ID"
echo
echo
echo
curl -s -X GET \
"http://localhost:4000/channels/mychannel/chaincodes/cc_movies?peer=peer0.org1.example.com&fcn=getMoviesByTimeSlot&args=%5B%2209am - 12pm%22%5D" \
-H "authorization: Bearer $ORG1_TOKEN" \
-H "content-type: application/json"
echo
echo
echo
echo " --- INSTALL BOOKING CHAINCODE - ORG1 --- "
curl -s -X POST \
http://localhost:4000/chaincodes \
-H "authorization: Bearer $ORG1_TOKEN" \
-H "content-type: application/json" \
-d "{
        \"peers\": [\"peer0.org1.example.com\",\"peer1.org1.example.com\"],
        \"chaincodeName\":\"cc_bookings\",
        \"chaincodePath\":\"$CC_BOOKING_SRC_PATH\",
        \"chaincodeType\": \"$LANGUAGE\",
        \"chaincodeVersion\":\"v0\"
}"
echo
echo
echo " --- INSTALL BOOKING CHAINCODE - ORG2 --- "
curl -s -X POST \
http://localhost:4000/chaincodes \
-H "authorization: Bearer $ORG2_TOKEN" \
-H "content-type: application/json" \
-d "{
        \"peers\": [\"peer0.org2.example.com\",\"peer1.org2.example.com\"],
        \"chaincodeName\":\"cc_bookings\",
        \"chaincodePath\":\"$CC_BOOKING_SRC_PATH\",
        \"chaincodeType\": \"$LANGUAGE\",
        \"chaincodeVersion\":\"v0\"
}"
echo
echo
echo
echo " --- INSTATIATE BOOKING CHAINCODE --- "
curl -s -X POST \
http://localhost:4000/channels/mychannel/chaincodes \
-H "authorization: Bearer $ORG1_TOKEN" \
-H "content-type: application/json" \
-d "{
        \"chaincodeName\":\"cc_bookings\",
        \"chaincodeVersion\":\"v0\",
        \"chaincodeType\": \"$LANGUAGE\",
        \"args\":[]
}"
echo
echo
echo
echo " --- INVOKE BOOKING CHAINCODE - ORG1 --- "
TRX_ID=$(
    curl -s -X POST \
    http://localhost:4000/channels/mychannel/chaincodes/cc_bookings \
    -H "authorization: Bearer $ORG1_TOKEN" \
    -H "content-type: application/json" \
    -d '{
                "peers": ["peer0.org1.example.com","peer1.org1.example.com"],
                "fcn":"initBookingDetails",
                "args":["Iron Man", "Shawshank Redemption", "09am - 12pm", "6"]
}'
)
echo "Transaction ID is $TRX_ID"
echo
echo
echo " --- INVOKE BOOKING CHAINCODE - ORG2 --- "
TRX_ID=$(
    curl -s -X POST \
    http://localhost:4000/channels/mychannel/chaincodes/cc_bookings \
    -H "authorization: Bearer $ORG2_TOKEN" \
    -H "content-type: application/json" \
    -d '{
            "peers": ["peer0.org2.example.com","peer1.org2.example.com"],
            "fcn":"initBookingDetails",
            "args":["Captain America", "Green Mile", "09am - 12pm", "9"]
}'
)
echo "Transaction ID is $TRX_ID"
echo
curl -s -X GET \
"http://localhost:4000/channels/mychannel/chaincodes/cc_bookings?peer=peer0.org1.example.com&fcn=getShowDetailsByTimeSlot&args=%5B%2209am - 12pm%22%5D" \
-H "authorization: Bearer $ORG1_TOKEN" \
-H "content-type: application/json"
echo
