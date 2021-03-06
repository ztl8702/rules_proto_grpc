load("//github.com/gogo/protobuf:gogo_grpc_compile.bzl", "gogo_grpc_compile")
load("@io_bazel_rules_go//go:def.bzl", "go_library")

def gogo_grpc_library(**kwargs):
    # Compile protos
    name_pb = kwargs.get("name") + "_pb"
    gogo_grpc_compile(
        name = name_pb,
        prefix_path = kwargs.get("importpath", ""),
        **{k: v for (k, v) in kwargs.items() if k in ("deps", "verbose")} # Forward args
    )

    # Create gogo library
    go_library(
        name = kwargs.get("name"),
        srcs = [name_pb],
        deps = kwargs.get("go_deps", []) + GRPC_DEPS,
        importpath = kwargs.get("importpath"),
        visibility = kwargs.get("visibility"),
    )

GRPC_DEPS = [
    "@com_github_gogo_protobuf//proto:go_default_library",
    "@org_golang_google_grpc//:go_default_library",
    "@org_golang_x_net//context:go_default_library",
]
