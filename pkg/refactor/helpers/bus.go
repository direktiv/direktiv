package helpers

import (
	"encoding/json"

	"github.com/direktiv/direktiv/pkg/refactor/filestore"
	"github.com/direktiv/direktiv/pkg/refactor/pubsub"
)

func PublishEventDirektivFileChange(bus *pubsub.Bus, fileType filestore.FileType, topic string, event *pubsub.FileChangeEvent) error {
	eventData, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return bus.DebouncedPublish(string(fileType)+"_"+topic, string(eventData))
}
