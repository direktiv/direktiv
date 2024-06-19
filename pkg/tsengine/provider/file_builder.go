package provider

import (
	"context"
	"log/slog"
	"path/filepath"

	"github.com/direktiv/direktiv/pkg/compiler"
	"github.com/fsnotify/fsnotify"
)

type FileBuilder struct {
	files     []compiler.File
	baseFS    string
	provider  FileWriter
	namespace string
}

func NewFileBuilder(provider FileWriter, namespace string, fi compiler.FlowInformation, baseFS string) *FileBuilder {
	return &FileBuilder{
		files:     fi.Files,
		baseFS:    baseFS,
		namespace: namespace,
		provider:  provider,
	}
}

func (b *FileBuilder) Build(ctx context.Context) FileWatcher {
	// read files
	b.watcher(ctx)
	return FileWatcher{sync: b.watcher}
}

func (b *FileBuilder) watcher(ctx context.Context) {
	_ = ctx
	for a := range b.files {
		file := b.files[a]
		if file.Scope == "shared" {
			err := b.provider.WriteFile(ctx, b.namespace, file)
			if err != nil {
				slog.Error("TODO in watcher", "error", err)
			}
		}
	}
}

type FileWatcher struct {
	sync func(context.Context)
}

func (f FileWatcher) Watch(ctx context.Context, flow string) {

	// dir to watch
	dir := filepath.Dir(flow)

	// file to watch
	file := filepath.Base(flow)

	slog.Info("watching flow", slog.String("flow", flow))

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		panic(err)
	}
	defer watcher.Close()

	// listening for flow changes
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					continue
				}
				if filepath.Base(event.Name) == file && event.Has(fsnotify.Write) {
					slog.Info("updating flow")
					f.sync(ctx)
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					continue
				}
				slog.Error("error occurred watching flow file:", slog.Any("error", err))
			}
		}
	}()
	err = watcher.Add(dir)
	if err != nil {
		panic(err)
	}
	<-make(chan struct{})
}
