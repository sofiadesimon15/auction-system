package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"sync"
	"time"

	pb "github.com/sofiadesimon15/auction-system/proto"

	"google.golang.org/grpc"
)

// AuctionState maintains the state of the auction
type AuctionState struct {
	HighestBid    int32
	HighestBidder string
	IsEnded       bool
	StartTime     time.Time
	EndTime       time.Time
	mutex         sync.Mutex
}

// TokenRingServer handles token passing
type TokenRingServer struct {
	pb.UnimplementedTokenRingServer
	nodeID       int32
	hasToken     bool
	nextNode     string
	auctionState *AuctionState
	mutex        sync.Mutex
	lastPass     time.Time
	timeout      time.Duration
}

// AuctionServiceServer handles auction operations
type AuctionServiceServer struct {
	pb.UnimplementedAuctionServiceServer
	auctionState *AuctionState
	node         *TokenRingServer
}

// NewTokenRingServer creates a new TokenRingServer instance
func NewTokenRingServer(nodeID int32, nextNode string, initialToken bool, auctionState *AuctionState) *TokenRingServer {
	return &TokenRingServer{
		nodeID:       nodeID,
		hasToken:     initialToken,
		nextNode:     nextNode,
		auctionState: auctionState,
		lastPass:     time.Now(),
		timeout:      10 * time.Second, //  timeout duration for token passing
	}
}

// PassToken receives the token from the previous node in the ring
func (s *TokenRingServer) PassToken(ctx context.Context, tokenMsg *pb.TokenMessage) (*pb.Response, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	fmt.Printf("Node %d received token message: TokenHolderId = %d, NodeID = %d\n", s.nodeID, tokenMsg.TokenHolderId)

	// Validate TokenHolderId
	if tokenMsg.TokenHolderId == s.nodeID-1 || (s.nodeID == 1 && tokenMsg.TokenHolderId == 3) { // Adjust for first node case
		s.hasToken = true
		fmt.Printf("Node %d has received the token and is entering the critical section\n", s.nodeID)
		go s.enterCriticalSection()
	} else {
		fmt.Printf("Node %d received token message but it does not match its ID. TokenHolderId = %d\n", s.nodeID, tokenMsg.TokenHolderId)
	}

	return &pb.Response{Message: "Token passed"}, nil
}

// enterCriticalSection simulates work in the critical section and then passes the token
func (s *TokenRingServer) enterCriticalSection() {
	fmt.Printf("Node %d is entering critical section\n", s.nodeID)
	time.Sleep(2 * time.Second) // simulate critical section work
	fmt.Printf("Node %d leaving critical section\n", s.nodeID)

	// After finishing the critical section, pass the token to the next node
	s.mutex.Lock()
	s.hasToken = false
	s.lastPass = time.Now()
	s.mutex.Unlock()
	s.passToken()
}

// passToken sends the token to the next node in the ring
func (s *TokenRingServer) passToken() {
	for {
		conn, err := grpc.Dial(s.nextNode, grpc.WithInsecure())
		if err != nil {
			log.Printf("Node %d failed to connect to next node %s: %v", s.nodeID, s.nextNode, err)
			time.Sleep(1 * time.Second) // Retry after a delay
			continue
		}
		defer conn.Close()

		client := pb.NewTokenRingClient(conn)
		// Pass the current node's ID as TokenHolderId
		_, err = client.PassToken(context.Background(), &pb.TokenMessage{TokenHolderId: s.nodeID})
		if err != nil {
			log.Printf("Node %d failed to pass token: %v", s.nodeID, err)
			time.Sleep(1 * time.Second) // Retry after a delay
			continue
		}

		// Log token passing and print that the node passed the token to the next node
		log.Printf("Node %d passed the token to node at %s", s.nodeID, s.nextNode)
		fmt.Printf("Node %d passed the token to node at %s\n", s.nodeID, s.nextNode)
		break
	}
}

// moniterTokenPassing moniters token passing and regenerates token if timeout occurs
func (s *TokenRingServer) moniterTokenPassing() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		<-ticker.C
		s.mutex.Lock()
		if s.hasToken && time.Since(s.lastPass) > s.timeout {
			log.Printf("Node %d detected token pass timeout. Regenerating token", s.nodeID)
			s.hasToken = false
			go s.passToken()
		}
		s.mutex.Unlock()

		// Check if Auction has ended
		s.auctionState.mutex.Lock()
		ended := s.auctionState.IsEnded
		s.auctionState.mutex.Unlock()
		if ended {
			log.Printf("Node %d detected auction end. Stopping token passing.", s.nodeID)
			return
		}
	}
}

/*// AuctionService implements the AuctionServiceServer interface
type AuctionService struct {
	pb.UnimplementedAuctionServiceServer
	auctionState *AuctionState
	node         *TokenRingServer
}*/

