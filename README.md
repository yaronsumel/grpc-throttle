# grpc-throttle
grpc-throttle interceptor for go-grpc-middleware. inspired by [jbrandhorst](https://jbrandhorst.com/post/go-semaphore/)

## Get

`$ go get github.com/yaronsumel/grpc-throttle`

## Usage

Make SemaphoreMap with specific size per methods

```go
var sMap = throttle.SemaphoreMap{
    "/authpb.Auth/Method": make(throttle.Semaphore, 1),
}
```

Create ThrottleFunc which returns Semaphore for method.. or control it in any other way using the the context
```go

func ThrottleFunc(ctx context.Context,fullMethod string) (throttle.Semaphore, bool) {
    if s, ok := sMap[fullMethod]; ok {
        return s, true
    }
    return nil, false
}

```

Use it as interceptor

```go

server := grpc.NewServer(
    grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(
        // keep it last in the interceptor chain
        throttle.StreamServerInterceptor(ThrottleFunc)
    )),
    grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
        // keep it last in the interceptor chain
        throttle.UnaryServerInterceptor(ThrottleFunc),
    )),
)

```
