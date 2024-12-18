// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v5.28.3
// source: proto/auction.proto

package proto

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.64.0 or later.
const _ = grpc.SupportPackageIsVersion9

const (
	TokenRing_PassToken_FullMethodName   = "/auction.TokenRing/PassToken"
	TokenRing_GetNextNode_FullMethodName = "/auction.TokenRing/GetNextNode"
)

// TokenRingClient is the client API for TokenRing service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
//
// TokenRing service for token passing between nodes
type TokenRingClient interface {
	// PassToken is used by a node to pass the token to the next node in the ring
	PassToken(ctx context.Context, in *TokenMessage, opts ...grpc.CallOption) (*Response, error)
	// GetNextNode is used to get the next node in the ring
	GetNextNode(ctx context.Context, in *GetNextNodeRequest, opts ...grpc.CallOption) (*GetNextNodeResponse, error)
}

type tokenRingClient struct {
	cc grpc.ClientConnInterface
}

func NewTokenRingClient(cc grpc.ClientConnInterface) TokenRingClient {
	return &tokenRingClient{cc}
}

func (c *tokenRingClient) PassToken(ctx context.Context, in *TokenMessage, opts ...grpc.CallOption) (*Response, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(Response)
	err := c.cc.Invoke(ctx, TokenRing_PassToken_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *tokenRingClient) GetNextNode(ctx context.Context, in *GetNextNodeRequest, opts ...grpc.CallOption) (*GetNextNodeResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(GetNextNodeResponse)
	err := c.cc.Invoke(ctx, TokenRing_GetNextNode_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// TokenRingServer is the server API for TokenRing service.
// All implementations must embed UnimplementedTokenRingServer
// for forward compatibility.
//
// TokenRing service for token passing between nodes
type TokenRingServer interface {
	// PassToken is used by a node to pass the token to the next node in the ring
	PassToken(context.Context, *TokenMessage) (*Response, error)
	// GetNextNode is used to get the next node in the ring
	GetNextNode(context.Context, *GetNextNodeRequest) (*GetNextNodeResponse, error)
	mustEmbedUnimplementedTokenRingServer()
}

// UnimplementedTokenRingServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedTokenRingServer struct{}

func (UnimplementedTokenRingServer) PassToken(context.Context, *TokenMessage) (*Response, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PassToken not implemented")
}
func (UnimplementedTokenRingServer) GetNextNode(context.Context, *GetNextNodeRequest) (*GetNextNodeResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetNextNode not implemented")
}
func (UnimplementedTokenRingServer) mustEmbedUnimplementedTokenRingServer() {}
func (UnimplementedTokenRingServer) testEmbeddedByValue()                   {}

// UnsafeTokenRingServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to TokenRingServer will
// result in compilation errors.
type UnsafeTokenRingServer interface {
	mustEmbedUnimplementedTokenRingServer()
}

func RegisterTokenRingServer(s grpc.ServiceRegistrar, srv TokenRingServer) {
	// If the following call pancis, it indicates UnimplementedTokenRingServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&TokenRing_ServiceDesc, srv)
}

func _TokenRing_PassToken_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(TokenMessage)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TokenRingServer).PassToken(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: TokenRing_PassToken_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TokenRingServer).PassToken(ctx, req.(*TokenMessage))
	}
	return interceptor(ctx, in, info, handler)
}

func _TokenRing_GetNextNode_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetNextNodeRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TokenRingServer).GetNextNode(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: TokenRing_GetNextNode_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TokenRingServer).GetNextNode(ctx, req.(*GetNextNodeRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// TokenRing_ServiceDesc is the grpc.ServiceDesc for TokenRing service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var TokenRing_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "auction.TokenRing",
	HandlerType: (*TokenRingServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "PassToken",
			Handler:    _TokenRing_PassToken_Handler,
		},
		{
			MethodName: "GetNextNode",
			Handler:    _TokenRing_GetNextNode_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "proto/auction.proto",
}

const (
	AuctionService_Bid_FullMethodName    = "/auction.AuctionService/Bid"
	AuctionService_Result_FullMethodName = "/auction.AuctionService/Result"
)

// AuctionServiceClient is the client API for AuctionService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
//
// AuctionService for handling bids and querying results
type AuctionServiceClient interface {
	// Bid allows a client to place a bid in the auction
	Bid(ctx context.Context, in *BidRequest, opts ...grpc.CallOption) (*BidResponse, error)
	// Result allows a client to query the current highest bid or the auction winner
	Result(ctx context.Context, in *ResultRequest, opts ...grpc.CallOption) (*ResultResponse, error)
}

type auctionServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewAuctionServiceClient(cc grpc.ClientConnInterface) AuctionServiceClient {
	return &auctionServiceClient{cc}
}

func (c *auctionServiceClient) Bid(ctx context.Context, in *BidRequest, opts ...grpc.CallOption) (*BidResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(BidResponse)
	err := c.cc.Invoke(ctx, AuctionService_Bid_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *auctionServiceClient) Result(ctx context.Context, in *ResultRequest, opts ...grpc.CallOption) (*ResultResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(ResultResponse)
	err := c.cc.Invoke(ctx, AuctionService_Result_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// AuctionServiceServer is the server API for AuctionService service.
// All implementations must embed UnimplementedAuctionServiceServer
// for forward compatibility.
//
// AuctionService for handling bids and querying results
type AuctionServiceServer interface {
	// Bid allows a client to place a bid in the auction
	Bid(context.Context, *BidRequest) (*BidResponse, error)
	// Result allows a client to query the current highest bid or the auction winner
	Result(context.Context, *ResultRequest) (*ResultResponse, error)
	mustEmbedUnimplementedAuctionServiceServer()
}

// UnimplementedAuctionServiceServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedAuctionServiceServer struct{}

func (UnimplementedAuctionServiceServer) Bid(context.Context, *BidRequest) (*BidResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Bid not implemented")
}
func (UnimplementedAuctionServiceServer) Result(context.Context, *ResultRequest) (*ResultResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Result not implemented")
}
func (UnimplementedAuctionServiceServer) mustEmbedUnimplementedAuctionServiceServer() {}
func (UnimplementedAuctionServiceServer) testEmbeddedByValue()                        {}

// UnsafeAuctionServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to AuctionServiceServer will
// result in compilation errors.
type UnsafeAuctionServiceServer interface {
	mustEmbedUnimplementedAuctionServiceServer()
}

func RegisterAuctionServiceServer(s grpc.ServiceRegistrar, srv AuctionServiceServer) {
	// If the following call pancis, it indicates UnimplementedAuctionServiceServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&AuctionService_ServiceDesc, srv)
}

func _AuctionService_Bid_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(BidRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AuctionServiceServer).Bid(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: AuctionService_Bid_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AuctionServiceServer).Bid(ctx, req.(*BidRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _AuctionService_Result_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ResultRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AuctionServiceServer).Result(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: AuctionService_Result_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AuctionServiceServer).Result(ctx, req.(*ResultRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// AuctionService_ServiceDesc is the grpc.ServiceDesc for AuctionService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var AuctionService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "auction.AuctionService",
	HandlerType: (*AuctionServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Bid",
			Handler:    _AuctionService_Bid_Handler,
		},
		{
			MethodName: "Result",
			Handler:    _AuctionService_Result_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "proto/auction.proto",
}
