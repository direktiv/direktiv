package dlog

import "sync"

type BrokerManager struct {
	brokerMap map[string]*Broker
	sync.Mutex
}

func (bm *BrokerManager) SetBroker(instance string) (*Broker, bool) {

	// Create Broker if it doest exist
	if _, ok := bm.brokerMap[instance]; !ok {
		bm.Lock()
		bm.brokerMap[instance] = newBroker()
		go bm.brokerMap[instance].Start()
		bm.Unlock()
		return bm.brokerMap[instance], true
	}

	return bm.brokerMap[instance], false

}

func (bm *BrokerManager) GetBroker(instance string) (*Broker, bool) {
	b, ok := bm.brokerMap[instance]
	return b, ok
}

func (bm *BrokerManager) DeleteBroker(instance string) bool {

	// Delete Broker if it exist
	if _, ok := bm.brokerMap[instance]; ok {
		bm.Lock()
		delete(bm.brokerMap, instance)
		bm.Unlock()
		return true
	}

	return false
}

// Publish to broker instance, if instance does not exists, it will get created
func (bm *BrokerManager) Publish(instance string, entry LogEntry) error {
	b, _ := bm.SetBroker(instance)
	b.publishCh <- entry
	return nil
}

func NewBrokerManager() *BrokerManager {
	return &BrokerManager{
		brokerMap: make(map[string]*Broker),
	}
}

type Broker struct {
	stopCh    chan struct{}
	publishCh chan interface{}
	subCh     chan chan interface{}
	unsubCh   chan chan interface{}
}

func newBroker() *Broker {
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

func (b *Broker) Publish(entry LogEntry) error {
	b.publishCh <- entry
	return nil
}
