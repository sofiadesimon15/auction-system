# Distributed Auction System with Token Ring

This project implements a distributed auction system using a token ring to coordinate activity across multiple nodes. Clients can interact with the system to place bids (`bid`) or query the auction result (`result`).


# Running the Server
The server represents the nodes in the distributed system. Each node is configured with a nodeID, the address of the next node, and whether it initially holds the token.

Command to start a node:
```
go run server.go <nodeID> <nextNodeAddress> <initialToken>
```
Example:
```
go run server.go 1 localhost:5002 true
go run server.go 2 localhost:5003 false
go run server.go 3 localhost:5001 false
```
# Logs
Logs are saved in logs/log.txt.

# Running the Client
The client can interact with the system by sending requests to place a bid (bid) or query the result (result).

Command to run the client:
```
go run client.go --server=<serverAddress> --action=<bid|result> --bidder=<BidderID> --amount=<BidAmount>
```
Parameters:
--server: Server address (e.g., localhost:5001).
--action: bid to place a bid or result to query the result.
--bidder: Bidder ID (required for bid).
--amount: Bid amount (only required for bid).
Examples:
Place a bid:
```
go run client.go --server=localhost:5001 --action=bid --bidder=BidderA --amount=100
```
Query the result:
```
go run client.go --server=localhost:5001 --action=result
```

# Auction End
The auction automatically ends after a predefined time (100 seconds by default). The server logs the winner and stops passing the token.
