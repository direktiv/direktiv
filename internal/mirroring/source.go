package mirroring

import (
	"fmt"
	"os"
	"strings"

	"github.com/direktiv/direktiv/internal/datastore"
	"github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/plumbing"
	gitssh "github.com/go-git/go-git/v6/plumbing/transport/ssh"
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
	PullInPath(config *datastore.MirrorConfig, dir string) error
}

// GitSource implements sourcing git remote mirror into a local directory.
type GitSource struct{}

var _ Source = &GitSource{} // Ensures GitSource struct conforms to Source interface.

func (m *GitSource) PullInPath(config *datastore.MirrorConfig, dst string) error {
	uri := config.URL
	prefix := "https://"

	cloneOptions := &git.CloneOptions{
		InsecureSkipTLS: config.Insecure,
		URL:             uri,
		Progress:        os.Stdout,
		ReferenceName:   plumbing.NewBranchReferenceName(config.GitRef),
	}

	if config.AuthType == "ssh" {
		publicKeys, err := gitssh.NewPublicKeys("git", []byte(config.PrivateKey), config.PrivateKeyPassphrase)
		if err != nil {
			return err
		}
		publicKeys.HostKeyCallbackHelper = gitssh.HostKeyCallbackHelper{
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		}
		cloneOptions.Auth = publicKeys
	}

	if config.AuthType == "token" {
		uri = fmt.Sprintf("%s%s@", prefix, config.PrivateKeyPassphrase) + strings.TrimPrefix(uri, prefix)
		cloneOptions.URL = uri
	}

	_, err := git.PlainClone(dst, cloneOptions)
	if err != nil {
		return err
	}

	return nil
}
