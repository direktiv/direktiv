package mirror

import (
	"fmt"
	"os"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	gitssh "github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"golang.org/x/crypto/ssh"
)

// Source is an interface that represent a mirror source (git repo is a valid mirror source example).
// In Direktiv, a Mirror is a directory of files that sits somewhere (local or remote) and a user wants to mirror (copy)
// his direktiv namespace files from it.
// Source knows how to access the mirror files (connecting to a remote server in case of git) and copy the files in
// the user's direktiv namespace. Parameter 'settings' is used to configure the sourcing (pulling) mirror process.
// Parameter 'dir' specifies the directory where Source should copy the mirror to.
type Source interface {
	// PullInPath pulls (copies) mirror into local directory specified by 'dir' parameter.
	PullInPath(config *Config, dir string) error
}

// MockedSource mocks Source interface.
type MockedSource struct {
	Paths map[string]string
}

var _ Source = &MockedSource{} // Ensures MockedSource struct conforms to Source interface.

//nolint:revive
func (m MockedSource) PullInPath(config *Config, dst string) error {
	for k, v := range m.Paths {
		//nolint:gomnd
		if err := os.WriteFile(dst+k, []byte(v), 0o600); err != nil {
			return err
		}
	}

	return nil
}

// GitSource implements sourcing git remote mirror into a local directory.
type GitSource struct{}

var _ Source = &GitSource{} // Ensures GitSource struct conforms to Source interface.

func (m *GitSource) PullInPath(config *Config, dst string) error {
	uri := config.URL
	prefix := "https://"

	cloneOptions := &git.CloneOptions{
		InsecureSkipTLS: true, // This has to be a toggle in the UI
		URL:             uri,
		Progress:        os.Stdout,
		ReferenceName:   plumbing.NewBranchReferenceName(config.GitRef),
	}

	// https with access token case. Put passphrase inside the git url.
	if strings.HasPrefix(uri, prefix) && len(config.PrivateKeyPassphrase) > 0 {
		if !strings.Contains(uri, "@") {
			uri = fmt.Sprintf("%s%s@", prefix, config.PrivateKeyPassphrase) + strings.TrimPrefix(uri, prefix)
			cloneOptions.URL = uri
		}
	}

	// ssh case. Configure cloneOptions.Auth field.
	if !strings.HasPrefix(uri, prefix) {
		publicKeys, err := gitssh.NewPublicKeys("git", []byte(config.PrivateKey), config.PrivateKeyPassphrase)
		if err != nil {
			return err
		}
		publicKeys.HostKeyCallbackHelper = gitssh.HostKeyCallbackHelper{
			//nolint:gosec
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		}
		cloneOptions.Auth = publicKeys
	}

	_, err := git.PlainClone(dst, false, cloneOptions)
	if err != nil {
		return err
	}

	return nil
}
