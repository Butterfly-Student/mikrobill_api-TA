package tenant_test

import (
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	. "github.com/smartystreets/goconvey/convey"
	"go.uber.org/zap"

	"prabogo/internal/domain/tenant"
	"prabogo/internal/model"
	mock_outbound_port "prabogo/tests/mocks/port"
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

		tenantID := uuid.New()
		userID := uuid.New()
		now := time.Now()

		input := model.TenantInput{
			Name: "Test Tenant",
		}

		tenant := &model.Tenant{
			ID:           tenantID,
			Name:         "Test Tenant",
			IsActive:     true,
			Status:       "active",
			MaxMikrotiks: 10,
			CreatedAt:    now,
			UpdatedAt:    now,
		}

		Convey("CreateTenant", func() {
			Convey("Database error", func() {
				mockTenantDatabasePort.EXPECT().CreateTenant(gomock.Any()).Return(nil, errors.New("db error")).Times(1)

				result, err := tenantDomain.CreateTenant(input, userID)
				So(err, ShouldNotBeNil)
				So(result, ShouldBeNil)
			})

			Convey("Success", func() {
				mockTenantDatabasePort.EXPECT().CreateTenant(gomock.Any()).Return(tenant, nil).Times(1)

				result, err := tenantDomain.CreateTenant(input, userID)
				So(err, ShouldBeNil)
				So(result, ShouldNotBeNil)
				So(result.ID, ShouldEqual, tenantID)
			})
		})

		Convey("GetTenant", func() {
			Convey("Tenant not found", func() {
				mockTenantDatabasePort.EXPECT().GetByID(tenantID).Return(nil, errors.New("not found")).Times(1)

				result, err := tenantDomain.GetTenant(tenantID)
				So(err, ShouldNotBeNil)
				So(result, ShouldBeNil)
			})

			Convey("Success", func() {
				mockTenantDatabasePort.EXPECT().GetByID(tenantID).Return(tenant, nil).Times(1)

				result, err := tenantDomain.GetTenant(tenantID)
				So(err, ShouldBeNil)
				So(result, ShouldNotBeNil)
				So(result.ID, ShouldEqual, tenantID)
			})
		})

		Convey("ListTenants", func() {
			filter := model.TenantFilter{}

			Convey("Database error", func() {
				mockTenantDatabasePort.EXPECT().List(filter).Return(nil, errors.New("db error")).Times(1)

				results, err := tenantDomain.ListTenants(filter)
				So(err, ShouldNotBeNil)
				So(results, ShouldBeNil)
			})

			Convey("Success", func() {
				mockTenantDatabasePort.EXPECT().List(filter).Return([]model.Tenant{*tenant}, nil).Times(1)

				results, err := tenantDomain.ListTenants(filter)
				So(err, ShouldBeNil)
				So(results, ShouldHaveLength, 1)
				So(results[0].ID, ShouldEqual, tenantID)
			})
		})

		Convey("UpdateTenant", func() {
			Convey("Tenant not found", func() {
				mockTenantDatabasePort.EXPECT().GetByID(tenantID).Return(nil, errors.New("not found")).Times(1)

				result, err := tenantDomain.UpdateTenant(tenantID, input, userID)
				So(err, ShouldNotBeNil)
				So(result, ShouldBeNil)
			})

			Convey("Database error on update", func() {
				mockTenantDatabasePort.EXPECT().GetByID(tenantID).Return(tenant, nil).Times(1)
				mockTenantDatabasePort.EXPECT().Update(tenantID, input).Return(errors.New("db error")).Times(1)

				result, err := tenantDomain.UpdateTenant(tenantID, input, userID)
				So(err, ShouldNotBeNil)
				So(result, ShouldBeNil)
			})

			Convey("Success", func() {
				mockTenantDatabasePort.EXPECT().GetByID(tenantID).Return(tenant, nil).Times(2)
				mockTenantDatabasePort.EXPECT().Update(tenantID, input).Return(nil).Times(1)

				result, err := tenantDomain.UpdateTenant(tenantID, input, userID)
				So(err, ShouldBeNil)
				So(result, ShouldNotBeNil)
				So(result.ID, ShouldEqual, tenantID)
			})
		})

		Convey("DeleteTenant", func() {
			Convey("Tenant not found", func() {
				mockTenantDatabasePort.EXPECT().GetByID(tenantID).Return(nil, errors.New("not found")).Times(1)

				err := tenantDomain.DeleteTenant(tenantID, userID)
				So(err, ShouldNotBeNil)
			})

			Convey("Database error on delete", func() {
				mockTenantDatabasePort.EXPECT().GetByID(tenantID).Return(tenant, nil).Times(1)
				mockTenantDatabasePort.EXPECT().Delete(tenantID).Return(errors.New("db error")).Times(1)

				err := tenantDomain.DeleteTenant(tenantID, userID)
				So(err, ShouldNotBeNil)
			})

			Convey("Success", func() {
				mockTenantDatabasePort.EXPECT().GetByID(tenantID).Return(tenant, nil).Times(1)
				mockTenantDatabasePort.EXPECT().Delete(tenantID).Return(nil).Times(1)

				err := tenantDomain.DeleteTenant(tenantID, userID)
				So(err, ShouldBeNil)
			})
		})

		Convey("GetTenantStats", func() {
			stats := &model.TenantStats{
				TenantID:       tenantID,
				MikrotiksCount: 5,
			}

			Convey("Database error", func() {
				mockTenantDatabasePort.EXPECT().GetStats(tenantID).Return(nil, errors.New("db error")).Times(1)

				result, err := tenantDomain.GetTenantStats(tenantID)
				So(err, ShouldNotBeNil)
				So(result, ShouldBeNil)
			})

			Convey("Success", func() {
				mockTenantDatabasePort.EXPECT().GetStats(tenantID).Return(stats, nil).Times(1)

				result, err := tenantDomain.GetTenantStats(tenantID)
				So(err, ShouldBeNil)
				So(result, ShouldNotBeNil)
				So(result.MikrotiksCount, ShouldEqual, 5)
			})
		})

		Convey("CheckResourceLimit", func() {
			stats := &model.TenantStats{
				TenantID:       tenantID,
				MikrotiksCount: 5,
				MaxMikrotiks:   10,
			}

			Convey("Tenant not found", func() {
				mockTenantDatabasePort.EXPECT().GetByID(tenantID).Return(nil, errors.New("not found")).Times(1)

				err := tenantDomain.CheckResourceLimit(tenantID, "mikrotik")
				So(err, ShouldNotBeNil)
			})

			Convey("Stats error", func() {
				mockTenantDatabasePort.EXPECT().GetByID(tenantID).Return(tenant, nil).Times(1)
				mockTenantDatabasePort.EXPECT().GetStats(tenantID).Return(nil, errors.New("error")).Times(1)

				err := tenantDomain.CheckResourceLimit(tenantID, "mikrotik")
				So(err, ShouldNotBeNil)
			})

			Convey("Limit exceeded", func() {
				fullStats := &model.TenantStats{
					TenantID:       tenantID,
					MikrotiksCount: 10,
					MaxMikrotiks:   10,
				}
				mockTenantDatabasePort.EXPECT().GetByID(tenantID).Return(tenant, nil).Times(1)
				mockTenantDatabasePort.EXPECT().GetStats(tenantID).Return(fullStats, nil).Times(1)

				err := tenantDomain.CheckResourceLimit(tenantID, "mikrotik")
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "limit reached")
			})

			Convey("Success", func() {
				mockTenantDatabasePort.EXPECT().GetByID(tenantID).Return(tenant, nil).Times(1)
				mockTenantDatabasePort.EXPECT().GetStats(tenantID).Return(stats, nil).Times(1)

				err := tenantDomain.CheckResourceLimit(tenantID, "mikrotik")
				So(err, ShouldBeNil)
			})

			Convey("Unknown resource", func() {
				mockTenantDatabasePort.EXPECT().GetByID(tenantID).Return(tenant, nil).Times(1)
				mockTenantDatabasePort.EXPECT().GetStats(tenantID).Return(stats, nil).Times(1)

				err := tenantDomain.CheckResourceLimit(tenantID, "unknown")
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "unknown resource type")
			})
		})
	})
}
