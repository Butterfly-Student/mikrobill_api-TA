package gin_inbound_adapter

import (
	"prabogo/internal/domain"
	inbound_port "prabogo/internal/port/inbound"
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
