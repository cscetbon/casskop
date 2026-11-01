# cass-config-builder with the Cassandra definitions that upstream stopped publishing after 4.0.0
# (https://github.com/datastax/cass-config-definitions). Files under definitions/ follow the upstream
# layout and are copied over the ones shipped in the base image: new versions are added and the
# *-transforms.edn files are replaced to register them.
ARG BASE_VERSION=1.0.4
FROM datastax/cass-config-builder:${BASE_VERSION}

LABEL org.opencontainers.image.description="cass-config-builder with Cassandra 4.1 and 5.0 definitions for CassKop"
LABEL org.opencontainers.image.url="https://github.com/cscetbon/casskop"
LABEL org.opencontainers.image.source="https://github.com/cscetbon/casskop"

COPY definitions/ /definitions/
