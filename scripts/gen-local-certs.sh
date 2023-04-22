#!/usr/bin/env bash
mkdir -p testdata/selfSignedTLS

openssl ecparam -genkey -name secp384r1 -out testdata/selfSignedTLS/server.key
openssl req -new -x509 -sha256 -key ./testdata/selfSignedTLS/server.key -out testdata/selfSignedTLS/server.crt -days 3650