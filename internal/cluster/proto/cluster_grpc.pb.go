// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v6.30.0--rc1
// source: internal/cluster/proto/cluster.proto

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
	ElectionService_RequestVote_FullMethodName = "/cluster.ElectionService/RequestVote"
	ElectionService_Heartbeat_FullMethodName   = "/cluster.ElectionService/Heartbeat"
)

// ElectionServiceClient is the client API for ElectionService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
//
// ****************************************************************
//
//	ElectionService                         *
//
// ***************************************************************
type ElectionServiceClient interface {
	RequestVote(ctx context.Context, in *VoteRequest, opts ...grpc.CallOption) (*VoteResponse, error)
	Heartbeat(ctx context.Context, in *HeartbeatRequest, opts ...grpc.CallOption) (*HeartbeatResponse, error)
}

type electionServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewElectionServiceClient(cc grpc.ClientConnInterface) ElectionServiceClient {
	return &electionServiceClient{cc}
}

func (c *electionServiceClient) RequestVote(ctx context.Context, in *VoteRequest, opts ...grpc.CallOption) (*VoteResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(VoteResponse)
	err := c.cc.Invoke(ctx, ElectionService_RequestVote_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *electionServiceClient) Heartbeat(ctx context.Context, in *HeartbeatRequest, opts ...grpc.CallOption) (*HeartbeatResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(HeartbeatResponse)
	err := c.cc.Invoke(ctx, ElectionService_Heartbeat_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ElectionServiceServer is the server API for ElectionService service.
// All implementations must embed UnimplementedElectionServiceServer
// for forward compatibility.
//
// ****************************************************************
//
//	ElectionService                         *
//
// ***************************************************************
type ElectionServiceServer interface {
	RequestVote(context.Context, *VoteRequest) (*VoteResponse, error)
	Heartbeat(context.Context, *HeartbeatRequest) (*HeartbeatResponse, error)
	mustEmbedUnimplementedElectionServiceServer()
}

// UnimplementedElectionServiceServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedElectionServiceServer struct{}

func (UnimplementedElectionServiceServer) RequestVote(context.Context, *VoteRequest) (*VoteResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RequestVote not implemented")
}
func (UnimplementedElectionServiceServer) Heartbeat(context.Context, *HeartbeatRequest) (*HeartbeatResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Heartbeat not implemented")
}
func (UnimplementedElectionServiceServer) mustEmbedUnimplementedElectionServiceServer() {}
func (UnimplementedElectionServiceServer) testEmbeddedByValue()                         {}

// UnsafeElectionServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ElectionServiceServer will
// result in compilation errors.
type UnsafeElectionServiceServer interface {
	mustEmbedUnimplementedElectionServiceServer()
}

func RegisterElectionServiceServer(s grpc.ServiceRegistrar, srv ElectionServiceServer) {
	// If the following call pancis, it indicates UnimplementedElectionServiceServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&ElectionService_ServiceDesc, srv)
}

func _ElectionService_RequestVote_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(VoteRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ElectionServiceServer).RequestVote(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ElectionService_RequestVote_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ElectionServiceServer).RequestVote(ctx, req.(*VoteRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ElectionService_Heartbeat_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(HeartbeatRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ElectionServiceServer).Heartbeat(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ElectionService_Heartbeat_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ElectionServiceServer).Heartbeat(ctx, req.(*HeartbeatRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// ElectionService_ServiceDesc is the grpc.ServiceDesc for ElectionService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var ElectionService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "cluster.ElectionService",
	HandlerType: (*ElectionServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "RequestVote",
			Handler:    _ElectionService_RequestVote_Handler,
		},
		{
			MethodName: "Heartbeat",
			Handler:    _ElectionService_Heartbeat_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "internal/cluster/proto/cluster.proto",
}

const (
	ReplicationService_ForwardRequest_FullMethodName   = "/cluster.ReplicationService/ForwardRequest"
	ReplicationService_ReplicateRequest_FullMethodName = "/cluster.ReplicationService/ReplicateRequest"
	ReplicationService_SyncRequest_FullMethodName      = "/cluster.ReplicationService/SyncRequest"
)

// ReplicationServiceClient is the client API for ReplicationService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
//
// ****************************************************************
//
//	ReplicationService                        *
//
// ***************************************************************
type ReplicationServiceClient interface {
	ForwardRequest(ctx context.Context, in *CommandRequest, opts ...grpc.CallOption) (*CommandResponse, error)
	ReplicateRequest(ctx context.Context, in *Command, opts ...grpc.CallOption) (*ReplicationAck, error)
	SyncRequest(ctx context.Context, in *SyncRequestMessage, opts ...grpc.CallOption) (*SyncResponse, error)
}

type replicationServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewReplicationServiceClient(cc grpc.ClientConnInterface) ReplicationServiceClient {
	return &replicationServiceClient{cc}
}

func (c *replicationServiceClient) ForwardRequest(ctx context.Context, in *CommandRequest, opts ...grpc.CallOption) (*CommandResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(CommandResponse)
	err := c.cc.Invoke(ctx, ReplicationService_ForwardRequest_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *replicationServiceClient) ReplicateRequest(ctx context.Context, in *Command, opts ...grpc.CallOption) (*ReplicationAck, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(ReplicationAck)
	err := c.cc.Invoke(ctx, ReplicationService_ReplicateRequest_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *replicationServiceClient) SyncRequest(ctx context.Context, in *SyncRequestMessage, opts ...grpc.CallOption) (*SyncResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(SyncResponse)
	err := c.cc.Invoke(ctx, ReplicationService_SyncRequest_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ReplicationServiceServer is the server API for ReplicationService service.
// All implementations must embed UnimplementedReplicationServiceServer
// for forward compatibility.
//
// ****************************************************************
//
//	ReplicationService                        *
//
// ***************************************************************
type ReplicationServiceServer interface {
	ForwardRequest(context.Context, *CommandRequest) (*CommandResponse, error)
	ReplicateRequest(context.Context, *Command) (*ReplicationAck, error)
	SyncRequest(context.Context, *SyncRequestMessage) (*SyncResponse, error)
	mustEmbedUnimplementedReplicationServiceServer()
}

// UnimplementedReplicationServiceServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedReplicationServiceServer struct{}

func (UnimplementedReplicationServiceServer) ForwardRequest(context.Context, *CommandRequest) (*CommandResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ForwardRequest not implemented")
}
func (UnimplementedReplicationServiceServer) ReplicateRequest(context.Context, *Command) (*ReplicationAck, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ReplicateRequest not implemented")
}
func (UnimplementedReplicationServiceServer) SyncRequest(context.Context, *SyncRequestMessage) (*SyncResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SyncRequest not implemented")
}
func (UnimplementedReplicationServiceServer) mustEmbedUnimplementedReplicationServiceServer() {}
func (UnimplementedReplicationServiceServer) testEmbeddedByValue()                            {}

// UnsafeReplicationServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ReplicationServiceServer will
// result in compilation errors.
type UnsafeReplicationServiceServer interface {
	mustEmbedUnimplementedReplicationServiceServer()
}

func RegisterReplicationServiceServer(s grpc.ServiceRegistrar, srv ReplicationServiceServer) {
	// If the following call pancis, it indicates UnimplementedReplicationServiceServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&ReplicationService_ServiceDesc, srv)
}

func _ReplicationService_ForwardRequest_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CommandRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ReplicationServiceServer).ForwardRequest(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ReplicationService_ForwardRequest_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ReplicationServiceServer).ForwardRequest(ctx, req.(*CommandRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ReplicationService_ReplicateRequest_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Command)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ReplicationServiceServer).ReplicateRequest(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ReplicationService_ReplicateRequest_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ReplicationServiceServer).ReplicateRequest(ctx, req.(*Command))
	}
	return interceptor(ctx, in, info, handler)
}

func _ReplicationService_SyncRequest_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SyncRequestMessage)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ReplicationServiceServer).SyncRequest(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ReplicationService_SyncRequest_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ReplicationServiceServer).SyncRequest(ctx, req.(*SyncRequestMessage))
	}
	return interceptor(ctx, in, info, handler)
}

// ReplicationService_ServiceDesc is the grpc.ServiceDesc for ReplicationService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var ReplicationService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "cluster.ReplicationService",
	HandlerType: (*ReplicationServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "ForwardRequest",
			Handler:    _ReplicationService_ForwardRequest_Handler,
		},
		{
			MethodName: "ReplicateRequest",
			Handler:    _ReplicationService_ReplicateRequest_Handler,
		},
		{
			MethodName: "SyncRequest",
			Handler:    _ReplicationService_SyncRequest_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "internal/cluster/proto/cluster.proto",
}
