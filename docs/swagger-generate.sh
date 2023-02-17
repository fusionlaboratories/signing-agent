#!/bin/bash

# https://ldej.nl/post/generating-swagger-docs-from-go/
# https://goswagger.io/install.html
# https://editor.swagger.io/

SWAGGER_GENERATE_EXTENSION=false swagger generate spec -m -o docs/swagger/swagger.yaml \
      -c rest \
      -c signing-agent/api \
      -c config

SWAGGER_GENERATE_EXTENSION=false swagger generate spec -m -o docs/swagger/swagger.json \
      -c rest \
      -c signing-agent/api \
      -c config

curl -X POST https://converter.swagger.io/api/convert -d @./docs/swagger/swagger.json --header 'Content-Type: application/json' > ./docs/swagger/openapi.json
