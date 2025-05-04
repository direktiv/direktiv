package datasql

import (
	"context"
	"fmt"

	"github.com/direktiv/direktiv/pkg/datastore"
	"gorm.io/gorm"
)

type sqlHeartBeatsStore struct {
	db *gorm.DB
}

func (s *sqlHeartBeatsStore) Set(ctx context.Context, heartBeat *datastore.HeartBeat) error {
	res := s.db.WithContext(ctx).Exec(`
				INSERT INTO system_heart_beats("group", "key") VALUES(?,?) 
				ON CONFLICT ("group", "key") DO UPDATE SET updated_at = NOW()`, heartBeat.Group, heartBeat.Key)

	if res.Error != nil {
		return res.Error
	}

	return nil
}

func (s *sqlHeartBeatsStore) Since(ctx context.Context, group string, secondsAgo int) ([]*datastore.HeartBeat, error) {
	var list []*datastore.HeartBeat

	secondsAgoStr := fmt.Sprintf("'%d seconds'", secondsAgo)

	res := s.db.WithContext(ctx).Raw(`
							SELECT "group", "key", created_at, updated_at 
							FROM system_heart_beats
							WHERE "group"=? AND updated_at + INTERVAL `+secondsAgoStr+` > NOW()`, group).
		Find(&list)
	if res.Error != nil {
		return nil, res.Error
	}

	return list, nil
}

var _ datastore.HeartBeatsStore = &sqlHeartBeatsStore{}
