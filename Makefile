BUILD_VERSION = ''
LDFLAGS = -s -w -X github.com/semihbkgr/yamldiff/cmd.buildVersion=${BUILD_VERSION}

build:
	go build -v --ldflags='${LDFLAGS}' .
