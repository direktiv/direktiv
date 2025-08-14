# Direktiv

This component installs Direktiv. It can be deployed with TLS certificates with `--set DIREKTIV_WITH_CERTIFICATE=true` added to the command line during deployment. This setting looks for two files `server.key` and `server.crt` and uses those for TLS connectivity. If they don't exist these files are getting create based on the value provided with `DIREKTIV_HOST`.

zarf package deploy --set DIREKTIV_INGRESS_HOSTPORT=true --set DIREKTIV_INGRESS_SERVICE_TYPE=ClusterIP --set DIREKTIV_IMAGE=direktiv --set DIREKTIV_TAG=dev  --set DIREKTIV_REGISTRY=localhost:5001

## Variables

***DATABASE_HOST***: Database host, if not set it fetches the internal Postgres installation automatically 

***DATABASE_PORT***: Database port, if not set it fetches the internal Postgres installation automatically (Default: 5432)

***DATABASE_USER***: Database user, if not set it fetches the internal Postgres installation automatically 

***DATABASE_PASSWORD***: Database password, if not set it fetches the internal Postgres installation automatically 

***DATABASE_DBNAME***: Database database name, if not set it fetches the internal Postgres installation automatically 

***DATABASE_SSLMODE***: Database ssl mode (Default: require)

***DIREKTIV_HTTP_PROXY***: HTTP proxy settings (Default : "")

***DIREKTIV_HTTPS_PROXY***: HTTPS proxy settings (Default : "")

***DIREKTIV_NO_PROXY***: HTTP no proxy settings (Default : "")

***DIREKTIV_REPLICAS***: Number of flow replicas (Default: 1)

***DIREKTIV_INGRESS_INSTALL***: Installation of Nginx (Default: true)

***ZARF_VAR_DIREKTIV_REGISTRY***: Empty for production installation but can be set to `localhost:5000` for local installations.

***DIREKTIV_INGRESS_HOSTPORT***: Set to true if installed in Kind cluster.

***DIREKTIV_INGRESS_SERVICE_TYPE***: Set to `ClusterIP` for Kind installations.

***DIREKTIV_IMAGE***: Direktiv image to be used (Default: direktiv/direktiv)

***DIREKTIV_TAG***: Direktiv tag to use

***DIREKTIV_REQUEST_TIMEOUT***: Nginx timeout (Default: 7200)

***DIREKTIV_FUNCTION_SIZES***: Sizes for the different function definitions. See `direktiv_function_sizes` in [zarf-config-example.yaml](zarf-config-example.yaml)

***DIREKTIV_FLUENTBIT_CONFIG***: Additional fluentbit configuration. See `direktiv_fluentbit_config` in [zarf-config-example.yaml](zarf-config-example.yaml)

***DIREKTIV_OTEL_CONFIG***: Additional opentelemetry configuration. See `direktiv_otel_config` in [zarf-config-example.yaml](zarf-config-example.yaml)

***DIREKTIV_HOST***: Hostname the ingress is listeing to. It can be set either statically or with ``zarf package deploy --set DIREKTIV_HOST= `hostname` ``

***DIREKTIV_WITH_CERTIFICATE***: If set to `true` the files `server.key` and `server.crt` are being used to create the `direktiv-tls` secret.

***DIREKTIV_CERTIFICATE***: Kubernetes secret to use for TLS. Automatically set if `DIREKTIV_WITH_CERTIFICATE` is set to `true` and the files `server.crt` and `server.key` are available.