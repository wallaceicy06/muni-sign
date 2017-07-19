load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["driver.go"],
    visibility = ["//visibility:private"],
    deps = [
        "//proto:go_default_library",
        "@org_golang_google_grpc//:go_default_library",
    ],
)

go_binary(
    name = "driver",
    library = ":go_default_library",
    visibility = ["//visibility:public"],
)
