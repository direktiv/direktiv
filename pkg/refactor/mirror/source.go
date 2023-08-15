package mirror

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	gitssh "github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"golang.org/x/crypto/ssh"
)

type Source interface {
	FS() fs.FS
	Unwrap() Source
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

func (src *DirectorySource) Unwrap() Source {
	return src
}

func (src *DirectorySource) Free() error {
	return nil
}

func (src *DirectorySource) Notes() map[string]string {
	return make(map[string]string)
}

type GitSourceConfig struct {
	URL    string
	GitRef string
}

type GitSourceOptions struct {
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

func basicCloneOpts(conf GitSourceConfig, opts GitSourceOptions) *git.CloneOptions {
	return &git.CloneOptions{
		InsecureSkipTLS: opts.InsecureSkipTLS,
		URL:             conf.URL,
		Progress:        nil,
		ReferenceName:   plumbing.NewBranchReferenceName(conf.GitRef),
	}
}

func clone(conf GitSourceConfig, cloneOpts *git.CloneOptions, opts GitSourceOptions) (Source, error) {
	path, err := os.MkdirTemp(opts.TempDir, "direktiv_clone_*")
	if err != nil {
		return nil, err
	}

	repo, err := git.PlainClone(path, false, cloneOpts)
	if err != nil {
		return nil, err
	}

	ref, err := repo.Head()
	if err != nil {
		return nil, err
	}

	hash := ref.Hash()

	return &gitSource{
		conf:            conf,
		path:            path,
		DirectorySource: NewDirectorySource(path),
		notes: map[string]string{
			"commit_hash": hash.String(),
			"ref":         conf.GitRef,
			"url":         conf.URL,
		},
	}, nil
}

func (src *gitSource) FS() fs.FS {
	return src.fs
}

func (src *gitSource) Unwrap() Source {
	return src
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

func NewGitSourceNoAuth(conf GitSourceConfig, opts GitSourceOptions) (Source, error) {
	return clone(conf, basicCloneOpts(conf, opts), opts)
}

type GitSourceTokenAuthConf struct {
	Token string
}

func newGitSourceToken(conf GitSourceConfig, auth GitSourceTokenAuthConf, opts GitSourceOptions) (Source, error) {
	cloneOpts := basicCloneOpts(conf, opts)

	prefix := "https://"
	cloneOpts.URL = fmt.Sprintf("%s%s@", prefix, auth.Token) + strings.TrimPrefix(conf.URL, prefix)

	return clone(conf, cloneOpts, opts)
}

type GitSourceSSHAuthConf struct {
	PublicKey            string
	PrivateKey           string
	PrivateKeyPassphrase string
}

func NewGitSourceSSH(conf GitSourceConfig, auth GitSourceSSHAuthConf, opts GitSourceOptions) (Source, error) {
	cloneOpts := basicCloneOpts(conf, opts)

	publicKeys, err := gitssh.NewPublicKeys("git", []byte(auth.PrivateKey), auth.PrivateKeyPassphrase)
	if err != nil {
		return nil, err
	}
	publicKeys.HostKeyCallbackHelper = gitssh.HostKeyCallbackHelper{
		//nolint:gosec
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	cloneOpts.Auth = publicKeys

	return clone(conf, cloneOpts, opts)
}

func (cfg *Config) GetSource(_ context.Context) (Source, error) {
	insecureSkipTLS := true
	tempDir := ""

	if cfg.PrivateKey == "" {
		return NewGitSourceNoAuth(GitSourceConfig{
			URL:    cfg.URL,
			GitRef: cfg.GitRef,
		}, GitSourceOptions{
			InsecureSkipTLS: insecureSkipTLS,
			TempDir:         tempDir,
		})
	}

	if strings.HasPrefix(cfg.URL, "http") {
		return newGitSourceToken(GitSourceConfig{
			URL:    cfg.URL,
			GitRef: cfg.GitRef,
		}, GitSourceTokenAuthConf{
			Token: cfg.PrivateKey,
		}, GitSourceOptions{
			InsecureSkipTLS: insecureSkipTLS,
			TempDir:         tempDir,
		})
	}

	return NewGitSourceSSH(GitSourceConfig{
		URL:    cfg.URL,
		GitRef: cfg.GitRef,
	}, GitSourceSSHAuthConf{
		PrivateKey:           cfg.PrivateKey,
		PublicKey:            cfg.PublicKey,
		PrivateKeyPassphrase: cfg.PrivateKeyPassphrase,
	}, GitSourceOptions{
		InsecureSkipTLS: insecureSkipTLS,
		TempDir:         tempDir,
	})
}