// NewAuctionService creates a new AuctionService instance
func NewAuctionService(auctionState *AuctionState, node *TokenRingServer) *AuctionServiceServer {
	return &AuctionServiceServer{
		auctionState: auctionState,
		node:         node,
	}
}

// Bid handles bid requests from clients
func (a *AuctionServiceServer) Bid(ctx context.Context, req *pb.BidRequest) (*pb.BidResponse, error) {
	a.node.mutex.Lock()
	hasToken := a.node.hasToken
	a.node.mutex.Unlock()

	if !hasToken {
		// Redirect to the node holding the token
		return &pb.BidResponse{Outcome: "fail: no token available"}, nil
	}

	a.auctionState.mutex.Lock()
	defer a.auctionState.mutex.Unlock()

	if a.auctionState.IsEnded {
		return &pb.BidResponse{Outcome: "fail: auction ended"}, nil
	}

	if req.Amount <= a.auctionState.HighestBid {
		return &pb.BidResponse{Outcome: "fail: bid too low"}, nil
	}

	// Update auction state
	a.auctionState.HighestBid = req.Amount
	a.auctionState.HighestBidder = req.BidderId

	log.Printf("[Auction] New highest bid: %d by %s", a.auctionState.HighestBid, a.auctionState.HighestBidder)

	return &pb.BidResponse{Outcome: "success"}, nil
}

// Result handles result queries from clients
func (a *AuctionServiceServer) Result(ctx context.Context, req *pb.ResultRequest) (*pb.ResultResponse, error) {
	a.auctionState.mutex.Lock()
	defer a.auctionState.mutex.Unlock()

	if a.auctionState.IsEnded {
		return &pb.ResultResponse{
			Outcome: fmt.Sprintf("Winner: %s with bid %d", a.auctionState.HighestBidder, a.auctionState.HighestBid),
		}, nil
	} else {
		return &pb.ResultResponse{
			Outcome: fmt.Sprintf("Current Highest Bid: %d by %s", a.auctionState.HighestBid, a.auctionState.HighestBidder),
		}, nil
	}
}

func main() {
	if len(os.Args) < 4 {
		log.Fatalf("Usage: go run node.go <nodeID> <nextNodeAddress> <initialToken(true|false)>")
	}

	// Parse the nodeID as an int32
	nodeID, err := strconv.Atoi(os.Args[1])
	if err != nil {
		log.Fatalf("Invalid nodeID: %v", err)
	}
	nextNode := os.Args[2]
	initialToken := os.Args[3] == "true"

	// Initialize Auction State
	auctionState := &AuctionState{
		HighestBid:    0,
		HighestBidder: "",
		IsEnded:       false,
		StartTime:     time.Now(),
		EndTime:       time.Now().Add(100 * time.Second), // 100-second auction
	}

	// Create a new TokenRingServer instance
	tokenRingServer := NewTokenRingServer(int32(nodeID), nextNode, initialToken, auctionState)

	// Create a new AuctionService instance
	auctionService := NewAuctionService(auctionState, tokenRingServer)

	// Setup logging to a file
	logFile, err := os.OpenFile("logs/log.txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	defer logFile.Close()
	log.SetOutput(logFile)
	log.Println("========================================")
	log.Printf("Node %d is starting. InitialToken: %v", tokenRingServer.nodeID, initialToken)

	// Start the auction end monitoring
	go func() {
		time.Sleep(time.Until(auctionState.EndTime))
		auctionState.mutex.Lock()
		auctionState.IsEnded = true
		auctionState.mutex.Unlock()
		log.Printf("[Auction] Auction ended. Winner: %s with bid %d", auctionState.HighestBidder, auctionState.HighestBid)

		// Automatically pass the token after auction ends
		tokenRingServer.mutex.Lock()
		if tokenRingServer.hasToken {
			tokenRingServer.hasToken = false
			go tokenRingServer.passToken()
		}
		tokenRingServer.mutex.Unlock()
	}()

	// Start the gRPC server
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", 5000+nodeID))
	if err != nil {
		log.Fatalf("Failed to listen on port: %v", err)
	}

	grpcServer := grpc.NewServer()

	// Register TokenRingServer
	pb.RegisterTokenRingServer(grpcServer, tokenRingServer)

	// Register AuctionServiceServer
	pb.RegisterAuctionServiceServer(grpcServer, auctionService)

	log.Printf("Node %d is listening on port %d\n", tokenRingServer.nodeID, 5000+nodeID)

	// Start monitoring token passing
	go tokenRingServer.moniterTokenPassing()

	// If the node has the initial token, start passing it after a brief delay
	if initialToken {
		go func() {
			time.Sleep(1 * time.Second) // Brief delay before starting token pass
			tokenRingServer.passToken()
		}()
	}

	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
