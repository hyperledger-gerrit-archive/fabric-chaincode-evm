#
# Copyright IBM Corp. All Rights Reserved
# SPDX-License-Identifier: Apache-2.0
#
# REQUIRES fabric-buildenv image
# git checkout github.com/hyperledger/fabric && cd fabric && make buildenv
#
ARG GO_TAGS
ARG CGO_LDFLAGS_ALLOW

FROM fabric-peer-evm-base:latest as evmscc-builder
COPY ./plugin/evmscc.go ./plugin/evmscc.go
COPY ./statemanager/statemanager.go ./statemanager/statemanager.go
RUN sed -i 's/fabric-chaincode-evm/fabric/g' ./plugin/evmscc.go
RUN dep ensure
RUN ${CGO_LDFLAGS_ALLOW} go build -o /go/lib/evmscc.so -tags '${GO_TAGS}' -buildmode=plugin ./plugin

FROM hyperledger/fabric-peer:latest
# TODO modify baseimage to include this
COPY --from=evmscc-builder /opt/gopath/src/github.com/hyperledger/fabric/.build/bin/peer /usr/local/bin/peer
COPY --from=evmscc-builder /go/lib/evmscc.so /opt/lib/evmscc.so
COPY config/core.yaml /etc/hyperledger/fabric
