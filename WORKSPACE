git_repository(
    name = "io_bazel_rules_go",
    remote = "https://github.com/bazelbuild/rules_go.git",
    tag = "0.5.2",
)
load("@io_bazel_rules_go//go:def.bzl", "go_repositories", "go_repository")

git_repository(
  name = "org_pubref_rules_protobuf",
  remote = "https://github.com/pubref/rules_protobuf",
  tag = "v0.7.2",
)

go_repository(
  name = "com_github_spf13_afero",
  importpath = "github.com/spf13/afero",
  commit = "9be650865eab0c12963d8753212f4f9c66cdcf12",
)

go_repository(
  name = "com_github_dinedal_nextbus",
  importpath = "github.com/dinedal/nextbus",
  commit = "c4ca90b01168a3f03b1699cf32038fa76047808c",
)

go_repository(
  name = "org_golang_x_text",
  importpath = "golang.org/x/text",
  commit = "836efe42bb4aa16aaa17b9c155d8813d336ed720",
)

load("@org_pubref_rules_protobuf//go:rules.bzl", "go_proto_repositories")
load("@org_pubref_rules_protobuf//python:rules.bzl", "py_proto_repositories")
go_proto_repositories()
py_proto_repositories()
go_repositories()


