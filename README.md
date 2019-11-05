# Pre
- Download Docker

# For TLS cert
### In development
`openssl req -newkey rsa:4096 -nodes -sha512 -x509 -days 3650 -nodes -out ./key/certs/mycert.pem -keyout ./key/private/mykey.pem`

### In production
`openssl genrsa -out ./grpcAWS_test/key/private/mykey.pem 4096`
`openssl req -new -key ./key/private/mykey.pem -out ./key/certs/mycert.csr`


# Docker 
Run this to pull
`docker pull amazon/dynamodb-local`
Run this to start
`docker run -t -p 10001:8000 amazon/dynamodb-local`

# AWS dynamo
https://docs.aws.amazon.com/en_pv/sdk-for-go/v1/developer-guide/using-dynamodb-with-go-sdk.html


# Server
#### Basic Method
##### CreateObject
Add object to dynamoDB
##### GetObject
Get object to dynamoDB