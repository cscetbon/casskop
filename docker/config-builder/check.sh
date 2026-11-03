#!/usr/bin/env bash
# Renders the configuration of every supported Cassandra line with the given config builder image
# and checks the version specific bits that casskop relies on.
set -euo pipefail

IMAGE=${1:?usage: check.sh <config-builder-image>}

render() {
  local version=$1 config=$2
  docker run --rm -u 0 -e PRODUCT_NAME=cassandra -e PRODUCT_VERSION="$version" -e CONFIG_FILE_DATA="$config" \
    -e CONFIG_OUTPUT_DIRECTORY=/bootstrap -e RACK_NAME=rack1 -e POD_IP=10.0.0.1 --entrypoint bash "$IMAGE" \
    -c 'mkdir -p /bootstrap && run_configbuilder >/dev/null && cd /bootstrap && for f in *; do echo "### $f"; cat "$f"; done'
}

expect() {
  local output=$1 pattern=$2
  grep -qE -- "$pattern" <<<"$output" || { echo "missing: $pattern"; grep "^###" <<<"$output"; exit 1; }
}

refuse() {
  local output=$1 pattern=$2
  ! grep -qE -- "$pattern" <<<"$output" || { echo "unexpected: $pattern"; exit 1; }
}

# config builder input with the given cassandra-yaml settings and JVM options section
config() {
  printf '{"cassandra-yaml":{%s},"%s":%s,%s}' "$1" "$2" \
    '{"initial_heap_size":"512M","max_heap_size":"512M","cassandra_ring_delay_ms":30000,"jmx-connection-type":"remote-no-auth"}' \
    '"cluster-info":{"name":"check","seeds":"seed"},"datacenter-info":{"name":"dc1"},"logback-xml":{"debuglog-enabled":false}'
}

echo "== 3.11.19"
out=$(render 3.11.19 "$(config '"read_request_timeout_in_ms":5000' jvm-options)")
expect "$out" "^### jvm.options"
expect "$out" "^read_request_timeout_in_ms: 5000"
refuse "$out" "^### jvm-server.options"

echo "== 4.0.2"
out=$(render 4.0.2 "$(config '"read_request_timeout_in_ms":5000' jvm-server-options)")
expect "$out" "^### jvm11-server.options"
expect "$out" "jamm-0.3.2.jar"
expect "$out" "^read_request_timeout_in_ms: 5000"

echo "== 4.1.5"
out=$(render 4.1.5 "$(config '"read_request_timeout":"5000ms"' jvm-server-options)")
expect "$out" "jamm-0.3.2.jar"
expect "$out" "^read_request_timeout: 5000ms"
expect "$out" "^compaction_throughput: 64MiB/s"
refuse "$out" "_in_ms:"
refuse "$out" "^### jvm17-server.options"

echo "== 5.0.1"
out=$(render 5.0.1 "$(config '"read_request_timeout":"5000ms"' jvm-server-options)")
expect "$out" "jamm-0.4.0.jar"
expect "$out" "^### jvm17-server.options"
expect "$out" "^storage_compatibility_mode: CASSANDRA_4"
expect "$out" "^read_request_timeout: 5000ms"
expect "$out" "chronicle.analytics.disable=true"
refuse "$out" "_in_ms:"
refuse "$out" "scripted_user_defined_functions_enabled"

echo "all checks passed"
