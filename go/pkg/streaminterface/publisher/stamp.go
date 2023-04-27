package publisher

import "nathejk.dk/pkg/streaminterface"

type publisher struct {
	p streaminterface.Publisher
}

func NewMetadataStamp(p streaminterface.Publisher, key string, value interface{}) *publisher {
	return &publisher{p: p}
}

func (p *publisher) Publish() {
}
