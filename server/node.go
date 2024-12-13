package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
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
	peerNodes    []string // list of peer node addresses in the ring order
	currentIndex int      // Index of the next node to pass the token to
	auctionState *AuctionState
	mutex        sync.Mutex
	lastPass     time.Time
	timeout      time.Duration
	bidChannel   chan *pb.BidRequest // Channel for incoming bids
	pendingBids  []*pb.BidRequest    // Slice to store pending bids
}

// AuctionServiceServer handles auction operations
type AuctionServiceServer struct {
	pb.UnimplementedAuctionServiceServer
	auctionState *AuctionState
	node         *TokenRingServer
}

// NewTokenRingServer creates a new TokenRingServer instance
func NewTokenRingServer(nodeID int32, peerNodes []string, initialToken bool, auctionState *AuctionState) *TokenRingServer {
	return &TokenRingServer{
		nodeID:       nodeID,
		hasToken:     initialToken,
		peerNodes:    peerNodes,
		currentIndex: 0, // Starts with the first peer
		auctionState: auctionState,
		lastPass:     time.Now(),
		timeout:      10 * time.Second,               //  timeout duration for token passing
		bidChannel:   make(chan *pb.BidRequest, 100), // Buffered channel for bids
	}
}

// NewAuctionService creates a new AuctionService instance
func NewAuctionService(auctionState *AuctionState, node *TokenRingServer) *AuctionServiceServer {
	return &AuctionServiceServer{
		auctionState: auctionState,
		node:         node,
	}
}

// PassToken receives the token from the previous node in the ring
func (s *TokenRingServer) PassToken(ctx context.Context, tokenMsg *pb.TokenMessage) (*pb.Response, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	log.Printf("Node %d received token from Node %d", s.nodeID, tokenMsg.TokenHolderId)

	s.hasToken = true
	s.lastPass = time.Now()

	// Process any pending bids first
	for _, bid := range s.pendingBids {
		s.processBid(bid)
		go s.passToken(bid.BidderId, bid.Amount)
	}
	// Clear pending bids after processing
	s.pendingBids = nil

	// If there's a bid in the token, process it
	if tokenMsg.Amount > 0 && tokenMsg.BidderId != "" {
		s.processBid(&pb.BidRequest{
			BidderId: tokenMsg.BidderId,
			Amount:   tokenMsg.Amount,
		})
	}

	// Check if auction has ended
	s.auctionState.mutex.Lock()
	auctionEnded := s.auctionState.IsEnded
	s.auctionState.mutex.Unlock()

	if auctionEnded {
		log.Printf("Node %d detected auction end. Token will no longer be passed.", s.nodeID)
		s.hasToken = false
		return &pb.Response{Message: "Auction ended. Token not passed."}, nil
	}

	// After processing, pass the token to the next node
	s.hasToken = false
	go s.passToken(tokenMsg.BidderId, tokenMsg.Amount)

	return &pb.Response{Message: "Token passed"}, nil
}

