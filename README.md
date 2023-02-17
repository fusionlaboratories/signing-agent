# Welcome to Qredo's Signing Agent

## What is Signing Agent?

Signing Agent is a programatic member of your Qredo custody group. It enables you to setup automated approval for transactions from a Qredo wallet. To understand how to set up a Qredo Wallet with a Signing Agent integration [see the docs](https://developers.qredo.com/signing-agent/).

The Signing Agent service provided here can operate as a standalone service (exposing a RESTful API to 3rd party applications), or as a library integrated into an application.

More details on how to use the Signing Agent can be found in [`/docs`](/docs).

## Who maintains Signing Agent?

The Qredo development team is responsible for maintaining the Signing Agent. When this tool is released to the community, external contributions will be accepted. If you have issues or feedback, please [create an issue](https://github.com/qredo/signing-agent/issues) and the team will support you.


## Dependencies
To build the Signing Agent, you will use ***AMCL - Apache Milagro Crypto Library***. Please refer to the
[apache/incubator-milagro-crypto-c](https://github.com/apache/incubator-milagro-crypto-c) repository for installation instructions.

## How to build
### Building the executable
(see the dependency above)
```shell
> make build
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
will run all unit and restAPI tests.

```shell
> make unittest
```
runs just unit test.
```shell
> make apitest
```
will run the restAPI test.

### End-to-end test (e2e)
In order to run the e2e test, the APIKEY and BASE64PKEY (the base64 of the private.pem file) for a Qredo account to test against are needed. Ensure both the APIKEY and BASE64PKEY are set in the environment before running the e2e test.
The following are required for the e2e test:

| Variable     | Description                                       |
|-------------|----------------------------------------------------|
| APIKEY      | The API key for the Qredo account to test against  |
| BASE64PKEY  | The Base64-encoded RSA private key for the account |

These should be set before the running the e2e test. And then:
```shell
> make e2etest
```
to run the e2e test.
