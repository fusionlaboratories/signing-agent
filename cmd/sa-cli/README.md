## Purpose
The purpose of the document is to demonstrate how to use signing agent cli (sa-cli) which uses the signing agent as a library.

## Prerequisites
Installed Golang toolchain and OpenSSL environments

## Generate keys
First we need to generate RSA keys, this can be done with commands
```openssl genrsa -out private.pem 4096```
```openssl rsa -in private.pem -pubout -out public.pem```

## Set variables
Second we need to set the PrivateKey and APIKey into defs.go file
Set content from generated private.pem file into PrivateKey variable and APIKey from the registration process.

## Compilation
For local compilation execute the build script:
```./build```

## Usages
For all available command options use:
```./sa-cli -h```

Every single command has its own help section for example:
```./sa-cli help register```
Will list all possible arguments regarding this command

## Supported commands

```./sa-cli register --name demo-client```
Will register a client with the given name

```./sa-cli create-company --name demo-company --city Sofia --country BG --domain demo-company.com --reference d-c-b-g```
Will register a company with the given arguments

```./sa-cli add-trustedparty --company-id {company_id} --agent-id {agent_id}```
Will add trusted party with given company_id and agent_id

```./sa-cli create-fund --company-id {company_id} --member-id {user_id} --fund-name newfund --fund-description a fund description```
Will create a new fund and wallets by given arguments

```./sa-cli add-whitelist --company-id {company_id} --fund-id {fund_id} --address {wallet_address}```
Will add to whitelist a given wallet address

```./sa-cli withdraw --company-id {company_id} --wallet-id {wallet_id} --address {wallet_address} --amount {amount}```
Will withdraw a given amount to the wallet address

```./sa-cli read-action --feed-url {feed_url}```
Will connect to the given feed_url and start to consume the events of the action, the operation can be stopped by CTRL+C

```./sa-cli approve-action --action-id {action_id}```
Will approve the action by given action_id
