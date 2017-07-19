package(
    default_visibility = ["//visibility:public"]
)

load("@org_pubref_rules_protobuf//go:rules.bzl", "go_proto_library")
load("@org_pubref_rules_protobuf//python:rules.bzl", "py_proto_compile")

filegroup(
    name = "protos",
    srcs = ["muni_sign.proto"],
)

go_proto_library(
    name = "go_default_library",
    protos = [":protos"],
    with_grpc = True,
)

py_proto_compile(
    name = "py",
    protos = [":protos"],
    with_grpc = True,
)
