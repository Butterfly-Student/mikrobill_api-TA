package client_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"

	"MikrOps/internal/domain/client"
	"MikrOps/internal/model"
	mock_outbound_port "MikrOps/tests/mocks/port"
)

func TestClient(t *testing.T) {
	Convey("Test Client", t, func() {
		mockCtrl := gomock.NewController(t)

		defer mockCtrl.Finish()

		mockDatabasePort := mock_outbound_port.NewMockDatabasePort(mockCtrl)
		mockCachePort := mock_outbound_port.NewMockCachePort(mockCtrl)
		mockMessagePort := mock_outbound_port.NewMockMessagePort(mockCtrl)

		mockClientDatabasePort := mock_outbound_port.NewMockClientDatabasePort(mockCtrl)
		mockClientMessagePort := mock_outbound_port.NewMockClientMessagePort(mockCtrl)
		mockClientCachePort := mock_outbound_port.NewMockClientCachePort(mockCtrl)

		mockDatabasePort.EXPECT().Client().Return(mockClientDatabasePort).AnyTimes()
		mockMessagePort.EXPECT().Client().Return(mockClientMessagePort).AnyTimes()
		mockCachePort.EXPECT().Client().Return(mockClientCachePort).AnyTimes()

		clientDomain := client.NewClientDomain(mockDatabasePort, mockMessagePort, mockCachePort)

		inputs := []model.ClientInput{
			{
				Name: "Test Client",
			},
		}

		outputs := []model.Client{
			{
				ID: 1,
				ClientInput: model.ClientInput{
					Name:      "Test Client",
					BearerKey: "test-bearer-key",
					UpdatedAt: time.Now(),
					CreatedAt: time.Now(),
				},
			},
		}

		filter := model.ClientFilter{
			BearerKeys: []string{"test-bearer-key"},
			IDs:        []int{1},
			Names:      []string{"Test Client"},
		}

		Convey("Upsert", func() {
			Convey("Input is empty", func() {
				_, err := clientDomain.Upsert(context.Background(), []model.ClientInput{})
				So(err, ShouldNotBeNil)
			})

			Convey("Database client upsert error", func() {
				mockClientDatabasePort.EXPECT().Upsert(gomock.Any(), gomock.Any()).Return(errors.New("error")).Times(1)

				_, err := clientDomain.Upsert(context.Background(), inputs)
				So(err, ShouldNotBeNil)
			})

			Convey("Database client find by filter error", func() {
				mockClientDatabasePort.EXPECT().Upsert(gomock.Any(), gomock.Any()).Return(nil).Times(1)
				mockClientDatabasePort.EXPECT().FindByFilter(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("error")).Times(1)

				_, err := clientDomain.Upsert(context.Background(), inputs)
				So(err, ShouldNotBeNil)
			})

			Convey("Success", func() {
				mockClientDatabasePort.EXPECT().Upsert(gomock.Any(), gomock.Any()).Return(nil).Times(1)
				mockClientDatabasePort.EXPECT().FindByFilter(gomock.Any(), gomock.Any(), gomock.Any()).Return(outputs, nil).Times(1)

				results, err := clientDomain.Upsert(context.Background(), inputs)
				So(err, ShouldBeNil)
				So(results, ShouldNotBeEmpty)
				So(results[0].Name, ShouldEqual, "Test Client")
			})
		})

		Convey("FindByFilter", func() {
			Convey("Filter is empty", func() {
				_, err := clientDomain.FindByFilter(context.Background(), model.ClientFilter{})
				So(err, ShouldNotBeNil)
			})

			Convey("Database client find by filter error", func() {
				mockClientDatabasePort.EXPECT().FindByFilter(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("error")).Times(1)

				_, err := clientDomain.FindByFilter(context.Background(), filter)
				So(err, ShouldNotBeNil)
			})

			Convey("Success", func() {
				mockClientDatabasePort.EXPECT().FindByFilter(gomock.Any(), gomock.Any(), gomock.Any()).Return(outputs, nil).Times(1)

				results, err := clientDomain.FindByFilter(context.Background(), filter)
				So(err, ShouldBeNil)
				So(results, ShouldNotBeEmpty)
				So(results[0].Name, ShouldEqual, "Test Client")
			})
		})

		Convey("DeleteByFilter", func() {
			Convey("Filter is empty", func() {
				err := clientDomain.DeleteByFilter(context.Background(), model.ClientFilter{})
				So(err, ShouldNotBeNil)
			})

			Convey("Database client delete by filter error", func() {
				mockClientDatabasePort.EXPECT().DeleteByFilter(gomock.Any(), gomock.Any()).Return(errors.New("error")).Times(1)

				err := clientDomain.DeleteByFilter(context.Background(), filter)
				So(err, ShouldNotBeNil)
			})

			Convey("Success", func() {
				mockClientDatabasePort.EXPECT().DeleteByFilter(gomock.Any(), gomock.Any()).Return(nil).Times(1)

				err := clientDomain.DeleteByFilter(context.Background(), filter)
				So(err, ShouldBeNil)
			})
		})

		Convey("PublishUpsert", func() {
			Convey("Input is empty", func() {
				err := clientDomain.PublishUpsert(context.Background(), []model.ClientInput{})
				So(err, ShouldNotBeNil)
			})

			Convey("Message client publish upsert error", func() {
				mockClientMessagePort.EXPECT().PublishUpsert(gomock.Any(), gomock.Any()).Return(errors.New("error")).Times(1)

				err := clientDomain.PublishUpsert(context.Background(), inputs)
				So(err, ShouldNotBeNil)
			})

			Convey("Success", func() {
				mockClientMessagePort.EXPECT().PublishUpsert(gomock.Any(), gomock.Any()).Return(nil).Times(1)

				err := clientDomain.PublishUpsert(context.Background(), inputs)
				So(err, ShouldBeNil)
			})
		})

		Convey("IsExists", func() {
			Convey("Bearer key is empty", func() {
				_, err := clientDomain.IsExists(context.Background(), "")
				So(err, ShouldNotBeNil)
			})

			Convey("Cache miss, database error", func() {
				mockClientCachePort.EXPECT().GetClient(gomock.Any(), gomock.Any()).Return(model.Client{}, false).Times(1)
				mockClientDatabasePort.EXPECT().IsExists(gomock.Any(), gomock.Any()).Return(false, errors.New("error")).Times(1)

				_, err := clientDomain.IsExists(context.Background(), "test-bearer-key")
				So(err, ShouldNotBeNil)
			})

			Convey("Cache miss, database find by filter error", func() {
				mockClientCachePort.EXPECT().GetClient(gomock.Any(), gomock.Any()).Return(model.Client{}, false).Times(1)
				mockClientDatabasePort.EXPECT().IsExists(gomock.Any(), gomock.Any()).Return(true, nil).Times(1)
				mockClientDatabasePort.EXPECT().FindByFilter(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("error")).Times(1)

				_, err := clientDomain.IsExists(context.Background(), "test-bearer-key")
				So(err, ShouldNotBeNil)
			})

			Convey("Cache miss, set client error", func() {
				mockClientCachePort.EXPECT().GetClient(gomock.Any(), gomock.Any()).Return(model.Client{}, false).Times(1)
				mockClientDatabasePort.EXPECT().IsExists(gomock.Any(), gomock.Any()).Return(true, nil).Times(1)
				mockClientDatabasePort.EXPECT().FindByFilter(gomock.Any(), gomock.Any(), gomock.Any()).Return(outputs, nil).Times(1)
				mockClientCachePort.EXPECT().SetClient(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("error")).Times(1)

				_, err := clientDomain.IsExists(context.Background(), "test-bearer-key")
				So(err, ShouldNotBeNil)
			})

			Convey("Cache miss, DB found, cache set success", func() {
				mockClientCachePort.EXPECT().GetClient(gomock.Any(), gomock.Any()).Return(model.Client{}, false).Times(1)
				mockClientDatabasePort.EXPECT().IsExists(gomock.Any(), gomock.Any()).Return(true, nil).Times(1)
				mockClientDatabasePort.EXPECT().FindByFilter(gomock.Any(), gomock.Any(), gomock.Any()).Return(outputs, nil).Times(1)
				mockClientCachePort.EXPECT().SetClient(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)

				result, err := clientDomain.IsExists(context.Background(), "test-bearer-key")
				So(err, ShouldBeNil)
				So(result, ShouldBeFalse)
			})

			Convey("Cache hit", func() {
				mockClientCachePort.EXPECT().GetClient(gomock.Any(), gomock.Any()).Return(outputs[0], true).Times(1)

				result, err := clientDomain.IsExists(context.Background(), "test-bearer-key")
				So(err, ShouldBeNil)
				So(result, ShouldBeTrue)
			})
		})
	})
}
