package tsengine

type Config struct {
	// Initializer string `env:"DIREKTIV_JSENGINE_INIT" envDefault:"db"`
	BaseDir string `env:"DIREKTIV_JSENGINE_BASEDIR,notEmpty" envDefault:"/direktiv"`

	LogLevel  string `env:"DIREKTIV_JS_ENGINE_LOGLEVEL" envDefault:"info"`
	FlowPath  string `env:"DIREKTIV_JSENGINE_FLOWPATH,notEmpty"`
	Namespace string `env:"DIREKTIV_JSENGINE_NAMESPACE,notEmpty"`

	Image string `env:"DIREKTIV_JSENGINE_ENGINE_IMAGE" envDefault:"localhost:5000/image"`

	SelfCopy string `env:"DIREKTIV_JSENGINE_SELFCOPY"`

	DBConfig string `env:"DIREKTIV_DB"`
}
