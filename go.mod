module github.com/dodo-cli/dodo-stage-docker-virtualbox

go 1.16

replace (
	github.com/hashicorp/go-immutable-radix => github.com/tonistiigi/go-immutable-radix v0.0.0-20170803185627-826af9ccf0fe
	github.com/jaguilar/vt100 => github.com/tonistiigi/vt100 v0.0.0-20190402012908-ad4c4a574305
)

require (
	github.com/docker/docker v20.10.2+incompatible
	github.com/dodo-cli/dodo-buildkit v0.1.1-0.20211104120639-386b829ef813 // indirect
	github.com/dodo-cli/dodo-core v0.2.1-0.20211108131047-2c0cd4c202d2
	github.com/dodo-cli/dodo-docker v0.1.1-0.20211104120605-cea72844a81b
	github.com/dodo-cli/dodo-stage v0.0.0-20211108150615-f2557d055213
	github.com/hashicorp/go-hclog v0.15.0
	github.com/oclaussen/go-gimme/ssh v0.0.0-20200205175519-d9560e60c720
	github.com/pkg/errors v0.9.1
)
