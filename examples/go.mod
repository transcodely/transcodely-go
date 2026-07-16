module github.com/transcodely/transcodely-go/examples

go 1.23

require (
	github.com/transcodely/transcodely-go v0.0.0
	google.golang.org/protobuf v1.36.5
)

require (
	buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go v1.36.5-20250912141014-52f32327d4b0.1 // indirect
	connectrpc.com/connect v1.18.1 // indirect
	github.com/google/uuid v1.6.0 // indirect
)

replace github.com/transcodely/transcodely-go => ..
