package dlog

type BrokerManager struct {
	brokerMap map[string]*Broker
}

func (l *BrokerManager) SetBroker(instance string) (*Broker, bool) {

	// Create Broker if it doest exist
	if _, ok := l.brokerMap[instance]; !ok {
		l.brokerMap[instance] = newBroker()
		go l.brokerMap[instance].Start()
		return l.brokerMap[instance], true
	}

	return l.brokerMap[instance], false

}

func (l *BrokerManager) GetBroker(instance string) (*Broker, bool) {

	if b, ok := l.brokerMap[instance]; ok {
		return b, true
	}

	return nil, false

}

func (l *BrokerManager) DeleteBroker(instance string) bool {

	// Create Broker if it doest exist
	if _, ok := l.brokerMap[instance]; ok {
		delete(l.brokerMap, instance)
		return true
	}

	return false
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
