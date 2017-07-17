# grpc-throttle
grpc-throttle interceptor for go-grpc-middleware

## Get

`$ go get github.com/yaronsumel/grpc-throttle`

## Usage

##### Import 

```go

    var sMap = throttle.SemaphoreMap{
        "/authpb.Auth/Method": make(throttle.Semaphore, 1),
    }
    
    func TFunc(fullMethod string) (throttle.Semaphore, bool) {
        if s, ok := sMap[fullMethod]; ok {
            return s, true
        }
        return nil, false
    }

	server := grpc.NewServer(
		grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(
					// keep it last in the interceptor chain
            throttle.StreamServerInterceptor
		)),
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			// keep it last in the interceptor chain
			throttle.UnaryServerInterceptor(TFunc),
		)),
	)

```