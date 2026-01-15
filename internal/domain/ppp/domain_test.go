package ppp

import (
	"context"
	"testing"

	"prabogo/internal/model"
	mock_outbound_port "prabogo/tests/mocks/port"

	"github.com/go-routeros/routeros/v3"
	"github.com/go-routeros/routeros/v3/proto"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	. "github.com/smartystreets/goconvey/convey"
)

func TestPPP(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mock_outbound_port.NewMockDatabasePort(ctrl)
	mockMikDB := mock_outbound_port.NewMockMikrotikDatabasePort(ctrl)
	mockFactory := mock_outbound_port.NewMockMikrotikClientFactory(ctrl)
	mockClient := mock_outbound_port.NewMockMikrotikClientPort(ctrl)

	mockDB.EXPECT().Mikrotik().Return(mockMikDB).AnyTimes()

	domain := NewPPPDomain(mockDB, mockFactory)
	tenantID := uuid.New()
	ctx := context.WithValue(context.Background(), "tenant_id", tenantID)

	Convey("Test PPP Domain", t, func() {
		Convey("Secrets", func() {
			Convey("MikrotikCreateSecret", func() {
				input := model.PPPSecretInput{Name: "user1", Password: "pass", Service: "pppoe"}
				mockMikDB.EXPECT().GetActiveMikrotik(tenantID).Return(&model.Mikrotik{}, nil)
				mockFactory.EXPECT().NewClient(gomock.Any()).Return(mockClient, nil)
				mockClient.EXPECT().RunArgs("/ppp/secret/add", gomock.Any()).Return(&routeros.Reply{}, nil)
				mockClient.EXPECT().Close().Return(nil)

				res, err := domain.MikrotikCreateSecret(ctx, input)
				So(err, ShouldBeNil)
				So(res.Name, ShouldEqual, "user1")
			})

			Convey("MikrotikGetSecret", func() {
				mockMikDB.EXPECT().GetActiveMikrotik(tenantID).Return(&model.Mikrotik{}, nil)
				mockFactory.EXPECT().NewClient(gomock.Any()).Return(mockClient, nil)
				mockClient.EXPECT().RunArgs("/ppp/secret/print", gomock.Any()).Return(&routeros.Reply{
					Re: []*proto.Sentence{{Map: map[string]string{"name": "user1", "profile": "default"}}},
				}, nil)
				mockClient.EXPECT().Close().Return(nil)

				res, err := domain.MikrotikGetSecret(ctx, "*1")
				So(err, ShouldBeNil)
				So(res.Name, ShouldEqual, "user1")
			})
		})

		Convey("Profiles", func() {
			Convey("MikrotikCreateProfile", func() {
				input := model.PPPProfileInput{Name: "prof1"}
				mockMikDB.EXPECT().GetActiveMikrotik(tenantID).Return(&model.Mikrotik{}, nil)
				mockFactory.EXPECT().NewClient(gomock.Any()).Return(mockClient, nil)
				mockClient.EXPECT().RunArgs("/ppp/profile/add", gomock.Any()).Return(&routeros.Reply{}, nil)
				mockClient.EXPECT().Close().Return(nil)

				res, err := domain.MikrotikCreateProfile(ctx, input)
				So(err, ShouldBeNil)
				So(res.Name, ShouldEqual, "prof1")
			})

			Convey("MikrotikListProfiles", func() {
				mockMikDB.EXPECT().GetActiveMikrotik(tenantID).Return(&model.Mikrotik{}, nil)
				mockFactory.EXPECT().NewClient(gomock.Any()).Return(mockClient, nil)
				mockClient.EXPECT().Run("/ppp/profile/print").Return(&routeros.Reply{
					Re: []*proto.Sentence{{Map: map[string]string{"name": "prof1", "rate-limit": "1M/1M"}}},
				}, nil)
				mockClient.EXPECT().Close().Return(nil)

				res, err := domain.MikrotikListProfiles(ctx)
				So(err, ShouldBeNil)
				So(len(res), ShouldEqual, 1)
				So(res[0].Name, ShouldEqual, "prof1")
			})
		})
	})
}
