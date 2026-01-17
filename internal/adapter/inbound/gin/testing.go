package gin_inbound_adapter

import (
	"MikrOps/internal/domain"
	inbound_port "MikrOps/internal/port/inbound"
)

type testingAdapter struct {
	domain domain.Domain
}

func NewTestingAdapter(
	domain domain.Domain,
) inbound_port.TestingHttpPort {
	return &testingAdapter{
		domain: domain,
	}
}

