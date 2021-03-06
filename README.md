## Movie Booking

A sample Node.js app to demonstrate **__fabric-client__** & **__fabric-ca-client__** Node.js SDK APIs

### Prerequisites and setup:

* [Docker](https://www.docker.com/products/overview) - v19.03.8
* [Docker Compose](https://docs.docker.com/compose/overview/) - v1.25.4
* [Git client](https://git-scm.com/downloads) - needed for clone commands
* **Node.js** v8.17.0 or higher
* [Download Docker images](http://hyperledger-fabric.readthedocs.io/en/latest/samples.html#binaries)

```
cd moviebookings

```

Once you have completed the above setup, you will have provisioned a local network with the following docker container configuration:

* 2 CAs
* A SOLO orderer
* 4 peers (2 peers per Org)
* CouchDB for each peer

Chaincode location:
artifacts/src/github.com/chaincode/bookings - chaincode for Ticket booking management
artifacts/src/github.com/chaincode/movies - chaincode for Movie management


##### Terminal Window 1

Following file is to make the Blockchain network up and running with clinet node running at 4000
```
cd moviebookings

./runApp.sh

```

* This launches the required network on your local machine
* Installs the fabric-client and fabric-ca-client node modules
* And, starts the node app on PORT 4000

##### Terminal Window 2


In order for the following shell script to properly parse the JSON, you must install ``jq``:

instructions [https://stedolan.github.io/jq/](https://stedolan.github.io/jq/)

With the application started in terminal 1, next, test the APIs by executing the script - **testAPIs.sh**:
```
cd moviebookings

## For Channel Creation, Join Channel and Install/Instantiate channel run following 

./networkSetup.sh

#To run test cases

./testAPIs.sh

```
