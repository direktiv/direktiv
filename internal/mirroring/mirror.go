package mirroring

import (
	"context"
	"fmt"

	"github.com/direktiv/direktiv/internal/datastore"
	"github.com/direktiv/direktiv/internal/datastore/datasql"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Manager struct{}

func NewManager() *Manager {
	return &Manager{}
}

type mirrorJob struct {
	manager *Manager
	db      *gorm.DB

	process   *datastore.MirrorProcess
	mirrorDir string

	err error
}

func (m *Manager) Exec(ctx context.Context, db *gorm.DB, cfg *datastore.MirrorConfig, typ string) (*datastore.MirrorProcess, error) {
	pro := &datastore.MirrorProcess{
		ID:        uuid.New(),
		Namespace: cfg.Namespace,
		Typ:       typ,
		Status:    datastore.ProcessStatusPending,
	}

	var err error
	pro, err = datasql.NewStore(db).Mirror().CreateProcess(ctx, pro)
	if err != nil {
		return nil, fmt.Errorf("datastore create mirror process: %w", err)
	}

	job := &mirrorJob{
		manager: m,
		db:      db,
		process: pro,
	}

	go func() {
		fmt.Printf(">>>>> starting mirror process: %v\n", pro)

		job.setProcessStatue(datastore.ProcessStatusExecuting)
		job.createTempDir()
		job.setProcessStatue(datastore.ProcessStatusComplete)

		if job.err == nil {
			return
		}

		job.setProcessStatue(datastore.ProcessStatusFailed)

		err = job.err
		if err != nil {
			fmt.Printf(">>>>> error creating mirror process: %v\n", err)
		}
	}()

	return pro, nil
}

func (j *mirrorJob) setProcessStatue(status string) {
	if j.err != nil {
		return
	}

	var err error
	j.process.Status = status
	j.process, err = datasql.NewStore(j.db).Mirror().UpdateProcess(context.Background(), j.process)

	if err != nil {
		j.err = fmt.Errorf("datastore set mirror process status: %w", err)
		return
	}
}

func (j *mirrorJob) createTempDir() {
	if j.err != nil {
		return
	}
}
