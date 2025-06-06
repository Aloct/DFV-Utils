// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v6.30.2
// source: keymanager.proto

package keyManagerProto

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
	KeyManager_DecryptKEKAndGetReference_FullMethodName = "/keymanager.KeyManager/DecryptKEKAndGetReference"
	KeyManager_RegisterKEK_FullMethodName               = "/keymanager.KeyManager/RegisterKEK"
	KeyManager_RegisterDEK_FullMethodName               = "/keymanager.KeyManager/RegisterDEK"
)

// KeyManagerClient is the client API for KeyManager service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type KeyManagerClient interface {
	DecryptKEKAndGetReference(ctx context.Context, in *DEKGetter, opts ...grpc.CallOption) (*DEKIdentAndKEK, error)
	RegisterKEK(ctx context.Context, opts ...grpc.CallOption) (grpc.BidiStreamingClient[KEKAndDefaultDEKs, RegisterResponse], error)
	RegisterDEK(ctx context.Context, in *DEKRegistration, opts ...grpc.CallOption) (*DekRegisterAndKEK, error)
}

type keyManagerClient struct {
	cc grpc.ClientConnInterface
}

func NewKeyManagerClient(cc grpc.ClientConnInterface) KeyManagerClient {
	return &keyManagerClient{cc}
}

func (c *keyManagerClient) DecryptKEKAndGetReference(ctx context.Context, in *DEKGetter, opts ...grpc.CallOption) (*DEKIdentAndKEK, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(DEKIdentAndKEK)
	err := c.cc.Invoke(ctx, KeyManager_DecryptKEKAndGetReference_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *keyManagerClient) RegisterKEK(ctx context.Context, opts ...grpc.CallOption) (grpc.BidiStreamingClient[KEKAndDefaultDEKs, RegisterResponse], error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	stream, err := c.cc.NewStream(ctx, &KeyManager_ServiceDesc.Streams[0], KeyManager_RegisterKEK_FullMethodName, cOpts...)
	if err != nil {
		return nil, err
	}
	x := &grpc.GenericClientStream[KEKAndDefaultDEKs, RegisterResponse]{ClientStream: stream}
	return x, nil
}

// This type alias is provided for backwards compatibility with existing code that references the prior non-generic stream type by name.
type KeyManager_RegisterKEKClient = grpc.BidiStreamingClient[KEKAndDefaultDEKs, RegisterResponse]

func (c *keyManagerClient) RegisterDEK(ctx context.Context, in *DEKRegistration, opts ...grpc.CallOption) (*DekRegisterAndKEK, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(DekRegisterAndKEK)
	err := c.cc.Invoke(ctx, KeyManager_RegisterDEK_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// KeyManagerServer is the server API for KeyManager service.
// All implementations must embed UnimplementedKeyManagerServer
// for forward compatibility.
type KeyManagerServer interface {
	DecryptKEKAndGetReference(context.Context, *DEKGetter) (*DEKIdentAndKEK, error)
	RegisterKEK(grpc.BidiStreamingServer[KEKAndDefaultDEKs, RegisterResponse]) error
	RegisterDEK(context.Context, *DEKRegistration) (*DekRegisterAndKEK, error)
	mustEmbedUnimplementedKeyManagerServer()
}

// UnimplementedKeyManagerServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedKeyManagerServer struct{}

func (UnimplementedKeyManagerServer) DecryptKEKAndGetReference(context.Context, *DEKGetter) (*DEKIdentAndKEK, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DecryptKEKAndGetReference not implemented")
}
func (UnimplementedKeyManagerServer) RegisterKEK(grpc.BidiStreamingServer[KEKAndDefaultDEKs, RegisterResponse]) error {
	return status.Errorf(codes.Unimplemented, "method RegisterKEK not implemented")
}
func (UnimplementedKeyManagerServer) RegisterDEK(context.Context, *DEKRegistration) (*DekRegisterAndKEK, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RegisterDEK not implemented")
}
func (UnimplementedKeyManagerServer) mustEmbedUnimplementedKeyManagerServer() {}
func (UnimplementedKeyManagerServer) testEmbeddedByValue()                    {}

// UnsafeKeyManagerServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to KeyManagerServer will
// result in compilation errors.
type UnsafeKeyManagerServer interface {
	mustEmbedUnimplementedKeyManagerServer()
}

func RegisterKeyManagerServer(s grpc.ServiceRegistrar, srv KeyManagerServer) {
	// If the following call pancis, it indicates UnimplementedKeyManagerServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&KeyManager_ServiceDesc, srv)
}

func _KeyManager_DecryptKEKAndGetReference_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DEKGetter)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(KeyManagerServer).DecryptKEKAndGetReference(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: KeyManager_DecryptKEKAndGetReference_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(KeyManagerServer).DecryptKEKAndGetReference(ctx, req.(*DEKGetter))
	}
	return interceptor(ctx, in, info, handler)
}

func _KeyManager_RegisterKEK_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(KeyManagerServer).RegisterKEK(&grpc.GenericServerStream[KEKAndDefaultDEKs, RegisterResponse]{ServerStream: stream})
}

// This type alias is provided for backwards compatibility with existing code that references the prior non-generic stream type by name.
type KeyManager_RegisterKEKServer = grpc.BidiStreamingServer[KEKAndDefaultDEKs, RegisterResponse]

func _KeyManager_RegisterDEK_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DEKRegistration)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(KeyManagerServer).RegisterDEK(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: KeyManager_RegisterDEK_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(KeyManagerServer).RegisterDEK(ctx, req.(*DEKRegistration))
	}
	return interceptor(ctx, in, info, handler)
}

// KeyManager_ServiceDesc is the grpc.ServiceDesc for KeyManager service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var KeyManager_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "keymanager.KeyManager",
	HandlerType: (*KeyManagerServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "DecryptKEKAndGetReference",
			Handler:    _KeyManager_DecryptKEKAndGetReference_Handler,
		},
		{
			MethodName: "RegisterDEK",
			Handler:    _KeyManager_RegisterDEK_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "RegisterKEK",
			Handler:       _KeyManager_RegisterKEK_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
	},
	Metadata: "keymanager.proto",
}
