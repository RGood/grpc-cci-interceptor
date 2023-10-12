package wrapper

import (
	"context"

	"google.golang.org/grpc"
)

// ConnectionInterceptor is the function that gets called with call options whenever a new connection is opened
type ConnectionInterceptor func(ctx context.Context, method string, opts []grpc.CallOption, next func(context.Context) error) error

// Wrap accepts a grpc.ClientConnInterface and interceptor and calls the interceptor whenever a connection is made
func Wrap(cci grpc.ClientConnInterface, interceptor ConnectionInterceptor) grpc.ClientConnInterface {
	return &wrappedCCI{
		cci:         cci,
		interceptor: interceptor,
	}
}

type wrappedCCI struct {
	cci         grpc.ClientConnInterface
	interceptor ConnectionInterceptor
}

type wrappedStream struct {
	grpc.ClientStream
}

var _ grpc.ClientConnInterface = (*wrappedCCI)(nil)

// Invoke implements grpc.ClientConnInterface.
func (wrapper *wrappedCCI) Invoke(ctx context.Context, method string, args any, reply any, opts ...grpc.CallOption) error {
	return wrapper.interceptor(ctx, method, opts, func(ctx context.Context) error {
		return wrapper.cci.Invoke(ctx, method, args, reply, opts...)
	})
}

// NewStream implements grpc.ClientConnInterface.
func (wrapper *wrappedCCI) NewStream(ctx context.Context, desc *grpc.StreamDesc, method string, opts ...grpc.CallOption) (stream grpc.ClientStream, err error) {
	// wg := sync.WaitGroup{}
	// go func() {
	// 	calledNext := make(chan struct{})
	// 	err = wrapper.interceptor(ctx, method, opts, func(ctx context.Context) error {
	// 		close(calledNext)
	// 		innerStream, err := wrapper.cci.NewStream(ctx, desc, method, opts...)
	// 		stream = innerStream
	// 		if err != nil {
	// 			defer wg.Done()
	// 			return err
	// 		}
	// 		wg.Done()

	// 		<-stream.Context().Done()

	// 		return nil
	// 	})
	// 	select {
	// 	case <-calledNext:
	// 	default:
	// 		close(calledNext)
	// 		wg.Done()
	// 	}
	// }()
	// wg.Wait()
	// if err != nil {
	// 	return nil, err
	// }
	stream, err = wrapper.cci.NewStream(ctx, desc, method, opts...)
	if err != nil {
		return nil, err
	}
	return &wrappedStream{
		ClientStream: stream,
	}, nil
}
