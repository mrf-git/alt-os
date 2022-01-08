module alt-os

go 1.16

replace alt-os => ./
replace oci => ../third-party/oci

require (
	oci v0.0.0
	github.com/gogo/protobuf v1.3.2
	google.golang.org/grpc v1.43.0
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
)
