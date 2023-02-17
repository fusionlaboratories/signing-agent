---
comment: 
proofedDate: noProof
supportsVersion: v 1.0.0 SA
---

# Deployment Guide

## Overview

This document provides details on how to deploy the Signing Agent service.

## Supported platforms

Should work on UNIX systems. Windows is untested as a deployment environment, but there's nothing inherently non-portable, so it may well work.

## Building the binary and testing

The Signing Agent service is written in Go, using version go1.18+. It is a single executable. That executable binary can be built by running:

```bash
    make build
```

in the project root folder. This should be done on the same platform as the deployment platform, so that it uses the same CPU architecture.

This will create an executable called `out/signing-agent` in the top level of the project.

For testing, the following will run the default Go testing suite:

```bash
    make test
```

## Installation

**Note:** You should never run the service directly connected to the internet. All external requests to the API endpoints exposed by the Signing Agent should be done through a reverse proxy service (e.g. Nginx or HAProxy) and over TLS.

If running as a Docker container, a host directory must be mounted as a virtual volume, in order to ensure data persistency between different container versions.


## Service configuration data

Please see the [YAML configuration template](configuration.md) documentation for a description of each configuration option available.

## Verify your deployment

There are currently three healthcheck endpoints available that allow you to verify that your deployment is successful.

### /healthcheck/version

The version healthcheck endpoint expects a `GET` request, and it responds with an HTTP 200 status code and a JSON payload containing the current build information:

```json
{
    "buildVersion": "101c354",
    "buildType": "dev",
    "buildDate": "Fri Oct 7 12:53:12 UTC 2022"
}
```

### /healthcheck/config

The config healthcheck endpoint accepts a `GET` request, and it responds with an HTTP 200 status code and a JSON payload containing the current configuration file data:

```json
{
    "base": {
        "qredoAPI":"https://play-api.qredo.network/api/v1/p",
        ...
    },
    "http": {
        "addr":"0.0.0.0:8007",
        "CORSAllowOrigins":["*"],
        ...
    },
    "autoApproval": {
        "enabled": false,
        ...
    },
    ...
}
```

### /healthcheck/status

The status healthcheck endpoint accepts a `GET` request, and it responds with an HTTP 200 status code and a JSON payload containing the status information for various services or external connections that are related to the Signing Agent service:

```json
{
    "websocket": {
        "readyState":"CLOSED",
        "remoteFeedURL": "wss://play-api.qredo.network/api/v1/p/coreclient/feed",
        "localFeedURL":"ws://0.0.0.0:8007/api/v1/client/feed",
        "connectedClients": 2
    }
}
```
