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

# Add security directive to fix redocly lint error.  signing-agent api doesn't use authentication directives. The
# linter checks for authentication. Adding "security":[] satisfies the linter. It's needed because swagger generate
# doesn't add the directive from the code.
# The following injects "security":[] at root level in the three swagger files.
sed -i '' "s/^{\"openapi\":\"3.0.1\",/{\"openapi\":\"3.0.1\",\"security\":\[\],/" ./docs/swagger/openapi.json
sed -i '' "s/^{/{\n  \"security\":\[\],/" ./docs/swagger/swagger.json
sed -i '' "s/^basePath/security: \[\]\nbasePath/" ./docs/swagger/swagger.yaml