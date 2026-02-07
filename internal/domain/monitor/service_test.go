package monitor

import (
	"context"
	"testing"
	"time"

	"MikrOps/internal/model"
	mock_outbound_port "MikrOps/tests/mocks/port"

	"github.com/go-routeros/routeros/v3"
	"github.com/go-routeros/routeros/v3/proto"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	. "github.com/smartystreets/goconvey/convey"
)

func TestMonitor(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mock_outbound_port.NewMockDatabasePort(ctrl)
	mockCustDB := mock_outbound_port.NewMockCustomerDatabasePort(ctrl)
	mockMikDB := mock_outbound_port.NewMockMikrotikDatabasePort(ctrl)
	mockFactory := mock_outbound_port.NewMockMikrotikClientFactory(ctrl)
	mockClient := mock_outbound_port.NewMockMikrotikClientPort(ctrl)
	mockCache := mock_outbound_port.NewMockCachePort(ctrl)

	mockDB.EXPECT().Customer().Return(mockCustDB).AnyTimes()
	mockDB.EXPECT().Mikrotik().Return(mockMikDB).AnyTimes()

	domain := NewMonitorDomain(mockDB, mockFactory, mockCache)
	ctx := context.Background()

	Convey("Test Monitor Service", t, func() {
		Convey("PingCustomer", func() {
			custID := uuid.New()
			ip := "192.168.1.1"
			customer := &model.Customer{
				ID:          custID.String(),
				ServiceType: "pppoe",
				AssignedIP:  &ip,
			}

			Convey("Success", func() {
				mockCustDB.EXPECT().GetByID(gomock.Any(), custID).Return(customer, nil)
				mockMikDB.EXPECT().GetActiveMikrotik(gomock.Any()).Return(&model.Mikrotik{}, nil)
				mockFactory.EXPECT().NewClient(gomock.Any()).Return(mockClient, nil)
				mockClient.EXPECT().RunArgs("/ping", gomock.Any()).Return(&routeros.Reply{
					Done: &proto.Sentence{Map: map[string]string{
						"sent": "3", "received": "3", "packet-loss": "0",
					}},
				}, nil)
				mockClient.EXPECT().Close().Return(nil)

				res, err := domain.PingCustomer(ctx, custID.String())
				So(err, ShouldBeNil)
				So(res["is_reachable"], ShouldBeTrue)
				So(res["target"], ShouldEqual, ip)
			})
		})

		Convey("StreamTraffic - Setup", func() {
			custID := uuid.New()
			iface := "pppoe-out1"
			customer := &model.Customer{
				ID:        custID.String(),
				Name:      "Test",
				Interface: &iface,
			}

			Convey("Success - Starting new monitor", func() {
				mockCustDB.EXPECT().GetByID(gomock.Any(), custID).Return(customer, nil)

				// Monitor loop calls
				mockMikDB.EXPECT().GetActiveMikrotik(gomock.Any()).Return(nil, nil).AnyTimes()

				res, err := domain.StreamTraffic(ctx, custID.String())
				So(err, ShouldBeNil)
				So(res, ShouldNotBeNil)

				// Wait a bit for goroutine to start and hit the mock
				time.Sleep(10 * time.Millisecond)

				StopMonitoring(custID.String())
			})
		})
	})
}

func StopMonitoring(customerID string) {
	mu.Lock()
	defer mu.Unlock()
	if m, ok := activeMonitors[customerID]; ok {
		m.Cancel()
		delete(activeMonitors, customerID)
	}
}
