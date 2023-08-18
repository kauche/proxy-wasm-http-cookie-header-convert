# proxy-wasm-http-cookie-header-convert

A [proxy-wasm](https://github.com/proxy-wasm/spec) compliant WebAssembly module for making proxies convert an HTTP Cookie to a Header.

## Overview

This [proxy-wasm](https://github.com/proxy-wasm/spec) compliant WebAssembly module makes proxies convert a HTTP Cookie to a Header.

## Usage

1. Download the latest WebAssembly module binary from the [release page](https://github.com/kauche/proxy-wasm-http-cookie-header-convert/releases).

2. Configure the proxy to use the WebAssembly module like below (this assumes [Envoy](https://www.envoyproxy.io/) as the proxy):

```yaml
listeners:
  - name: example
    filter_chains:
      - filters:
          - name: envoy.filters.network.http_connection_manager
            typed_config:
              # ...
              http_filters:
                - name: envoy.filters.http.wasm
                  typed_config:
                    '@type': type.googleapis.com/udpa.type.v1.TypedStruct
                    type_url: type.googleapis.com/envoy.extensions.filters.http.wasm.v3.Wasm
                    value:
                      config:
                        vm_config:
                          runtime: envoy.wasm.runtime.v8
                          code:
                            local:
                              filename: /etc/envoy/proxy-wasm-http-cookie-header-convert.wasm
                        configuration:
                          "@type": type.googleapis.com/google.protobuf.StringValue
                          value: |
                            {
                              "cookie_name": "access_token",
                              "header_name": "authorization",
                              "header_value_prefix": "bearer "
                            }
                - name: envoy.filters.http.router
                  typed_config:
                    '@type': type.googleapis.com/envoy.extensions.filters.http.router.v3.Router
```

We can also configure this WebAssembly module for the per Route basis by using the [Composite Filter](https://www.envoyproxy.io/docs/envoy/v1.24.1/configuration/http/http_filters/composite_filter). See the [example `envoy.yaml`](https://github.com/kauche/proxy-wasm-http-cookie-header-convert/blob/main/test/envoy.yaml) for more details.

### Plugin Configurations

- `cookie_name` (Required)
    - The name of a HTTP Cookie you want to convert to a Header.
- `header_name` (Required)
    - The name of a destination HTTP Header.
- `header_value_prefix` (Optional)
    - The prefix of a destination HTTP Header.
