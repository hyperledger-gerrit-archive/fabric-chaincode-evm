#
# Copyright IBM Corp. All Rights Reserved
# SPDX-License-Identifier: Apache2.0
#
FROM hyperledger/fabric-buildenv:latest
WORKDIR $GOPATH/src/github.com/hyperledger/
RUN git clone https://github.com/hyperledger/fabric.git
WORKDIR $GOPATH/src/github.com/hyperledger/fabric
RUN git checkout master
RUN EXECUTABLES=go GO_TAGS=pluginsenabled EXPERIMENTAL=false DOCKER_DYNAMIC_LINK=true make peer

COPY ./plugin/evmscc.go ./plugin/evmscc.go
COPY ./statemanager/statemanager.go ./statemanager/statemanager.go
RUN sed -i 's/fabric-chaincode-evm/fabric/g' ./plugin/evmscc.go
RUN sed -i '/[[prune]]/ i [[override]]\n  name="github.com/tmthrgd/go-hex"\n\  revision="13ed8ac0738b1390630f86f6bf0943730c6d8037"\n\n[[override]]\n\  version = "~0.6.0"\n  name = "github.com/tendermint/tmlibs"\n\n[[override]]\n  name="github.com/tendermint/go-crypto"\n  revision="dd20358a264c772b4a83e477b0cfce4c88a7001d"' ./Gopkg.toml
RUN dep ensure
RUN CGO_LDFLAGS_ALLOW="-I/usr/local/share/libtool" GO_TAGS=nopkcs11 GOOS=linux GOARCH=amd64 go build -o /go/lib/evmscc.so -buildmode=plugin ./plugin

FROM hyperledger/fabric-peer:latest
RUN apt-get update -qqy && apt-get dist-upgrade -qqy && apt-get install libltdl-dev -qqy
COPY --from=0 /opt/gopath/src/github.com/hyperledger/fabric/.build/bin/peer /usr/local/bin/peer
RUN ls -l /usr/local/bin
COPY --from=0 /go/lib/evmscc.so /opt/lib/evmscc.so
COPY config/core.yaml /etc/hyperledger/fabric
