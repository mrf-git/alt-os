module alt-os-parent

go 1.16

replace alt-os => ./pkg

replace oci => ./third-party/oci

require (
	alt-os v0.0.0
	github.com/gogo/googleapis v1.4.1
	github.com/gogo/protobuf v1.3.2
	golang.org/x/net v0.0.0-20220105145211-5b0dc2dfae98 // indirect
	golang.org/x/sys v0.0.0-20211216021012-1d35b9e2eb4e // indirect
	golang.org/x/text v0.3.7 // indirect
	google.golang.org/grpc v1.43.0
	github.com/sirupsen/logrus v1.8.1
)
