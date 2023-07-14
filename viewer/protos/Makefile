# Makefile

# Find all .proto files in the current directory
PROTO_FILES := $(wildcard *.proto)

# Generate Go code from .proto files
generate: $(PROTO_FILES)
	protoc --go_out=. --go_opt=paths=source_relative $^

# Clean generated Go code
clean:
	rm *.pb.go
