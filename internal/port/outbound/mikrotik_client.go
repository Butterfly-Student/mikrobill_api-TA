package outbound_port

//go:generate mockgen -source=mikrotik_client.go -destination=./../../../tests/mocks/port/mock_mikrotik_client.go

import (
	"MikrOps/internal/model"

	"github.com/go-routeros/routeros/v3"
	"github.com/go-routeros/routeros/v3/proto"
)

type MikrotikClientPort interface {
	Run(sentence ...string) (*routeros.Reply, error)
	RunArgs(sentence string, args map[string]string) (*routeros.Reply, error)
	ListenArgs(sentence string, args map[string]string) (<-chan *proto.Sentence, error)
	Close() error
}

type MikrotikClientFactory interface {
	NewClient(mikrotik *model.Mikrotik) (MikrotikClientPort, error)
}

