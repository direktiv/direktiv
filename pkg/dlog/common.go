package dlog

import (
	"context"
	"io"

	"github.com/inconshreveable/log15"
)

type Logger interface {
	log15.Logger
	io.Closer
}

type Log interface {
	LoggerFunc(namespace, instance string) (Logger, error)
	NamespaceLogger(namespace string) (Logger, error)
	QueryLogs(ctx context.Context, instance string, limit, offset int) (QueryReponse, error)
	StreamLogs(ctx context.Context, instance string) (chan interface{}, error)
	DeleteNamespaceLogs(namespace string) error
	DeleteInstanceLogs(instance string) error
}

type LogEntry struct {
	Level     string            `json:"lvl"`
	Timestamp int64             `json:"time"`
	Message   string            `json:"msg"`
	Context   map[string]string `json:"ctx"`
}

type QueryReponse struct {
	Count  int `json:"count"`
	Limit  int `json:"limit,omitempty"`
	Offset int `json:"offset"`
	// Data   []map[string]interface{} `json:"data"`
	Logs []LogEntry `json:"data"`
}

type Broker struct {
	stopCh    chan struct{}
	publishCh chan interface{}
	subCh     chan chan interface{}
	unsubCh   chan chan interface{}
}

func NewBroker() *Broker {
	return &Broker{
		stopCh:    make(chan struct{}),
		publishCh: make(chan interface{}, 1),
		subCh:     make(chan chan interface{}, 1),
		unsubCh:   make(chan chan interface{}, 1),
	}
}

func (b *Broker) Start() {
	subs := map[chan interface{}]struct{}{}
	for {
		select {
		case <-b.stopCh:
			return
		case msgCh := <-b.subCh:
			subs[msgCh] = struct{}{}
		case msgCh := <-b.unsubCh:
			delete(subs, msgCh)
		case msg := <-b.publishCh:
			for msgCh := range subs {
				select {
				case msgCh <- msg:
				default:
				}
			}
		}
	}
}

func (b *Broker) Stop() {
	close(b.stopCh)
}

func (b *Broker) Subscribe() chan interface{} {
	msgCh := make(chan interface{}, 5)
	b.subCh <- msgCh
	return msgCh
}

func (b *Broker) Unsubscribe(msgCh chan interface{}) {
	b.unsubCh <- msgCh
}
