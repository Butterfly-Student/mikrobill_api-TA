package profile

import (
	"context"
	"testing"

	"prabogo/internal/model"
	outbound_port "prabogo/internal/port/outbound"
	mock_outbound_port "prabogo/tests/mocks/port"

	"github.com/go-routeros/routeros/v3"
	"github.com/go-routeros/routeros/v3/proto"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	. "github.com/smartystreets/goconvey/convey"
)

func TestProfile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mock_outbound_port.NewMockDatabasePort(ctrl)
	mockProfDB := mock_outbound_port.NewMockProfileDatabasePort(ctrl)
	mockMikDB := mock_outbound_port.NewMockMikrotikDatabasePort(ctrl)
	mockFactory := mock_outbound_port.NewMockMikrotikClientFactory(ctrl)
	mockClient := mock_outbound_port.NewMockMikrotikClientPort(ctrl)

	mockDB.EXPECT().Profile().Return(mockProfDB).AnyTimes()
	mockDB.EXPECT().Mikrotik().Return(mockMikDB).AnyTimes()

	domain := NewProfileDomain(mockDB, mockFactory)
	tenantID := uuid.New()
	ctx := context.WithValue(context.Background(), "tenant_id", tenantID)

	Convey("Test Profile Domain", t, func() {
		Convey("CreateProfile", func() {
			input := model.ProfileInput{Name: "Prof1", Price: 50000}
			mikrotikID := uuid.New()
			activeMikrotik := &model.Mikrotik{ID: mikrotikID}

			Convey("Success", func() {
				mockMikDB.EXPECT().GetActiveMikrotik(tenantID).Return(activeMikrotik, nil)
				mockDB.EXPECT().DoInTransaction(gomock.Any()).DoAndReturn(func(f func(txDB outbound_port.DatabasePort) (interface{}, error)) (interface{}, error) {
					return f(mockDB)
				})
				mockProfDB.EXPECT().CreateProfile(tenantID, gomock.Any(), mikrotikID).Return(&model.Profile{ID: uuid.New()}, nil)
				mockProfDB.EXPECT().CreateProfilePPPoE(tenantID, gomock.Any(), gomock.Any()).Return(nil)
				mockFactory.EXPECT().NewClient(activeMikrotik).Return(mockClient, nil)
				mockClient.EXPECT().RunArgs("/ppp/profile/add", gomock.Any()).Return(&routeros.Reply{Done: &proto.Sentence{Map: map[string]string{"ret": "*A"}}}, nil)
				mockClient.EXPECT().Close().Return(nil)
				mockProfDB.EXPECT().UpdateMikrotikObjectID(tenantID, gomock.Any(), "*A").Return(nil)
				mockProfDB.EXPECT().GetByID(tenantID, gomock.Any()).Return(&model.ProfileWithPPPoE{}, nil)

				res, err := domain.CreateProfile(ctx, input)
				So(err, ShouldBeNil)
				So(res, ShouldNotBeNil)
			})

			Convey("No Active Mikrotik", func() {
				mockMikDB.EXPECT().GetActiveMikrotik(tenantID).Return(nil, nil)
				res, err := domain.CreateProfile(ctx, input)
				So(err, ShouldNotBeNil)
				So(res, ShouldBeNil)
			})
		})

		Convey("GetProfile", func() {
			id := uuid.New()
			Convey("Success", func() {
				mockProfDB.EXPECT().GetByID(tenantID, id).Return(&model.ProfileWithPPPoE{}, nil)
				res, err := domain.GetProfile(ctx, id.String())
				So(err, ShouldBeNil)
				So(res, ShouldNotBeNil)
			})
		})

		Convey("UpdateProfile", func() {
			id := uuid.New()
			input := model.ProfileInput{Name: "Updated"}
			mikrotikID := uuid.New()
			activeMikrotik := &model.Mikrotik{ID: mikrotikID}
			existing := &model.ProfileWithPPPoE{
				Profile: model.Profile{MikrotikObjectID: "*A"},
			}

			Convey("Success", func() {
				mockProfDB.EXPECT().GetByID(tenantID, id).Return(existing, nil)
				mockMikDB.EXPECT().GetActiveMikrotik(tenantID).Return(activeMikrotik, nil)
				mockDB.EXPECT().DoInTransaction(gomock.Any()).DoAndReturn(func(f func(txDB outbound_port.DatabasePort) (interface{}, error)) (interface{}, error) {
					return f(mockDB)
				})
				mockProfDB.EXPECT().Update(tenantID, id, gomock.Any()).Return(nil)
				mockFactory.EXPECT().NewClient(activeMikrotik).Return(mockClient, nil)
				mockClient.EXPECT().RunArgs("/ppp/profile/set", gomock.Any()).Return(&routeros.Reply{}, nil)
				mockClient.EXPECT().Close().Return(nil)
				mockProfDB.EXPECT().GetByID(tenantID, id).Return(&model.ProfileWithPPPoE{}, nil)

				res, err := domain.UpdateProfile(ctx, id.String(), input)
				So(err, ShouldBeNil)
				So(res, ShouldNotBeNil)
			})
		})

		Convey("DeleteProfile", func() {
			id := uuid.New()
			mikrotikID := uuid.New()
			activeMikrotik := &model.Mikrotik{ID: mikrotikID}
			existing := &model.ProfileWithPPPoE{
				Profile: model.Profile{MikrotikObjectID: "*A"},
			}

			Convey("Success", func() {
				mockProfDB.EXPECT().GetByID(tenantID, id).Return(existing, nil)
				mockMikDB.EXPECT().GetActiveMikrotik(tenantID).Return(activeMikrotik, nil)
				mockDB.EXPECT().DoInTransaction(gomock.Any()).DoAndReturn(func(f func(txDB outbound_port.DatabasePort) (interface{}, error)) (interface{}, error) {
					return f(mockDB)
				})
				mockFactory.EXPECT().NewClient(activeMikrotik).Return(mockClient, nil)
				mockClient.EXPECT().RunArgs("/ppp/profile/remove", gomock.Any()).Return(&routeros.Reply{}, nil)
				mockClient.EXPECT().Close().Return(nil)
				mockProfDB.EXPECT().Delete(tenantID, id).Return(nil)

				err := domain.DeleteProfile(ctx, id.String())
				So(err, ShouldBeNil)
			})

			Convey("Error - Mikrotik Object ID missing", func() {
				mockProfDB.EXPECT().GetByID(tenantID, id).Return(&model.ProfileWithPPPoE{}, nil)
				err := domain.DeleteProfile(ctx, id.String())
				So(err, ShouldNotBeNil)
			})
		})
	})
}

