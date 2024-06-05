package tsengine

type Config struct {
	BaseDir string `env:"DIREKTIV_JSENGINE_BASEDIR,notEmpty" envDefault:"/direktiv"`

	LogLevel  string `env:"DIREKTIV_JS_ENGINE_LOGLEVEL" envDefault:"info"`
	FlowPath  string `env:"DIREKTIV_JSENGINE_FLOWPATH,notEmpty"`
	Namespace string `env:"DIREKTIV_JSENGINE_NAMESPACE,notEmpty"`

	SelfCopy     string `env:"DIREKTIV_JSENGINE_SELFCOPY"`
	SelfCopyExit bool   `env:"DIREKTIV_JSENGINE_SELFCOPY_EXIT"`

	DBConfig  string `env:"DIREKTIV_DB"`
	SecretKey string `env:"DIREKTIV_SECRET_KEY,notEmpty"`
}
