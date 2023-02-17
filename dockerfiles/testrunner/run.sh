#!/bin/bash

docker build . -t signing-agent-testrunner
docker run -ti -v $PWD/../..:/src/signing-agent -p 8007:8007 signing-agent-testrunner sh
