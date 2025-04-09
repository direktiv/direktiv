# KNATIVE

Installation of Knative operator and an Knative instance. Due to issues with Knative images the `prepare-images.sh` needs to be called before a version upgrade and the `registry/default` value changed to the new values.

## Variables

**KNATIVE_HTTP_PROXY**: Proxy settings for HTTP requests
**KNATIVE_HTTPS_PROXY**: Proxy settings for HTTPS requests
**KNATIVE_NO_PROXY**: IP/Domains not using the proxy
