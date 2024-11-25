package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"strings"
	"time"

	pb "github.com/sofiadesimon15/auction-system/proto" // replace with actual path

	"google.golang.org/grpc"
)

func main() {
	// Define command-line flags
	serverAddr := flag.String("server", "localhost:5001", "gRPC server address")
	action := flag.String("action", "bid", "Action to perform: bid or result")
	bidderID := flag.String("bidder", "BidderA", "Bidder ID")
	amount := flag.Int("amount", 0, "Bid amount (required for bid action)")
	flag.Parse()

	// Validate action and required parameters
	if *action == "bid" && *amount <= 0 {
		log.Fatalf("Bid action requires a positive bid amount")
	}

	// Perform the requested action
	err := performAction(*serverAddr, *action, *bidderID, *amount)
	if err != nil {
		log.Fatalf("Action failed: %v", err)
	}
}

// performAction handles the bid or result action
func performAction(serverAddr, action, bidderID string, amount int) error {
	// Establish connection to the server
	conn, err := grpc.Dial(serverAddr, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return fmt.Errorf("could not connect to server %s: %v", serverAddr, err)
	}
	defer conn.Close()

	// Create AuctionService client
	client := pb.NewAuctionServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	switch action {
	case "bid":
		// Create BidRequest
		req := &pb.BidRequest{
			BidderId: bidderID,
			Amount:   int32(amount),
		}
		// Send Bid RPC
		resp, err := client.Bid(ctx, req)
		if err != nil {
			return fmt.Errorf("bid RPC failed: %v", err)
		}
		// Handle response
		if strings.HasPrefix(resp.Outcome, "fail") {
			fmt.Printf("Bid Outcome: %s\n", resp.Outcome)
		} else {
			fmt.Printf("Bid Outcome: %s\n", resp.Outcome)
		}
	case "result":
		// Create ResultRequest
		req := &pb.ResultRequest{}
		// Send Result RPC
		resp, err := client.Result(ctx, req)
		if err != nil {
			return fmt.Errorf("result RPC failed: %v", err)
		}
		// Display result
		fmt.Printf("Auction Result: %s\n", resp.Outcome)
	default:
		return fmt.Errorf("unknown action: %s", action)
	}
	return nil
}
