# Go parameters
        GOCMD=go
        GOBUILD=$(GOCMD) build
        GOCLEAN=$(GOCMD) clean
        GOTEST=$(GOCMD) test
        GOGET=$(GOCMD) get
        BINARY_NAME=billmeta
        BINARY_UNIX=$(BINARY_NAME)_unix$
    
    all: test build
    build:
	    if [ ! -d "cmd/bin" ]; then \
			mkdir cmd/bin; \
		fi 	
	    for go_pkg in $$(go list ./... | grep cmd); do \
	    	echo "Building $$go_pkg"; \
	    	$(GOBUILD) -o cmd/bin/ $$go_pkg ; \
	    done
    test: 
	    $(GOTEST) -v .
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