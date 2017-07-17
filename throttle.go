//	MIT License
//
//	Copyright (c) 2017 Yaron Sumel
//
//	Permission is hereby granted, free of charge, to any person obtaining a copy
//	of this software and associated documentation files (the "Software"), to deal
//	in the Software without restriction, including without limitation the rights
//	to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
//	copies of the Software, and to permit persons to whom the Software is
//	furnished to do so, subject to the following conditions:
//
//	The above copyright notice and this permission notice shall be included in all
//	copies or substantial portions of the Software.
//
//	THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
//	IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
//	FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
//	AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
//	LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
//	OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
//	SOFTWARE.

package throttle

import (
	"github.com/grpc-ecosystem/go-grpc-middleware"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

// ThrottleFunc will return Semaphore for each fullMethod string
type ThrottleFunc func(fullMethod string) (Semaphore, bool)

// SemaphoreMap its is map of FullMethod and Semaphore
type SemaphoreMap map[string]chan struct{}

// Semaphore chan
type Semaphore chan struct{}

// ReleaseSlot release Semaphore
func (s Semaphore) ReleaseSlot() {
	// Read to release a slot
	<-s
}

// WaitForSlotAvailable wait for Available Semaphore
func (s Semaphore) WaitForSlotAvailable(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
		// Blocks while channel is full
	case s <- struct{}{}:
	}
	return nil
}

// UnaryServerInterceptor returns a new unary server interceptors that performs per-request throttling.
func UnaryServerInterceptor(fn ThrottleFunc) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		semaphore, ok := fn(info.FullMethod)
		if !ok {
			goto next
		}
		// wait for available slot
		if err := semaphore.WaitForSlotAvailable(ctx); err != nil {
			return nil, err
		}
		defer semaphore.ReleaseSlot()
		//
	next:
		return handler(ctx, req)
	}
}

// StreamServerInterceptor returns a new unary server interceptors that performs per-request throttling.
func StreamServerInterceptor(fn ThrottleFunc) grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		ctx := stream.Context()
		semaphore, ok := fn(info.FullMethod)
		if !ok {
			goto next
		}
		// wait for available slot
		if err := semaphore.WaitForSlotAvailable(ctx); err != nil {
			return err
		}
		defer semaphore.ReleaseSlot()
	next:
		wrapped := grpc_middleware.WrapServerStream(stream)
		wrapped.WrappedContext = ctx
		return handler(srv, wrapped)
	}
}
