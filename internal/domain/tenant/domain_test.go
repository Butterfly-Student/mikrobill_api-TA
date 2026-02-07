package tenant_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	. "github.com/smartystreets/goconvey/convey"
	"go.uber.org/zap"

	"MikrOps/internal/domain/tenant"
	"MikrOps/internal/model"
	mock_outbound_port "MikrOps/tests/mocks/port"
)

func TestTenant(t *testing.T) {
	Convey("Test Tenant Domain", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockDatabasePort := mock_outbound_port.NewMockDatabasePort(mockCtrl)
		mockTenantDatabasePort := mock_outbound_port.NewMockTenantDatabasePort(mockCtrl)

		mockDatabasePort.EXPECT().Tenant().Return(mockTenantDatabasePort).AnyTimes()

		logger := zap.NewNop()
		tenantDomain := tenant.NewTenantDomain(mockDatabasePort, logger)

		ctx := context.Background()
		tenantID := uuid.New()
		userID := uuid.New()
		now := time.Now()

		input := model.CreateTenantRequest{
			Name: "Test Tenant",
		}

		updateInput := model.UpdateTenantRequest{
			Name: strPtr("Test Tenant Updated"),
		}

		tenantObj := &model.Tenant{
			ID:           tenantID.String(),
			Name:         "Test Tenant",
			IsActive:     true,
			Status:       "active",
			MaxMikrotiks: 10,
			CreatedAt:    now,
			UpdatedAt:    now,
		}

		Convey("CreateTenant", func() {
			Convey("Database error", func() {
				mockTenantDatabasePort.EXPECT().CreateTenant(gomock.Any(), gomock.Any()).Return(nil, errors.New("db error")).Times(1)

				result, err := tenantDomain.CreateTenant(ctx, input, userID)
				So(err, ShouldNotBeNil)
				So(result, ShouldBeNil)
			})

			Convey("Success", func() {
				mockTenantDatabasePort.EXPECT().CreateTenant(gomock.Any(), gomock.Any()).Return(tenantObj, nil).Times(1)

				result, err := tenantDomain.CreateTenant(ctx, input, userID)
				So(err, ShouldBeNil)
				So(result, ShouldNotBeNil)
				So(result.ID, ShouldEqual, tenantID.String())
			})
		})

		Convey("GetTenant", func() {
			Convey("Tenant not found", func() {
				mockTenantDatabasePort.EXPECT().GetByID(gomock.Any(), tenantID).Return(nil, errors.New("not found")).Times(1)

				result, err := tenantDomain.GetTenant(ctx, tenantID)
				So(err, ShouldNotBeNil)
				So(result, ShouldBeNil)
			})

			Convey("Success", func() {
				mockTenantDatabasePort.EXPECT().GetByID(gomock.Any(), tenantID).Return(tenantObj, nil).Times(1)

				result, err := tenantDomain.GetTenant(ctx, tenantID)
				So(err, ShouldBeNil)
				So(result, ShouldNotBeNil)
				So(result.ID, ShouldEqual, tenantID.String())
			})
		})

		Convey("ListTenants", func() {
			filter := model.TenantFilter{}

			Convey("Database error", func() {
				mockTenantDatabasePort.EXPECT().List(gomock.Any(), filter).Return(nil, errors.New("db error")).Times(1)

				results, err := tenantDomain.ListTenants(ctx, filter)
				So(err, ShouldNotBeNil)
				So(results, ShouldBeNil)
			})

			Convey("Success", func() {
				mockTenantDatabasePort.EXPECT().List(gomock.Any(), filter).Return([]model.Tenant{*tenantObj}, nil).Times(1)

				results, err := tenantDomain.ListTenants(ctx, filter)
				So(err, ShouldBeNil)
				So(results, ShouldHaveLength, 1)
				So(results[0].ID, ShouldEqual, tenantID.String())
			})
		})

		Convey("UpdateTenant", func() {
			Convey("Tenant not found", func() {
				mockTenantDatabasePort.EXPECT().GetByID(gomock.Any(), tenantID).Return(nil, errors.New("not found")).Times(1)

				result, err := tenantDomain.UpdateTenant(ctx, tenantID, updateInput, userID)
				So(err, ShouldNotBeNil)
				So(result, ShouldBeNil)
			})

			Convey("Database error on update", func() {
				mockTenantDatabasePort.EXPECT().GetByID(gomock.Any(), tenantID).Return(tenantObj, nil).Times(1)
				mockTenantDatabasePort.EXPECT().Update(gomock.Any(), tenantID, updateInput).Return(nil, errors.New("db error")).Times(1)

				result, err := tenantDomain.UpdateTenant(ctx, tenantID, updateInput, userID)
				So(err, ShouldNotBeNil)
				So(result, ShouldBeNil)
			})

			Convey("Success", func() {
				mockTenantDatabasePort.EXPECT().GetByID(gomock.Any(), tenantID).Return(tenantObj, nil).Times(1)
				mockTenantDatabasePort.EXPECT().Update(gomock.Any(), tenantID, updateInput).Return(tenantObj, nil).Times(1)

				result, err := tenantDomain.UpdateTenant(ctx, tenantID, updateInput, userID)
				So(err, ShouldBeNil)
				So(result, ShouldNotBeNil)
				So(result.ID, ShouldEqual, tenantID.String())
			})
		})

		Convey("DeleteTenant", func() {
			Convey("Tenant not found", func() {
				mockTenantDatabasePort.EXPECT().GetByID(gomock.Any(), tenantID).Return(nil, errors.New("not found")).Times(1)

				err := tenantDomain.DeleteTenant(ctx, tenantID, userID)
				So(err, ShouldNotBeNil)
			})

			Convey("Database error on delete", func() {
				mockTenantDatabasePort.EXPECT().GetByID(gomock.Any(), tenantID).Return(tenantObj, nil).Times(1)
				mockTenantDatabasePort.EXPECT().Delete(gomock.Any(), tenantID).Return(errors.New("db error")).Times(1)

				err := tenantDomain.DeleteTenant(ctx, tenantID, userID)
				So(err, ShouldNotBeNil)
			})

			Convey("Success", func() {
				mockTenantDatabasePort.EXPECT().GetByID(gomock.Any(), tenantID).Return(tenantObj, nil).Times(1)
				mockTenantDatabasePort.EXPECT().Delete(gomock.Any(), tenantID).Return(nil).Times(1)

				err := tenantDomain.DeleteTenant(ctx, tenantID, userID)
				So(err, ShouldBeNil)
			})
		})

		Convey("GetTenantStats", func() {
			stats := &model.TenantStatsResponse{
				TenantID:       tenantID,
				MikrotiksCount: 5,
			}

			Convey("Database error", func() {
				mockTenantDatabasePort.EXPECT().GetStats(gomock.Any(), tenantID).Return(nil, errors.New("db error")).Times(1)

				result, err := tenantDomain.GetTenantStats(ctx, tenantID)
				So(err, ShouldNotBeNil)
				So(result, ShouldBeNil)
			})

			Convey("Success", func() {
				mockTenantDatabasePort.EXPECT().GetStats(gomock.Any(), tenantID).Return(stats, nil).Times(1)

				result, err := tenantDomain.GetTenantStats(ctx, tenantID)
				So(err, ShouldBeNil)
				So(result, ShouldNotBeNil)
				So(result.MikrotiksCount, ShouldEqual, 5)
			})
		})

		Convey("CheckResourceLimit", func() {
			stats := &model.TenantStatsResponse{
				TenantID:       tenantID,
				MikrotiksCount: 5,
				MaxMikrotiks:   10,
			}

			Convey("Tenant not found", func() {
				mockTenantDatabasePort.EXPECT().GetByID(gomock.Any(), tenantID).Return(nil, errors.New("not found")).Times(1)

				err := tenantDomain.CheckResourceLimit(ctx, tenantID, "mikrotik")
				So(err, ShouldNotBeNil)
			})

			Convey("Stats error", func() {
				mockTenantDatabasePort.EXPECT().GetByID(gomock.Any(), tenantID).Return(tenantObj, nil).Times(1)
				mockTenantDatabasePort.EXPECT().GetStats(gomock.Any(), tenantID).Return(nil, errors.New("error")).Times(1)

				err := tenantDomain.CheckResourceLimit(ctx, tenantID, "mikrotik")
				So(err, ShouldNotBeNil)
			})

			Convey("Limit exceeded", func() {
				fullStats := &model.TenantStatsResponse{
					TenantID:       tenantID,
					MikrotiksCount: 10,
					MaxMikrotiks:   10,
				}
				mockTenantDatabasePort.EXPECT().GetByID(gomock.Any(), tenantID).Return(tenantObj, nil).Times(1)
				mockTenantDatabasePort.EXPECT().GetStats(gomock.Any(), tenantID).Return(fullStats, nil).Times(1)

				err := tenantDomain.CheckResourceLimit(ctx, tenantID, "mikrotik")
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "limit reached")
			})

			Convey("Success", func() {
				mockTenantDatabasePort.EXPECT().GetByID(gomock.Any(), tenantID).Return(tenantObj, nil).Times(1)
				mockTenantDatabasePort.EXPECT().GetStats(gomock.Any(), tenantID).Return(stats, nil).Times(1)

				err := tenantDomain.CheckResourceLimit(ctx, tenantID, "mikrotik")
				So(err, ShouldBeNil)
			})

			Convey("Unknown resource", func() {
				mockTenantDatabasePort.EXPECT().GetByID(gomock.Any(), tenantID).Return(tenantObj, nil).Times(1)
				mockTenantDatabasePort.EXPECT().GetStats(gomock.Any(), tenantID).Return(stats, nil).Times(1)

				err := tenantDomain.CheckResourceLimit(ctx, tenantID, "unknown")
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "unknown resource type")
			})
		})
	})
}

func strPtr(s string) *string {
	return &s
}
