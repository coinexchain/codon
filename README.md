# codon

For better performance, compatibility and usability, we develop this serialization/deserialization library for blockchain applications, named codon. Its target is to replace [go-amino](https://github.com/tendermint/go-amino). Its name comes from "Each codon either codes for an amino acid or tells the cell to stop making a protein chain."

The advantages of codon are:

1. It is fast (about 4~5 times faster than go-amino). It does not use runtime reflection. Instead, it generates source code for serialization/deserialization beforehand, just like protobuf.
2. Its binary format is compatible with protobuf3. It is easy to inter-operate with other applications based on protobuf3.
3. It is lightweight. Its code-generation function does not depend on gogo-proto or other protobuf implementations.
4. It is friendly to Golang. You do not need to write '.proto' files to use it. It understands Golang's type definitions and infer message definitions from them. It can also dump the inferred message definitions to '.proto' files, which will be used by other programming language to inter-operate with it.
5. It keeps the same API as go-amino (almost). So it can be integrate into Cosmos-SDK without many modifications.
6. It supports filling structures randomly. You can write fuzz tests for the generated serialization/deserialization code.
7. It supports deepcopy of structures.

Currently [a branch of Cosmos-SDK](https://github.com/coinexchain/cosmos-sdk/tree/use_codon) is using codon. But codon is not mature yet. More tests are needed before its finally deployment.

It also has some limitations:

1. It does not support maps. Anyway, blockchain application would not serialize maps.
2. It does not support nil members in struct.
3. It does not support private members in struct.



### Code Generation

You must generate codec source code using codon beforehand. This directory [codongen](https://github.com/coinexchain/cosmos-sdk/tree/use_codon/codongen) is an example.

codongen/codec/types.go: In this file, you use type aliases to include the types which the codec will support.

codongen/codec/prepare.go: In this file, there are some glue logic to generate codec source code using codon.

codongen/gen/main.go: This file will be used for `go run`. It calls the glue logic in prepare.go.

codongen/codec/codec.go: `go run main.go` will print the generated source code to stdout. Please redirect its stdout to `codec/codec.txt` and examine its content. If there are no error reports in this file, you can rename it as `codec/codec.go`.

### Benchmark and Fuzz Test

In the directory [codongen](https://github.com/coinexchain/cosmos-sdk/tree/use_codon/codongen) there are also a benchmark and a fuzz tester.

codongen/bench/main.go: This is the benchmark. To run it, you need to prepare a large file with random content. The command to run it is `go run bench/main.go some_large_random_file.dat`. It shows that codon is about 4~5 times faster than go-amino.

codongen/fuzz/main.go: This is the fuzz tester. The command to run it is `go run fuzz/main.go some_large_random_file.dat`. 

### Dump .proto file for other programming language

codon strictly adheres to [the protobuf3 encoding specification](https://developers.google.com/protocol-buffers/docs/encoding). It can generate a .proto file for other programming languages, which descripts the binary messages' formats it reads and writes.

codongen/types.proto: This is the dumped .proto file, which contains some types from Cosmos-SDK.

