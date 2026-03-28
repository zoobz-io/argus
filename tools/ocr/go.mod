module github.com/zoobz-io/argus/tools/ocr

go 1.25.5

require google.golang.org/grpc v1.79.3

require (
	github.com/zoobz-io/argus/proto v0.0.0
	golang.org/x/net v0.48.0 // indirect
	golang.org/x/sys v0.39.0 // indirect
	golang.org/x/text v0.32.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20251202230838-ff82c1b0f217 // indirect
	google.golang.org/protobuf v1.36.11 // indirect
)

replace github.com/zoobz-io/argus/proto => ../../proto
