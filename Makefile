EXAMPLE_NAME := example

.PHONY: all
all: build test

build:
	go build -o $(EXAMPLE_NAME) ./examples/main.go 

test:
	go test -v -race -cover ./workerpool

clean:
	rm $(EXAMPLE_NAME)

run-examples: build
	./$(EXAMPLE_NAME)