// passToken sends the token to the next node in the ring
func (s *TokenRingServer) passToken(bidderID string, amount int32) {
	for {
		s.mutex.Lock()
		if len(s.peerNodes) == 0 {
			log.Printf("Node %d has no peers to pass the token to.", s.nodeID)
			s.mutex.Unlock()
			return
		}

		if s.currentIndex >= len(s.peerNodes) {
			s.currentIndex = 0 // Wrap around the ring
		}

		nextNodeAddr := s.peerNodes[s.currentIndex]
		s.currentIndex++
		s.mutex.Unlock()

		if nextNodeAddr == "" {
			log.Printf("Node %d: Next node address is empty. Skipping.", s.nodeID)
			continue
		}

		conn, err := grpc.Dial(nextNodeAddr, grpc.WithInsecure(), grpc.WithBlock(), grpc.WithTimeout(5*time.Second))
		if err != nil {
			log.Printf("Node %d failed to connect to next node %s: %v", s.nodeID, nextNodeAddr, err)
			// Attempt to get the next node from the failed node
			nextNodeAddrNew, err := s.getNextNodeAddr(nextNodeAddr)
			if err != nil || nextNodeAddrNew == "" {
				log.Printf("Node %d could not retrieve next node after failure. Skipping.", s.nodeID)
				continue
			}
			// Remove the failed node from the peer list
			s.mutex.Lock()
			index := -1
			for i, addr := range s.peerNodes {
				if addr == nextNodeAddr {
					index = i
					break
				}
			}
			if index != -1 {
				s.peerNodes = append(s.peerNodes[:index], s.peerNodes[index+1:]...)
				if s.currentIndex > index {
					s.currentIndex-- // Adjust index after removal
				}
				log.Printf("Node %d updated peer nodes, skipping failed node %s.", s.nodeID, nextNodeAddr)
			}
			s.mutex.Unlock()
			continue
		}

		client := pb.NewTokenRingClient(conn)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		token := &pb.TokenMessage{
			TokenHolderId: s.nodeID,
			BidderId:      bidderID,
			Amount:        amount,
		}

		// Pass the current node's ID as TokenHolderId
		_, err = client.PassToken(ctx, token)
		if err != nil {
			log.Printf("Node %d failed to pass token to %s: %v", s.nodeID, nextNodeAddr, err)
			conn.Close()
			// Attempt to get the next node from the failed node
			nextNodeAddrNew, err := s.getNextNodeAddr(nextNodeAddr)
			if err != nil || nextNodeAddrNew == "" {
				log.Printf("Node %d could not retrieve next node after failure. Skipping.", s.nodeID)
				continue
			}
			// Remove the failed node from the peer list
			s.mutex.Lock()
			index := -1
			for i, addr := range s.peerNodes {
				if addr == nextNodeAddr {
					index = i
					break
				}
			}
			if index != -1 {
				s.peerNodes = append(s.peerNodes[:index], s.peerNodes[index+1:]...)
				if s.currentIndex > index {
					s.currentIndex--
				}
				log.Printf("Node %d updated peer nodes, skipping failed node %s.", s.nodeID, nextNodeAddr)
			}
			s.mutex.Unlock()
			continue
		}

		// Log token passing and print that the node passed the token to the next node
		log.Printf("Node %d passed the token to node at %s with bid (%s, %d)", s.nodeID, nextNodeAddr, bidderID, amount)
		conn.Close()
		break
	}
}

// getNextNodeAddr connects to the failed node and retrieves its next node address via RPC.
func (s *TokenRingServer) getNextNodeAddr(failedNodeAddr string) (string, error) {
	conn, err := grpc.Dial(failedNodeAddr, grpc.WithInsecure(), grpc.WithBlock(), grpc.WithTimeout(5*time.Second))
	if err != nil {
		return "", fmt.Errorf("failed to connect to node %s: %v", failedNodeAddr, err)
	}
	defer conn.Close()

	client := pb.NewTokenRingClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := &pb.GetNextNodeRequest{CurrentNodeAddr: failedNodeAddr}
	resp, err := client.GetNextNode(ctx, req)
	if err != nil {
		return "", fmt.Errorf("GetNextNode RPC failed for node %s: %v", failedNodeAddr, err)
	}

	return resp.NextNodeAddr, nil
}

// GetNextNode asks the current next node for its next node
func (s *TokenRingServer) GetNextNode(ctx context.Context, req *pb.GetNextNodeRequest) (*pb.GetNextNodeResponse, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Find the index of the current node address
	var currentIndex int = -1
	for i, addr := range s.peerNodes {
		if addr == req.CurrentNodeAddr {
			currentIndex = i
			break
		}
	}

	if currentIndex == -1 {
		return &pb.GetNextNodeResponse{NextNodeAddr: ""}, fmt.Errorf("Current node address not found")
	}

	// Calculate the next node index
	nextIndex := (currentIndex + 1) % len(s.peerNodes)
	nextNodeAddr := s.peerNodes[nextIndex]

	return &pb.GetNextNodeResponse{NextNodeAddr: nextNodeAddr}, nil
}

func (s *TokenRingServer) processBid(bid *pb.BidRequest) {
	s.auctionState.mutex.Lock()
	defer s.auctionState.mutex.Unlock()

	if s.auctionState.IsEnded {
		log.Printf("Auction has ended. Bid from %s with amount %d is rejected.", bid.BidderId, bid.Amount)
		return
	}

	if bid.Amount <= s.auctionState.HighestBid {
		log.Printf("Bidder %s's bid of %d is too low. Current highest bid is %d.", bid.BidderId, bid.Amount, s.auctionState.HighestBid)
		return
	}

	// Update auction state with the new highest bid
	s.auctionState.HighestBid = bid.Amount
	s.auctionState.HighestBidder = bid.BidderId

	log.Printf("New highest bid: %d by %s", s.auctionState.HighestBid, s.auctionState.HighestBidder)
}

