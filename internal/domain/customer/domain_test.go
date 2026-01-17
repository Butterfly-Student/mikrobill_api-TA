package customer

import (
	"context"
	"errors"
	"testing"
	"time"

	"MikrOps/internal/model"
	outbound_port "MikrOps/internal/port/outbound"
	mock_outbound_port "MikrOps/tests/mocks/port"

	"github.com/go-routeros/routeros/v3"
	"github.com/go-routeros/routeros/v3/proto"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	. "github.com/smartystreets/goconvey/convey"
)

func TestCustomer(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mock_outbound_port.NewMockDatabasePort(ctrl)
	mockCustDB := mock_outbound_port.NewMockCustomerDatabasePort(ctrl)
	mockMikDB := mock_outbound_port.NewMockMikrotikDatabasePort(ctrl)
	mockProfDB := mock_outbound_port.NewMockProfileDatabasePort(ctrl)
	mockFactory := mock_outbound_port.NewMockMikrotikClientFactory(ctrl)
	mockClient := mock_outbound_port.NewMockMikrotikClientPort(ctrl)
	mockCache := mock_outbound_port.NewMockCachePort(ctrl)
	mockPubSub := mock_outbound_port.NewMockRedisPubSubPort(ctrl)

	mockDB.EXPECT().Customer().Return(mockCustDB).AnyTimes()
	mockDB.EXPECT().Mikrotik().Return(mockMikDB).AnyTimes()
	mockDB.EXPECT().Profile().Return(mockProfDB).AnyTimes()
	mockCache.EXPECT().PubSub().Return(mockPubSub).AnyTimes()

	domain := NewCustomerDomain(mockDB, mockFactory, mockCache)
	tenantID := uuid.New()
	ctx := context.WithValue(context.Background(), "tenant_id", tenantID)

	Convey("Test Customer Domain", t, func() {
		Convey("CreateCustomer", func() {
			input := model.CustomerInput{
				Username:    "cust1",
				Password:    "pass1",
				ProfileID:   uuid.New(),
				ServiceType: model.ServiceTypePPPoE,
				StartDate:   &[]time.Time{time.Now()}[0],
			}
			mikrotikID := uuid.New()
			activeMikrotik := &model.Mikrotik{ID: mikrotikID}
			profile := &model.ProfileWithPPPoE{
				Profile: model.Profile{ID: input.ProfileID, Name: "prof1", Price: 100000},
			}

			Convey("Success", func() {
				mockMikDB.EXPECT().GetActiveMikrotik(tenantID).Return(activeMikrotik, nil)
				mockProfDB.EXPECT().GetByMikrotikID(tenantID, mikrotikID, input.ProfileID).Return(profile, nil)
				mockCustDB.EXPECT().GetByUsername(tenantID, mikrotikID, input.Username).Return(nil, nil)
				mockFactory.EXPECT().NewClient(activeMikrotik).Return(mockClient, nil)
				mockClient.EXPECT().RunArgs("/ppp/secret/add", gomock.Any()).Return(&routeros.Reply{Done: &proto.Sentence{Map: map[string]string{"ret": "*1"}}}, nil)
				mockClient.EXPECT().Close().Return(nil)

				mockDB.EXPECT().DoInTransaction(gomock.Any()).DoAndReturn(func(f func(txDB outbound_port.DatabasePort) (interface{}, error)) (interface{}, error) {
					return f(mockDB)
				})
				mockCustDB.EXPECT().CreateCustomer(gomock.Any(), tenantID, mikrotikID, "*1").Return(&model.Customer{ID: uuid.New()}, nil)
				mockCustDB.EXPECT().CreateCustomerService(tenantID, gomock.Any(), input.ProfileID, gomock.Any(), gomock.Any(), gomock.Any()).Return(&model.CustomerService{}, nil)
				mockCustDB.EXPECT().GetByID(tenantID, gomock.Any()).Return(&model.CustomerWithService{}, nil)

				res, err := domain.CreateCustomer(ctx, input)
				So(err, ShouldBeNil)
				So(res, ShouldNotBeNil)
			})

			Convey("MikroTik Error - Rollback redundant", func() {
				mockMikDB.EXPECT().GetActiveMikrotik(tenantID).Return(activeMikrotik, nil)
				mockProfDB.EXPECT().GetByMikrotikID(tenantID, mikrotikID, input.ProfileID).Return(profile, nil)
				mockCustDB.EXPECT().GetByUsername(tenantID, mikrotikID, input.Username).Return(nil, nil)
				mockFactory.EXPECT().NewClient(activeMikrotik).Return(mockClient, nil)
				mockClient.EXPECT().RunArgs("/ppp/secret/add", gomock.Any()).Return(nil, errors.New("mikrotik error"))
				mockClient.EXPECT().Close().Return(nil)

				res, err := domain.CreateCustomer(ctx, input)
				So(err, ShouldNotBeNil)
				So(res, ShouldBeNil)
			})
		})

		Convey("GetCustomer", func() {
			custID := uuid.New()
			Convey("Success", func() {
				mockCustDB.EXPECT().GetByID(tenantID, custID).Return(&model.CustomerWithService{}, nil)
				res, err := domain.GetCustomer(ctx, custID.String())
				So(err, ShouldBeNil)
				So(res, ShouldNotBeNil)
			})
		})

		Convey("ListCustomers", func() {
			activeMikrotik := &model.Mikrotik{ID: uuid.New()}
			Convey("Success", func() {
				mockMikDB.EXPECT().GetActiveMikrotik(tenantID).Return(activeMikrotik, nil)
				mockCustDB.EXPECT().List(tenantID, activeMikrotik.ID).Return([]model.CustomerWithService{{}}, nil)
				res, err := domain.ListCustomers(ctx)
				So(err, ShouldBeNil)
				So(len(res), ShouldEqual, 1)
			})
		})

		Convey("HandlePPPoEUp", func() {
			input := model.PPPoEUpInput{User: "user1", IPAddress: "1.1.1.1", MacAddress: "AA:BB", Interface: "eth1"}
			customer := &model.Customer{ID: uuid.New(), Name: "User One"}
			Convey("Success", func() {
				mockCustDB.EXPECT().GetByPPPoEUsername(tenantID, input.User).Return(customer, nil)
				mockCustDB.EXPECT().UpdateStatus(tenantID, customer.ID, model.CustomerStatusActive, &input.IPAddress, &input.MacAddress, &input.Interface).Return(nil)
				mockPubSub.EXPECT().Publish("mikrotik:events", gomock.Any()).Return(nil)

				err := domain.HandlePPPoEUp(ctx, input)
				So(err, ShouldBeNil)
			})
		})

		Convey("HandlePPPoEDown", func() {
			input := model.PPPoEDownInput{User: "user1"}
			customer := &model.Customer{ID: uuid.New(), Name: "User One"}
			Convey("Success", func() {
				mockCustDB.EXPECT().GetByPPPoEUsername(tenantID, input.User).Return(customer, nil)
				mockCustDB.EXPECT().UpdateStatus(tenantID, customer.ID, model.CustomerStatusInactive, nil, nil, nil).Return(nil)
				mockPubSub.EXPECT().Publish("mikrotik:events", gomock.Any()).Return(nil)

				err := domain.HandlePPPoEDown(ctx, input)
				So(err, ShouldBeNil)
			})
		})
	})
}


