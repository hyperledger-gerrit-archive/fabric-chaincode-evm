#
# Copyright IBM Corp. All Rights Reserved
# SPDX-License-Identifier: Apache2.0
#
FROM hyperledger/fabric-peer:latest
COPY .build/linux/lib/evmscc.so /opt/lib
COPY config/core.yaml /etc/hyperledger/fabric
