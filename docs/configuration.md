---
comment: 
proofedDate: noProof
supportsVersion: v 1.0.0 SA
---

# YAML config file template

> Note, in production you must use only one `store` method, i.e. `file`, `oci`, or `aws`.

```yaml
base:
  qredoAPI: https://play-api.qredo.network/api/v1/p
  pin: 0
autoApproval:
  enabled: false
  retryIntervalMaxSec: 300
  retryIntervalSec: 5
websocket:
  qredoWebsocket: wss://play-api.qredo.network/api/v1/p/coreclient/feed
  reconnectTimeoutSec: 300
  reconnectIntervalSec: 5
  pingPeriodSec: 5
  pongWaitSec: 10
  writeWaitSec: 10
  readBufferSize: 512
  writeBufferSize: 1024
http:
  addr: 0.0.0.0:8007
  CORSAllowOrigins:
    - '*'
  logAllRequests: false
  TLS:
    enabled: true
    certFile: tls/domain.crt
    keyFile: tls/domain.key
logging:
  format: text
  level: debug
loadBalancing:
  enable: false
  onLockErrorTimeoutMs: 300
  actionIDExpirationSec: 6
  redis:
    host: localhost
    port: 6379
    password: ""
    db: 0
store:
  type: file
  file: /volume/ccstore.db
  oci:
    compartment: ocid1.tenancy.oc1...
    vault: ocid1.vault.oc1...
    secretEncryptionKey: ocid1.key.oc1...
    configSecret: signing_agent_config
  aws:
    region: aws-region-...
    configSecret: secrets_manager_secret...
```

## Base

- **qredoAPI:** the url of the api you want to use
- **pin:** the pin number to use to provide a zero knowledge proof token for communication with the partner api

## Auto approval
- **enabled:** activate the automatic approval of every transaction that is received
- **retryIntervalMaxSec:** the maximum time in which the Signing Agent retries to approve an action. After that it’s considered as a failure
- **retryIntervalSec:** the interval in which the Signing Agent is attempting to approve an action. It will retry until the retryIntervalMaxSec is reached

## Websocket
- **qredoWebsocket:** the url of the websocket feed you want to use
- **reconnectTimeoutSec:** the reconnect timeout in seconds
- **reconnectIntervalSec:** the reconnect interval in seconds
- **pingPeriodSec:** the ping period for the ping handler in seconds
- **pongWaitSec:** the pong wait for the pong handler in seconds
- **writeWaitSec:** the write wait in seconds
- **readBufferSize:** the websocket upgrader read buffer size in bytes
- **writeBufferSize:** the websocket upgrader write buffer size in bytes

## HTTP

- **addr:** the address and port the service runs on [the bind address and port the build in api endpoints]
- **CORSAllowOrigins:** the value the the Access-Control-Allow-Origin of the responses of the build in api
- **logAllRequests:** log all incoming requests to the build in api
- **TLS**
  - **enabled:** wether or not you want to enable tls on the server side
  - **certFile:** path to the cert file you want to use
  - **keyFile:** path to the key file you want to use


## Logging

- **format:** the format of the logging, ex. text or json
- **level:** the level of logging, ex. debug, info, warn, error, default is debug

## Load balancing

- **enable:** enables the load-balancing logic
- **onLockErrorTimeoutMs:** on lock timeout in milliseconds
- **actionIDExpirationSec:** expiration of actionID variable in Redis in seconds
- **redis:**
  - **host:** Redis host
  - **port:** Redis port
  - **password:** Redis password
  - **db:** Redis database to be selected after connecting to the server

## Store

- **type:** the type of store to use to store the private key information for the Signing Agent, ex. file, oci, aws
- **file:** the path to the storage file when file store is used
- **oci:** the oracle cloud configuration to store the private keys in an oracle vault
  - **compartment:** the OCID where the vault and encryption key reside
  - **vault:** the OCID of the vault where the secret will be stored
  - **secretEncryptionKey:** the encryption key used for both the secret and the data inside the secret
  - **configSecret:** the name of secret that will be used to store the data
- **aws:** the amazon cloud configuration to store the private keys in amazon secrets manager
  - **region:** the AWS region where the secret is stored
  - **configSecret:** the name of the AWS Secrets Manager secret containing the encrypted data
