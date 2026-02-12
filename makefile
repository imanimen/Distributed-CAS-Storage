build:
	@go build -o bin/fs

run: build
	@./bin/cas

test:
	@go test ./... -v