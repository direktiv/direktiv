package api

// Config ..
type Config struct {
	Ingress struct {
		Endpoint string
		TLS      struct {
			Enabled  bool
			Secure   bool
			CertsDir string
		}
	}
	Server struct {
		Bind string
		TLS  struct {
			Enabled  bool
			CertsDir string
		}
	}
}
