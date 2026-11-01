# Config builder image for CassKop

CassKop generates the Cassandra configuration with [cass-config-builder](https://github.com/datastax/cass-config-builder),
which renders templates from [cass-config-definitions](https://github.com/datastax/cass-config-definitions). Upstream
stopped publishing definitions for Apache Cassandra after 4.0.0, so any newer version falls back on the 4.0.0 files:
`cassandra-env.sh` then references `jamm-0.3.2.jar` which no longer exists in 5.0 images, and the 4.1 settings
cannot be configured.

This image is `datastax/cass-config-builder` with the missing definitions copied over `/definitions`. The files under
[definitions](definitions) follow the upstream layout so they can be contributed back as is:

| Definition | Versions added | Notes |
|---|---|---|
| `cassandra-yaml` | 4.1.0, 5.0.0 | Settings renamed with their unit (`read_request_timeout: 5000ms`), `storage_compatibility_mode` defaults to `CASSANDRA_4` |
| `cassandra-env-sh` | 5.0.0 | Template taken from the 5.0 image (`jamm-0.4.0.jar`, Java 11/17 heap sizing) |
| `jvm-server-options` | 5.0.0 | Biased locking left to `jvm11-server.options`, chronicle analytics disabled |
| `jvm11-server-options` | 5.0.0 | G1 tuning and JPMS exports shipped with 5.0 |
| `jvm17-server-options` | 5.0.0 | New file, Java 17 is the default runtime of the 5.0 images |

The `*-transforms.edn` files register the versions and replace the upstream ones, every other upstream file is left
untouched: the configuration generated for 3.11 and 4.0 is identical to the one produced by the base image.

## Build and check

```bash
docker build -t cass-config-builder:dev docker/config-builder
./docker/config-builder/check.sh cass-config-builder:dev
```

`check.sh` renders the configuration of each supported Cassandra line and checks the version specific bits. The
image is published to `ghcr.io/cscetbon/cass-config-builder` by the *Config builder image* GitHub workflow, tagged
with the base version followed by the overlay revision (`1.0.4-1`). The workflow runs on every merge touching this
directory and publishes the alias set as its default, which must match `defaultConfigBuilderImage` in the operator;
it can also be dispatched by hand with another alias.
