package main

var rustWorkspaceTemplate = mustTemplate(`load("@rules_proto_grpc//{{ .Lang.Dir }}:repositories.bzl", rules_proto_grpc_{{ .Lang.Name }}_repos="{{ .Lang.Name }}_repos")

rules_proto_grpc_{{ .Lang.Name }}_repos()

load("@io_bazel_rules_rust//rust:repositories.bzl", "rust_repositories")

rust_repositories()

load("@io_bazel_rules_rust//:workspace.bzl", "bazel_version")

bazel_version(name = "bazel_version")

load("@io_bazel_rules_rust//proto:repositories.bzl", "rust_proto_repositories")

rust_proto_repositories()`)

var rustLibraryRuleTemplateString = `load("//{{ .Lang.Dir }}:{{ .Lang.Name }}_{{ .Rule.Kind }}_compile.bzl", "{{ .Lang.Name }}_{{ .Rule.Kind }}_compile")
load("//{{ .Lang.Dir }}:rust_proto_lib.bzl", "rust_proto_lib")
load("@io_bazel_rules_rust//rust:rust.bzl", "rust_library")

def {{ .Rule.Name }}(**kwargs):
    # Compile protos
    name_pb = kwargs.get("name") + "_pb"
    name_lib = kwargs.get("name") + "_lib"
    {{ .Lang.Name }}_{{ .Rule.Kind }}_compile(
        name = name_pb,
        **{k: v for (k, v) in kwargs.items() if k in ("deps", "verbose")} # Forward args
    )
`

var rustProtoLibraryRuleTemplate = mustTemplate(rustLibraryRuleTemplateString + `
    # Create lib file
    rust_proto_lib(
        name = name_lib,
        compilation = name_pb,
        grpc = False,
    )

    # Create {{ .Lang.Name }} library
    rust_library(
        name = kwargs.get("name"),
        srcs = [name_pb, name_lib],
        deps = PROTO_DEPS,
        visibility = kwargs.get("visibility"),
    )

PROTO_DEPS = [
    Label("//rust/raze:protobuf"),
]`)

var rustGrpcLibraryRuleTemplate = mustTemplate(rustLibraryRuleTemplateString + `
    # Create lib file
    rust_proto_lib(
        name = name_lib,
        compilation = name_pb,
        grpc = True,
    )

    # Create {{ .Lang.Name }} library
    rust_library(
        name = kwargs.get("name"),
        srcs = [name_pb, name_lib],
        deps = GRPC_DEPS,
        visibility = kwargs.get("visibility"),
    )

GRPC_DEPS = [
    Label("//rust/raze:futures"),
    Label("//rust/raze:grpcio"),
    Label("//rust/raze:protobuf"),
]`)

func makeRust() *Language {
	return &Language{
		Dir:  "rust",
		Name: "rust",
		DisplayName: "Rust",
		Notes: mustTemplate("Rules for generating Rust protobuf and gRPC `.rs` files and libraries using [rust-protobuf](https://github.com/stepancheg/rust-protobuf) and [grpc-rs](https://github.com/pingcap/grpc-rs). Libraries are created with `rust_library` from [rules_rust](https://github.com/bazelbuild/rules_rust). Due to upstream requirements, these rules require that the system has a valid GOPATH set."),
		Flags: commonLangFlags,
		SkipTestPlatforms: []string{"windows"}, // CI has no rust toolchain for windows
		Rules: []*Rule{
			&Rule{
				Name:             "rust_proto_compile",
				Kind:             "proto",
				Implementation:   aspectRuleTemplate,
				Plugins:          []string{"//rust:rust_plugin"},
				WorkspaceExample: rustWorkspaceTemplate,
				BuildExample:     protoCompileExampleTemplate,
				Doc:              "Generates Rust protobuf `.rs` artifacts",
				Attrs:            aspectProtoCompileAttrs,
			},
			&Rule{
				Name:             "rust_grpc_compile",
				Kind:             "grpc",
				Implementation:   aspectRuleTemplate,
				Plugins:          []string{"//rust:rust_plugin", "//rust:grpc_rust_plugin"},
				WorkspaceExample: rustWorkspaceTemplate,
				BuildExample:     grpcCompileExampleTemplate,
				Doc:              "Generates Rust protobuf+gRPC `.rs` artifacts",
				Attrs:            aspectProtoCompileAttrs,
			},
			&Rule{
				Name:             "rust_proto_library",
				Kind:             "proto",
				Implementation:   rustProtoLibraryRuleTemplate,
				WorkspaceExample: rustWorkspaceTemplate,
				BuildExample:     protoLibraryExampleTemplate,
				Doc:              "Generates a Rust protobuf library using `rust_library` from `rules_rust`",
				Attrs:            aspectProtoCompileAttrs,
			},
			&Rule{
				Name:             "rust_grpc_library",
				Kind:             "grpc",
				Implementation:   rustGrpcLibraryRuleTemplate,
				WorkspaceExample: rustWorkspaceTemplate,
				BuildExample:     grpcLibraryExampleTemplate,
				Doc:              "Generates a Rust protobuf+gRPC library using `rust_library` from `rules_rust`",
				Attrs:            aspectProtoCompileAttrs,
			},
		},
	}
}
