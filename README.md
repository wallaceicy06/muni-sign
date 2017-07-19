# Muni Sign [![Build Status](https://travis-ci.org/wallaceicy06/muni-sign.svg?branch=master)](https://travis-ci.org/wallaceicy06/muni-sign)

Software for displaying Nextbus arrival information on a physical LCD display.

## Requirements

In order to build this program, you will need to install
[Bazel](https://docs.bazel.build/versions/master/install.html).

```shell
bazel build //...
```

## Third Party

This project makes use of the following third party libraries:

* [Afero](https://github.com/spf13/afero) (Apache 2.0)
* [GRPC](https://github.com/grpc/grpc) (Apache 2.0)
* [Nextbus](https://github.com/dinedal/nextbus) (MIT)
* [Protocol Buffers](https://github.com/google/protobuf)
