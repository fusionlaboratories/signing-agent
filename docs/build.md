# Build

## Dependencies

> To use Signing Agent, we recommend building the Docker image. However, should you wish to develop or extend this tool, you can self-build.

To build the Signing Agent, you will use:

- ***AMCL - Apache Milagro Crypto Library***

> Please refer to the
[apache/incubator-milagro-crypto-c](https://github.com/apache/incubator-milagro-crypto-c) repository for installation instructions.

- Golang environment

## How to build

### Building the executable

> Tested on UNIX systems. Windows is untested as a deployment environment, but there's nothing inherently non-portable.

The Signing Agent service is written in Go, using version go1.18+. It is a single executable. That executable binary can be built by running the following in the project root folder:

```bash
    make build
```

This creates the executable: ***./out/signing-agent***

### Building the signing-agent docker container
```shell
> ./build.sh docker_latest
```
This creates the signing-agent docker image: ***signing-agent:latest***
## Running Tests
```shell
> make test
```
will run all unit tests.

```shell
> make unittest
```
