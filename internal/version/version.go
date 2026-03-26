package version

var (
	Version string
	GitSha  string
)

func IsDev() bool {
	return Version == "" || GitSha == ""
}
