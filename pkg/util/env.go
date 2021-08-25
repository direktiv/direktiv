package util

// Environtment variable keys
const (
	DBConn                   = "DIREKTIV_DB"
	DirektivDebug            = "DIREKTIV_DEBUG"
	DirektivServiceNamespace = "DIREKTIV_SERVICE_NAMESPACE"
	DirektivNamespace        = "DIREKTIV_NAMESPACE"

	DirektivFlowEndpoint      = "DIREKTIV_FLOW_ENDPOINT"
	DirektivFunctionsEndpoint = "DIREKTIV_FUNCTIONS_ENDPOINT"
	DirektivIngressEndpoint   = "DIREKTIV_INGRESS_ENDPOINT"
	DirektivMaxServerRcv      = "DIREKTIV_GRPC_MAX_SERVER_RCV"
	DirektivMaxClientRcv      = "DIREKTIV_GRPC_MAX_CLIENT_RCV"
	DirektivMaxServerSend     = "DIREKTIV_GRPC_MAX_SERVER_SEND"
	DirektivMaxClientSend     = "DIREKTIV_GRPC_MAX_CLIENT_SEND"

	DirektivFlowTLS  = "DIREKTIV_FLOW_TLS"
	DirektivFlowMTLS = "DIREKTIV_FLOW_MTLS"

	DirektivFluentbitTCP = "NO_FLUENTBIT_TCP"
)
