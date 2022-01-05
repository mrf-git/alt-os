module alt-os-parent

go 1.16

replace (
	alt-os => ./pkg
)

require (
	alt-os v0.0.0
	github.com/gogo/protobuf v1.3.2
	github.com/gogo/googleapis v1.4.1
	google.golang.org/grpc v1.43.0
)
