
* You need to have an account in some qredo environment (qa, dev, local setup, etc)
* Create a partner API key
* Generate RSA key pair
```
openssl genrsa -out private.pem 2048
openssl rsa -in private.pem -outform PEM -pubout -out public.pem
```
* Upload public.pem in the partner API screen of the qredo webapp
* Set the correct values for the consts on top of main.go (`partnerAPIKey`,`rsaPrivateKey`, `qredoURL`)
* `go run github.com/qredo/signing-agent/cmd/example`
