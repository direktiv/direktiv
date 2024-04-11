package mirror

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"strings"

	"github.com/direktiv/direktiv/pkg/refactor/datastore"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	gitssh "github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"golang.org/x/crypto/ssh"
)

type Source interface {
	FS() fs.FS
	Free() error
	Notes() map[string]string
}

type DirectorySource struct {
	fs fs.FS
}

var _ Source = &DirectorySource{}

func NewDirectorySource(dir string) *DirectorySource {
	return &DirectorySource{
		fs: os.DirFS(dir),
	}
}

func (src *DirectorySource) FS() fs.FS {
	return src.fs
}

func (src *DirectorySource) Free() error {
	return nil
}

func (src *DirectorySource) Notes() map[string]string {
	return make(map[string]string)
}

type GitSourceConfig struct {
	URL             string
	GitRef          string
	InsecureSkipTLS bool
	TempDir         string
}

type gitSource struct {
	*DirectorySource
	path  string
	conf  GitSourceConfig
	notes map[string]string
}

var _ Source = &gitSource{}

func basicCloneOpts(conf GitSourceConfig) *git.CloneOptions {
	return &git.CloneOptions{
		InsecureSkipTLS: conf.InsecureSkipTLS,
		URL:             conf.URL,
		Progress:        nil,
		ReferenceName:   plumbing.NewBranchReferenceName(conf.GitRef),
	}
}

func clone(conf GitSourceConfig, cloneOpts *git.CloneOptions) (Source, error) {
	path, err := os.MkdirTemp(conf.TempDir, "direktiv_clone_*")
	if err != nil {
		return nil, err
	}

	_, err = git.PlainClone(path, false, cloneOpts)
	if err != nil {
		return nil, err
	}

	return &gitSource{
		conf:            conf,
		path:            path,
		DirectorySource: NewDirectorySource(path),
		notes: map[string]string{
			"ref": conf.GitRef,
			"url": conf.URL,
		},
	}, nil
}

func (src *gitSource) FS() fs.FS {
	return src.fs
}

func (src *gitSource) Free() error {
	err := os.RemoveAll(src.path)
	if err != nil {
		return err
	}

	return nil
}

func (src *gitSource) Notes() map[string]string {
	return src.notes
}

func NewGitSourceNoAuth(conf GitSourceConfig) (Source, error) {
	return clone(conf, basicCloneOpts(conf))
}

type GitSourceTokenAuthConf struct {
	Token string
}

func newGitSourceToken(conf GitSourceConfig, auth GitSourceTokenAuthConf) (Source, error) {
	cloneOpts := basicCloneOpts(conf)

	prefix := "https://"
	cloneOpts.URL = fmt.Sprintf("%s%s@", prefix, auth.Token) + strings.TrimPrefix(conf.URL, prefix)

	return clone(conf, cloneOpts)
}

type GitSourceSSHAuthConf struct {
	PublicKey            string
	PrivateKey           string
	PrivateKeyPassphrase string
}

func NewGitSourceSSH(conf GitSourceConfig, auth GitSourceSSHAuthConf) (Source, error) {
	cloneOpts := basicCloneOpts(conf)

	publicKeys, err := gitssh.NewPublicKeys("git", []byte(auth.PrivateKey), auth.PrivateKeyPassphrase)
	if err != nil {
		return nil, err
	}
	publicKeys.HostKeyCallbackHelper = gitssh.HostKeyCallbackHelper{
		//nolint:gosec
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	cloneOpts.Auth = publicKeys

	return clone(conf, cloneOpts)
}

func GetSource(_ context.Context, cfg *datastore.MirrorConfig) (Source, error) {
	insecureSkipTLS := cfg.Insecure
	tempDir := ""

	if cfg.URL == "" {
		return nil, fmt.Errorf("URL is missing in the configuration")
	}
	if cfg.GitRef == "" {
		return nil, fmt.Errorf("GitRef is missing in the configuration")
	}

	if cfg.PrivateKeyPassphrase == "" && cfg.PublicKey == "" && cfg.PrivateKey == "" {
		return NewGitSourceNoAuth(GitSourceConfig{
			URL:             cfg.URL,
			GitRef:          cfg.GitRef,
			InsecureSkipTLS: insecureSkipTLS,
			TempDir:         tempDir,
		})
	}

	if strings.HasPrefix(cfg.URL, "http") {
		if cfg.PrivateKeyPassphrase == "" {
			return nil, fmt.Errorf("PrivateKeyPassphrase field has to be filled with the auth-token. This is required for token-based source")
		}

		return newGitSourceToken(GitSourceConfig{
			URL:             cfg.URL,
			GitRef:          cfg.GitRef,
			InsecureSkipTLS: cfg.Insecure,
			TempDir:         tempDir,
		}, GitSourceTokenAuthConf{
			Token: cfg.AuthToken,
		})
	}
	if cfg.PrivateKey != "" || cfg.PublicKey != "" {
		return NewGitSourceSSH(GitSourceConfig{
			URL:             cfg.URL,
			GitRef:          cfg.GitRef,
			InsecureSkipTLS: insecureSkipTLS,
			TempDir:         tempDir,
		}, GitSourceSSHAuthConf{
			PrivateKey:           cfg.PrivateKey,
			PublicKey:            cfg.PublicKey,
			PrivateKeyPassphrase: cfg.PrivateKeyPassphrase,
		})
	}

	return nil, fmt.Errorf("could not detect the git auth mode to use")
}
