#
# Copyright IBM Corp. All Rights Reserved
# SPDX-License-Identifier: Apache2.0
#
# FROM golang:1.10.1
FROM hyperledger/fabric-buildenv:latest
WORKDIR $GOPATH/src/github.com/hyperledger/fabric-chaincode-evm
COPY . .
RUN CGO_LDFLAGS_ALLOW="-I/usr/local/share/libtool" GO_TAGS=nopkcs11 GOOS=linux GOARCH=amd64 go build -o /go/lib/evmscc.so -buildmode=plugin ./plugin

FROM hyperledger/fabric-peer:latest
COPY --from=0 /go/lib/evmscc.so /opt/lib/evmscc.so
COPY config/core.yaml /etc/hyperledger/fabric
