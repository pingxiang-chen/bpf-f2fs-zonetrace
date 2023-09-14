# Protos Directory

This directory contains files related to the Protocol Buffers (Protobuf) definitions and their corresponding Go language bindings.

Protocol Buffers are a method developed by Google to serialize structured data, similar to XML or JSON.
Protocol Buffers serializes data in a binary format.
This binary serialization not only allows the data to be more compact but also facilitates quicker serialization and deserialization compared to textual formats.
Consequently, it substantially reduces the overhead and cost associated with data transfer between client and server, making it a favorable choice for performance.

You can define message types and services in the `.proto` file, which can then be compiled to generate code in various languages using the `protoc` compiler.


## Usage

To generate the Go bindings from the `.proto` files, you can use the following command:

```sh
make generate
```

To clean (remove) all generated Go files, use the following command:

```sh
make clean
```


## File Descriptions

### Makefile

The `Makefile` is a file that contains a set of directives used by the `make` utility to build the Go code generated from the Protobuf definitions.

### zns.proto

The `zns.proto` file contains the Protocol Buffers (protobuf) schema definitions.

### zns.pb.go

The `zns.pb.go` file contains the Go language bindings auto generated from the `zns.proto` file.
This file contains Go structs that represent the message types defined in the `zns.proto` file, along with additional code to serialize and deserialize these structs.



## Dependencies

Ensure you have the following dependencies installed before running the above commands:

1. Go: The Go programming language.
2. Protoc: The Protocol Buffers compiler.
3. Protoc Go Plugin: The Go plugin for the Protoc compiler.

See: https://grpc.io/docs/languages/go/quickstart/
