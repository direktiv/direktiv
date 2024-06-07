package pubsub

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/direktiv/direktiv/pkg/core"
	"github.com/google/uuid"
	"k8s.io/apimachinery/pkg/util/json"
)

type Bus struct {
	coreBus CoreBus

	subscribers  sync.Map
	fingerprints sync.Map
}

const defaultDebouncePublishDuration = 200 * time.Millisecond

func NewBus(coreBus CoreBus) *Bus {
	return &Bus{
		coreBus: coreBus,
	}
}

func (p *Bus) Loop(circuit *core.Circuit) error {
	return p.coreBus.Loop(circuit.Done(), func(channel string, data string) {
		p.subscribers.Range(func(key, f any) bool {
			k, _ := key.(string)
			h, _ := f.(func(data string))

			if strings.HasPrefix(k, channel) {
				go h(data)
			}

			return true
		})
	})
}

func (p *Bus) Publish(event any) error {
	channel := reflect.TypeOf(event).String()
	data, err := json.Marshal(event)
	if err != nil {
		panic("Logic error: " + err.Error())
	}

	return p.coreBus.Publish(channel, string(data))
}

func (p *Bus) debouncedPublishWithInterval(i time.Duration, event any) error {
	// This function works by associating input with a signature, sleep for a duration and nly publish the message
	// when the signature matches.

	channel := reflect.TypeOf(event).String()
	data, err := json.Marshal(event)
	if err != nil {
		panic("Logic error: " + err.Error())
	}

	input := fmt.Sprintf("%d_%s_%s", i, channel, data)
	signature := uuid.New()
	p.fingerprints.Store(input, signature)

	go func() {
		time.Sleep(i)
		currentSignature, _ := p.fingerprints.Load(input)
		// When signature matches, this means no later async publish was recorded.
		if signature == currentSignature {
			_ = p.coreBus.Publish(channel, string(data))
		}
	}()

	return nil
}

// DebouncedPublish prevents multiple concussive publishes of the same input during an interval.
func (p *Bus) DebouncedPublish(event any) error {
	return p.debouncedPublishWithInterval(defaultDebouncePublishDuration, event)
}

func (p *Bus) Subscribe(channel any, handler func(data string)) {
	if channel == nil {
		panic("nil channel")
	}
	if !reflect.TypeOf(channel).Comparable() {
		panic("channel is not comparable")
	}
	channelStr := reflect.TypeOf(channel).String()

	p.subscribers.Store(fmt.Sprintf("%s_%s", channelStr, uuid.New().String()), handler)
	err := p.coreBus.Listen(channelStr)
	if err != nil {
		panic("TODO: handle this pubsub error: " + err.Error())
	}
}

type FileSystemChangeEvent struct {
	Action       string
	Namespace    string
	NamespaceID  uuid.UUID
	FileType     string
	FilePath     string
	OldPath      string
	DeleteFileID uuid.UUID
}

type NamespacesChangeEvent struct {
	Name   string
	Action string
}

type InstanceMessageEvent struct {
	Message string
}
