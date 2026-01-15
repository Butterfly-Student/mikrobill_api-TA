package mikrotik

import (
	"context"
	"errors"
	"testing"

	"prabogo/internal/model"
	mock_outbound_port "prabogo/tests/mocks/port"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	. "github.com/smartystreets/goconvey/convey"
)

func TestMikrotik(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mock_outbound_port.NewMockDatabasePort(ctrl)
	mockMikDB := mock_outbound_port.NewMockMikrotikDatabasePort(ctrl)

	mockDB.EXPECT().Mikrotik().Return(mockMikDB).AnyTimes()

	domain := NewMikrotikDomain(mockDB)
	tenantID := uuid.New()
	ctx := context.WithValue(context.Background(), "tenant_id", tenantID)

	Convey("Test Mikrotik Domain", t, func() {
		Convey("Create", func() {
			input := model.MikrotikInput{Name: "Mik1", Host: "1.1.1.1"}
			Convey("Success", func() {
				mockMikDB.EXPECT().Create(tenantID, input).Return(&model.Mikrotik{ID: uuid.New()}, nil)
				res, err := domain.Create(ctx, input)
				So(err, ShouldBeNil)
				So(res, ShouldNotBeNil)
			})
			Convey("Error", func() {
				mockMikDB.EXPECT().Create(tenantID, input).Return(nil, errors.New("error"))
				res, err := domain.Create(ctx, input)
				So(err, ShouldNotBeNil)
				So(res, ShouldBeNil)
			})
		})

		Convey("GetByID", func() {
			id := uuid.New()
			Convey("Success", func() {
				mockMikDB.EXPECT().GetByID(tenantID, id).Return(&model.Mikrotik{ID: id}, nil)
				res, err := domain.GetByID(ctx, id)
				So(err, ShouldBeNil)
				So(res.ID, ShouldEqual, id)
			})
		})

		Convey("List", func() {
			filter := model.MikrotikFilter{}
			Convey("Success", func() {
				mockMikDB.EXPECT().List(tenantID, filter).Return([]model.Mikrotik{{}}, nil)
				res, err := domain.List(ctx, filter)
				So(err, ShouldBeNil)
				So(len(res), ShouldEqual, 1)
			})
		})

		Convey("SetActive", func() {
			id := uuid.New()
			Convey("Success", func() {
				mockMikDB.EXPECT().SetActive(tenantID, id).Return(nil)
				err := domain.SetActive(ctx, id)
				So(err, ShouldBeNil)
			})
		})

		Convey("GetActiveMikrotik", func() {
			Convey("Success", func() {
				mockMikDB.EXPECT().GetActiveMikrotik(tenantID).Return(&model.Mikrotik{}, nil)
				res, err := domain.GetActiveMikrotik(ctx)
				So(err, ShouldBeNil)
				So(res, ShouldNotBeNil)
			})
		})
	})
}
