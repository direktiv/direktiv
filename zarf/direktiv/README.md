# Direktiv

This component installs Direktiv.

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

***DIREKTIV_IMAGE***: Direktiv image to be used (Default: direktiv/direktiv)

***DIREKTIV_REQUEST_TIMEOUT***: Nginx timeout (Default: 7200)

***DIREKTIV_FUNCTION_SIZES***: Sizes for the different function definitions. See `direktiv_function_sizes` in [zarf-config-example.yaml](zarf-config-example.yaml)

***DIREKTIV_FLUENTBIT_CONFIG***: Additional fluentbit configuration. See `direktiv_fluentbit_config` in [zarf-config-example.yaml](zarf-config-example.yaml)

***DIREKTIV_OTEL_CONFIG***: Additional opentelemetry configuration. See `direktiv_otel_config` in [zarf-config-example.yaml](zarf-config-example.yaml)

