load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = ["nextbus.go"],
    visibility = ["//visibility:private"],
    deps = [
        "//proto:go_default_library",
        "@com_github_dinedal_nextbus//:go_default_library",
        "@org_golang_google_grpc//:go_default_library",
        "@org_golang_google_grpc//codes:go_default_library",
        "@org_golang_x_net//context:go_default_library",
    ],
)

go_binary(
    name = "nextbus",
    library = ":go_default_library",
    visibility = ["//visibility:public"],
)

go_test(
    name = "go_default_test",
    size = "small",
    srcs = ["nextbus_test.go"],
    library = ":go_default_library",
    deps = [
        "//proto:go_default_library",
        "@com_github_dinedal_nextbus//:go_default_library",
        "@com_github_golang_protobuf//proto:go_default_library",
        "@org_golang_google_grpc//:go_default_library",
        "@org_golang_google_grpc//codes:go_default_library",
    ],
)
