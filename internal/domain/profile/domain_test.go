package profile

import (
	"context"
	"testing"

	"MikrOps/internal/model"
	outbound_port "MikrOps/internal/port/outbound"
	mock_outbound_port "MikrOps/tests/mocks/port"

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
	ctx := context.Background()

	Convey("Test Profile Domain", t, func() {
		Convey("CreateProfile", func() {
			input := model.CreateProfileRequest{Name: "Prof1", Price: 50000, Type: model.ProfileTypePPPoE}
			mikrotikID := uuid.New()
			activeMikrotik := &model.Mikrotik{ID: mikrotikID.String()}

			Convey("Success", func() {
				mockMikDB.EXPECT().GetActiveMikrotik(gomock.Any()).Return(activeMikrotik, nil)
				mockDB.EXPECT().DoInTransaction(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, f outbound_port.InTransaction) (interface{}, error) {
					return f(mockDB)
				})
				mockProfDB.EXPECT().CreateProfile(gomock.Any(), gomock.Any(), mikrotikID).Return(&model.MikrotikProfile{ID: uuid.New().String()}, nil)
				mockProfDB.EXPECT().CreateProfilePPPoE(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
				mockFactory.EXPECT().NewClient(activeMikrotik).Return(mockClient, nil)
				mockClient.EXPECT().RunArgs("/ppp/profile/add", gomock.Any()).Return(&routeros.Reply{Done: &proto.Sentence{Map: map[string]string{"ret": "*A"}}}, nil)
				mockClient.EXPECT().Close().Return(nil)
				mockProfDB.EXPECT().UpdateMikrotikObjectID(gomock.Any(), gomock.Any(), "*A").Return(nil)
				mockProfDB.EXPECT().GetByID(gomock.Any(), gomock.Any()).Return(&model.MikrotikProfile{}, nil)

				res, err := domain.CreateProfile(ctx, input)
				So(err, ShouldBeNil)
				So(res, ShouldNotBeNil)
			})

			Convey("No Active Mikrotik", func() {
				mockMikDB.EXPECT().GetActiveMikrotik(gomock.Any()).Return(nil, nil)
				res, err := domain.CreateProfile(ctx, input)
				So(err, ShouldNotBeNil)
				So(res, ShouldBeNil)
			})
		})

		Convey("GetProfile", func() {
			id := uuid.New()
			Convey("Success", func() {
				mockProfDB.EXPECT().GetByID(gomock.Any(), id).Return(&model.MikrotikProfile{}, nil)
				res, err := domain.GetProfile(ctx, id.String())
				So(err, ShouldBeNil)
				So(res, ShouldNotBeNil)
			})
		})

		Convey("UpdateProfile", func() {
			id := uuid.New()
			input := model.CreateProfileRequest{Name: "Updated", Price: 60000, Type: model.ProfileTypePPPoE}
			activeMikrotik := &model.Mikrotik{ID: uuid.New().String()}
			objectID := "*A"
			existing := &model.MikrotikProfile{
				MikrotikObjectID: &objectID,
			}

			Convey("Success", func() {
				mockProfDB.EXPECT().GetByID(gomock.Any(), id).Return(existing, nil)
				mockMikDB.EXPECT().GetActiveMikrotik(gomock.Any()).Return(activeMikrotik, nil)
				mockDB.EXPECT().DoInTransaction(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, f outbound_port.InTransaction) (interface{}, error) {
					return f(mockDB)
				})
				mockProfDB.EXPECT().Update(gomock.Any(), id, gomock.Any()).Return(nil)
				mockFactory.EXPECT().NewClient(activeMikrotik).Return(mockClient, nil)
				mockClient.EXPECT().RunArgs("/ppp/profile/set", gomock.Any()).Return(&routeros.Reply{}, nil)
				mockClient.EXPECT().Close().Return(nil)
				mockProfDB.EXPECT().GetByID(gomock.Any(), id).Return(&model.MikrotikProfile{}, nil)

				res, err := domain.UpdateProfile(ctx, id.String(), input)
				So(err, ShouldBeNil)
				So(res, ShouldNotBeNil)
			})
		})

		Convey("DeleteProfile", func() {
			id := uuid.New()
			activeMikrotik := &model.Mikrotik{ID: uuid.New().String()}
			objectID := "*A"
			existing := &model.MikrotikProfile{
				MikrotikObjectID: &objectID,
			}

			Convey("Success", func() {
				mockProfDB.EXPECT().GetByID(gomock.Any(), id).Return(existing, nil)
				mockMikDB.EXPECT().GetActiveMikrotik(gomock.Any()).Return(activeMikrotik, nil)
				mockDB.EXPECT().DoInTransaction(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, f outbound_port.InTransaction) (interface{}, error) {
					return f(mockDB)
				})
				mockFactory.EXPECT().NewClient(activeMikrotik).Return(mockClient, nil)
				mockClient.EXPECT().RunArgs("/ppp/profile/remove", gomock.Any()).Return(&routeros.Reply{}, nil)
				mockClient.EXPECT().Close().Return(nil)
				mockProfDB.EXPECT().Delete(gomock.Any(), id).Return(nil)

				err := domain.DeleteProfile(ctx, id.String())
				So(err, ShouldBeNil)
			})

			Convey("Error - Mikrotik Object ID missing", func() {
				mockProfDB.EXPECT().GetByID(gomock.Any(), id).Return(&model.MikrotikProfile{}, nil)
				mockMikDB.EXPECT().GetActiveMikrotik(gomock.Any()).Return(activeMikrotik, nil)
				mockDB.EXPECT().DoInTransaction(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, f outbound_port.InTransaction) (interface{}, error) {
					return f(mockDB)
				})
				mockProfDB.EXPECT().Delete(gomock.Any(), id).Return(nil)

				err := domain.DeleteProfile(ctx, id.String())
				So(err, ShouldBeNil)
			})
		})
	})
}
