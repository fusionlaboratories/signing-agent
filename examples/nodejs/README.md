Running Signing Agent as a Docker container:

- create a separate Docker network: `docker network create signing-agent-net`
- create a stopped Docker container for the Signing Agent: `docker create --name temp_signing_agent signing-agent:latest`
- create an empty folder to store config and data (if store type is file): `mkdir -p ./volume`
- copy the config template from the created container local: `docker cp temp_signing_agent:/signing-agent/config-template.yaml ./volume/config.yaml`
- edit the config.yaml as per your requirements

    > See the guide to [configuration](https://developers.qredo.com/signing-agent/v2-signing-agent/configure)

- delete the temporary container `docker rm temp_signing_agent`
- start a Signing Agent container with the volume folder mounted in and connected to the previously created network:
```
docker run --detach \
 --name signing-agent \
 --net signing-agent-net \
 -v $PWD/volume:/volume \
 signing-agent:latest
```
- check the logs to see if the signing agent started correctly: `docker logs signing-agent`

Running the javascript example inside a container:

- go to the example folder
- download a nodejs Docker image: `docker pull node:18.12.0-slim`
- download the node packages: `npm install`
- run a Docker container with the js example and the private.pem file mounted in, set the environment variables as needed:
```
docker run -ti --rm --name signing-agent-js-example --net signing-agent-net \
-e SIGNING_AGENT_HOST=signing-agent \
-e SIGNING_AGENT_PORT=8007 \
-e SIGNING_AGENT_NAME=test-agent \
-e SIGNING_AGENT_PRIVATE_KEY=/private.pem \
-e APIKEY=$APIKEY \
-e COMPANYID=$COMPANYID \
-e COINBASE_APIKEY=$COINBASE_APIKEY \
-e COINBASE_APISECRET=$COINBASE_APISECRET \
-v $PWD/private.pem:/private.pem \
-v $PWD:/client \
node:18.12.0 \
node /client/signingagent-client.js
```

Building the example into a single js file: `npm run build`
Running the single js file:
```
docker run -ti --rm --name signing-agent-js-example --net signing-agent-net \
-e SIGNING_AGENT_HOST=signing-agent \
-e SIGNING_AGENT_PORT=8007 \
-e SIGNING_AGENT_NAME=test-agent \
-e SIGNING_AGENT_PRIVATE_KEY=/private.pem \
-e APIKEY=$APIKEY \
-e COMPANYID=$COMPANYID \
-e COINBASE_APIKEY=$COINBASE_APIKEY \
-e COINBASE_APISECRET=$COINBASE_APISECRET \
-v $PWD/private.pem:/private.pem \
-v $PWD:/client \
node:18.12.0 \
node /client/dist/signing-agent-js-example.js
```
