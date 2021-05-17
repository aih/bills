# Go parameters
        GOCMD=go
        GOBUILD=$(GOCMD) build
        GOCLEAN=$(GOCMD) clean
        GOTEST=$(GOCMD) test
        GOGET=$(GOCMD) get
        BINARY_NAME=billmeta
        BINARY_UNIX=$(BINARY_NAME)_unix
    
    all: test build
    build:
	    if [ ! -d "cmd/bin" ]; then \
			mkdir cmd/bin; \
		fi
	    $(GOBUILD) -o cmd/bin/billmeta -v cmd/billmeta/main.go
	    $(GOBUILD) -o cmd/bin/committees -v cmd/committees/main.go
	    $(GOBUILD) -o cmd/bin/comparematrix -v cmd/comparematrix/main.go
	    $(GOBUILD) -o cmd/bin/jsonpgx -v cmd/jsonpgx/main.go
	    $(GOBUILD) -o cmd/bin/legislators -v cmd/legislators/main.go
	    $(GOBUILD) -o cmd/bin/listxml -v cmd/listxml/main.go
	    $(GOBUILD) -o cmd/bin/unitedstates -v cmd/unitedstates/main.go
    test: 
	    $(GOTEST) -v ./...
    clean: 
	    $(GOCLEAN)
	    rm -f $(BINARY_NAME)
	    rm -f $(BINARY_UNIX)
    run:
	    $(GOBUILD) -o $(BINARY_NAME) -v ./...
	    ./$(BINARY_NAME)
    deps:
    #	$(GOGET) github.com/...
    
    
    # Cross compilation
    build-linux:
	    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BINARY_UNIX) -v