// continously listens for incoming bids and processes them
func (s *TokenRingServer) listenForBids() {
	for bid := range s.bidChannel {
		log.Printf("Node %d received a bid from %s with amount %d.", s.nodeID, bid.BidderId, bid.Amount)
		// If the node has the token, process the bid immediately
		if s.hasToken {
			log.Printf("Node %d has the token. Processing the bid", s.nodeID)
			s.processBid(bid)
			// Pass the token with the bid to propagate it.
			go s.passToken(bid.BidderId, bid.Amount)
		} else {
			// Otherwise, enqueue the bid to be processed when the token arrives
			s.mutex.Lock()
			s.pendingBids = append(s.pendingBids, bid)
			s.mutex.Unlock()
			log.Printf("Node %d added the bid from %s to pending bids.", s.nodeID, bid.BidderId)
		}
	}
	log.Printf("Node %d bidChannel closed. Stopping bid listener.", s.nodeID)
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

// Bid handles bid requests from clients
func (a *AuctionServiceServer) Bid(ctx context.Context, req *pb.BidRequest) (*pb.BidResponse, error) {
	log.Printf("Node %d received a bid from %s with amount %d.", a.node.nodeID, req.BidderId, req.Amount)

	// Enqueue the bid for processing
	select {
	case a.node.bidChannel <- req:
		return &pb.BidResponse{Outcome: "success: bid enqueued"}, nil
	default:
		log.Printf("Node %d's bid channel is full. Bid from %s with amount %d is rejected.", a.node.nodeID, req.BidderId, req.Amount)
		return &pb.BidResponse{Outcome: "fail: bid channel full"}, nil
	}
}

func main() {
	if len(os.Args) < 4 {
		log.Fatalf("Usage: go run node.go <nodeID> <peerNodes(comma-separated)> <initialToken(true|false)>")
	}

	// Parse the nodeID as an int32
	nodeID, err := strconv.Atoi(os.Args[1])
	if err != nil {
		log.Fatalf("Invalid nodeID: %v", err)
	}
	peerNodesInput := os.Args[2]
	initialToken := os.Args[3] == "true"

	// Split peer addresses by comma
	var peerNodes []string
	if peerNodesInput != "-" {
		peerNodes = strings.Split(peerNodesInput, ",")
	} else {
		peerNodes = []string{}
	}

	// Initialize Auction State
	auctionState := &AuctionState{
		HighestBid:    0,
		HighestBidder: "",
		IsEnded:       false,
		StartTime:     time.Now(),
		EndTime:       time.Now().Add(100 * time.Second), // 100 second auction
	}

	// Create a new TokenRingServer instance
	tokenRingServer := NewTokenRingServer(int32(nodeID), peerNodes, initialToken, auctionState)

	// Create a new AuctionService instance
	auctionService := NewAuctionService(auctionState, tokenRingServer)

	// Ensure logs directory exists
	if _, err := os.Stat("logs"); os.IsNotExist(err) {
		err := os.Mkdir("logs", 0755)
		if err != nil {
			log.Fatalf("Failed to create logs directory: %v", err)
		}
	}

	// Setup logging to a file
	logFileName := fmt.Sprintf("logs/node_%d_log.txt", nodeID)
	logFile, err := os.OpenFile(logFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	defer logFile.Close()
	log.SetOutput(logFile)
	log.Println("========================================")
	log.Printf("Node %d is starting. InitialToken: %v", tokenRingServer.nodeID, initialToken)

	go tokenRingServer.listenForBids()

	// Start the auction end monitoring
	go func() {
		time.Sleep(time.Until(auctionState.EndTime))
		auctionState.mutex.Lock()
		auctionState.IsEnded = true
		auctionState.mutex.Unlock()
		close(tokenRingServer.bidChannel)
		log.Printf("Auction ended. Winner: %s with bid %d", auctionState.HighestBidder, auctionState.HighestBid)

		// Automatically pass the token after auction ends
		tokenRingServer.mutex.Lock()
		if tokenRingServer.hasToken {
			tokenRingServer.hasToken = false
			go tokenRingServer.passToken("", 0)
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
	go tokenRingServer.monitorTokenPassing()

	// If the node has the initial token, start passing it after a brief delay
	if initialToken {
		go func() {
			time.Sleep(10 * time.Second) // Increased delay before starting token pass
			log.Printf("Node %d is initiating token passing.", tokenRingServer.nodeID)
			tokenRingServer.passToken("", 0)
		}()
	}

	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}

}

// monitorTokenPassing monitors token passing and regenerates token if timeout occurs
func (s *TokenRingServer) monitorTokenPassing() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		<-ticker.C
		s.mutex.Lock()
		if s.hasToken && time.Since(s.lastPass) > s.timeout {
			log.Printf("Node %d detected token pass timeout. Attempting to regenerate token", s.nodeID)
			s.hasToken = false
			go s.passToken("", 0)
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
