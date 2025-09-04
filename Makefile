.PHONY: run processor
run:
	go run cmd/server/main.go

processor:
	go run cmd/processor/main.go