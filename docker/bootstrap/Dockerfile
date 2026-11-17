# Copyright 2019 Orange
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

FROM debian:bookworm-slim

ENV \
    CASSANDRA_DATA=/var/lib/cassandra \
    CASSANDRA_CONF=/etc/cassandra \
    CASSANDRA_LIBS=/extra-lib \
    CASSANDRA_TOOLS=/opt/bin \
    BOOTSTRAP_CONF=/bootstrap/conf \
    BOOTSTRAP_LIBS=/bootstrap/libs \
    BOOTSTRAP_TOOLS=/bootstrap/tools \
    CONFIGMAP=/configmap \
    JOLOKIA_VERSION=1.7.2 \
    EXPORTER_VERSION=0.9.8 \
    EXPORTER_VERSION_41=0.9.14 \
    CURL_VERSION=8.23.0-1

ARG BUILD_DATE
ARG VCS_REF
ARG http_proxy
ARG https_proxy

LABEL \
    org.label-schema.build-date=$BUILD_DATE \
    org.label-schema.docker.dockerfile="/Dockerfile" \
    org.label-schema.license="Apache License 2.0" \
    org.label-schema.name="Cassandra container optimized for Kubernetes" \
    org.label-schema.url="https://github.com/cscetbon/" \
    org.label-schema.vcs-ref=$VCS_REF \
    org.label-schema.vcs-type="Git" \
    org.label-schema.vcs-url="https://github.com/cscetbon/CassKop" \
    org.label-schema.description="Cassandra Bootstrap Docker Image to be used with the CassKop Operator." \
    org.label-schema.summary="Cassandra is a NoSQL database providing scalability and hight availability without compromising performance." \
    org.label-schema.version="${CASSANDRA_VERSION}" \
    org.label-schema.changelog-url="/Changelog.md" \
    org.label-schema.maintainer="TODO" \
    org.label-schema.vendor="Orange" \
    org.label-schema.schema_version="RC1" \
    org.label-schema.usage='/README.md'

# set by buildkit to the architecture the image is built for
ARG TARGETARCH

RUN mkdir -p $CASSANDRA_DATA  $CASSANDRA_CONF $BOOTSTRAP_CONF $BOOTSTRAP_LIBS $CASSANDRA_LIBS $BOOTSTRAP_TOOLS $CASSANDRA_TOOLS /tmp \
    && apt-get update \
    && apt-get -qq -y install --no-install-recommends ca-certificates wget netcat-openbsd xz-utils \
    # cassandra-exporter agents are compiled against a Cassandra version: 0.9.8 works with 3.x, 0.9.14 with 4.1
    && wget -q -O ${BOOTSTRAP_LIBS}/cassandra-exporter-agent.jar https://github.com/instaclustr/cassandra-exporter/releases/download/v${EXPORTER_VERSION}/cassandra-exporter-agent-${EXPORTER_VERSION}.jar \
    && wget -q -O ${BOOTSTRAP_LIBS}/cassandra-exporter-agent-4.1.jar https://github.com/instaclustr/cassandra-exporter/releases/download/v${EXPORTER_VERSION_41}/cassandra-exporter-agent-${EXPORTER_VERSION_41}.jar \
    && wget -q -O ${BOOTSTRAP_LIBS}/jolokia-agent.jar https://repo1.maven.org/maven2/org/jolokia/jolokia-jvm/${JOLOKIA_VERSION}/jolokia-jvm-${JOLOKIA_VERSION}.jar \
    # static curl binary used by the readiness/liveness probes from within the cassandra container,
    # which must match the architecture of the image
    && case "${TARGETARCH:-amd64}" in \
        amd64) curl_arch=x86_64 curl_sha256=2e0c54ba696afd0df1e899e4344fa63d53d73107f9ac08173314d9be85bc76fd;; \
        arm64) curl_arch=aarch64 curl_sha256=e20f26e5b87c1883608f286f86bdabf39d3b8556e8a8e75dfe205e14bd206199;; \
        *) echo "unsupported architecture ${TARGETARCH}" >&2; exit 1;; \
    esac \
    && wget -q -O /tmp/curl.tar.xz https://github.com/stunnel/static-curl/releases/download/${CURL_VERSION}/curl-linux-${curl_arch}-musl-${CURL_VERSION}.tar.xz \
    # checksums of the downloads above, to update together with the versions
    && printf '%s  %s\n' \
        cd39bbd7c709a670ceabc9e594e9f07ca781debe241af0f18fccfcc59e23abb1 ${BOOTSTRAP_LIBS}/cassandra-exporter-agent.jar \
        7a1711e724cab41b12ee7a1a50abd83978574fa228ce9611f1664fb19a2bbfd5 ${BOOTSTRAP_LIBS}/cassandra-exporter-agent-4.1.jar \
        65d5645b07f7bdc22c7a1dbf060b5345b564674cc263d14f75ade2dabe80f1f2 ${BOOTSTRAP_LIBS}/jolokia-agent.jar \
        ${curl_sha256} /tmp/curl.tar.xz \
        | sha256sum -c - \
    && tar -xJf /tmp/curl.tar.xz -C ${BOOTSTRAP_TOOLS} curl \
    && rm /tmp/curl.tar.xz \
    && groupadd -r cassandra --gid=999 && useradd -r -g cassandra --uid=999 cassandra \
    && chown -R cassandra:cassandra $CASSANDRA_DATA $CASSANDRA_CONF $BOOTSTRAP_CONF $BOOTSTRAP_LIBS $CASSANDRA_LIBS $BOOTSTRAP_TOOLS $CASSANDRA_TOOLS /tmp \
    && apt-get -y purge \
    && apt-get -y autoremove \
    && apt-get -y clean \
    && rm -rf doc \
              man \
              info \
              locale \
              common-licenses \
              ~/.bashrc \
              /var/lib/apt/lists/* \
              /var/log/**/* \
              /var/cache/debconf/* \
              /etc/systemd \
              /lib/lsb \
              /lib/udev \
              /usr/share/doc/ \
              /usr/share/doc-base/ \
              /usr/share/man/ \
              /tmp/*
    
USER cassandra

COPY files /bootstrap/conf/

# install bootstrap entry point
COPY bootstrap.sh /
ENTRYPOINT ["/bootstrap.sh"]

