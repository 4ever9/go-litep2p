help: Makefile
	@echo "Choose a command run:"
	@sed -n 's/^##//p' $< | column -t -s ':' | sed -e 's/^/ /'

## make pb: build network message protobuf
proto:
	protoc --go_out=. message.proto

.PHONY: proto