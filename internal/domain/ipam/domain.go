package ipam

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/palantir/stacktrace"
	"go.uber.org/zap"

	"MikrOps/internal/model"
	contextutil "MikrOps/utils/context"
	"MikrOps/utils/logger"
)

// IPAMDatabasePort defines the database operations for IPAM
type IPAMDatabasePort interface {
	CreatePool(ctx context.Context, pool *model.IPAMPool) error
	GetPool(ctx context.Context, poolID uuid.UUID) (model.IPAMPool, error)
	ListPools(ctx context.Context, tenantID string) ([]model.IPAMPool, error)
	UpdatePool(ctx context.Context, pool *model.IPAMPool) error
	DeletePool(ctx context.Context, poolID uuid.UUID) error

	CreateAllocation(ctx context.Context, allocation *model.IPAllocation) error
	GetAllocation(ctx context.Context, allocationID uuid.UUID) (model.IPAllocation, error)
	GetAllocationByIP(ctx context.Context, ipAddress string) (*model.IPAllocation, error)
	ListAllocations(ctx context.Context, filter model.IPAllocationFilter) ([]model.IPAllocation, error)
	UpdateAllocation(ctx context.Context, allocation *model.IPAllocation) error
	CleanupExpiredAllocations(ctx context.Context) error
}

// MikroTikIPAMPort defines the MikroTik operations for IPAM
type MikroTikIPAMPort interface {
	// Reserved for future MikroTik IPAM integration
}

type IPAMDomain interface {
	CreatePool(ctx context.Context, pool model.IPAMPool) (*model.IPAMPool, error)
	GetPool(ctx context.Context, poolID uuid.UUID) (*model.IPAMPool, error)
	ListPools(ctx context.Context, tenantID string) ([]model.IPAMPool, error)
	UpdatePool(ctx context.Context, poolID uuid.UUID, pool model.IPAMPoolUpdateRequest) (*model.IPAMPool, error)
	DeletePool(ctx context.Context, poolID uuid.UUID) error

	AllocateIP(ctx context.Context, req model.IPAllocationRequest) (*model.IPAllocation, error)
	ReleaseIP(ctx context.Context, allocationID uuid.UUID) error
	GetAllocation(ctx context.Context, allocationID uuid.UUID) (*model.IPAllocation, error)
	ListAllocations(ctx context.Context, filters model.IPAllocationFilter) ([]model.IPAllocation, error)
	ReserveIP(ctx context.Context, poolID uuid.UUID, ipAddress string) (*model.IPAllocation, error)
	DeallocateIP(ctx context.Context, allocationID uuid.UUID) error

	GetStats(ctx context.Context, tenantID string) (*model.IPAMStats, error)
	Reconcile(ctx context.Context, poolID uuid.UUID) error
	CleanupExpiredAllocations(ctx context.Context) error
	ReallocateIP(ctx context.Context, allocationID uuid.UUID, newCustomerID string) (*model.IPAllocation, error)
}

type ipamDomain struct {
	db       IPAMDatabasePort
	mikrotik MikroTikIPAMPort
	logger   *zap.Logger
}

func NewIPAMDomain(db IPAMDatabasePort, mikrotik MikroTikIPAMPort) IPAMDomain {
	return &ipamDomain{
		db:       db,
		mikrotik: mikrotik,
		logger:   logger.GetLogger().With(zap.String("service", "ipam")),
	}
}

func (d *ipamDomain) CreatePool(ctx context.Context, pool model.IPAMPool) (*model.IPAMPool, error) {
	tenantID, err := contextutil.GetTenantID(ctx)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get tenant id from context")
	}

	pool.ID = uuid.New()
	pool.TenantID = tenantID.String()
	pool.CreatedAt = time.Now()
	pool.UpdatedAt = time.Now()

	logger.Info("Creating IPAM pool",
		zap.String("tenant_id", tenantID.String()),
		zap.String("pool_name", pool.Name),
		zap.String("pool_type", pool.PoolType),
		zap.String("network", pool.Network))

	if err := d.db.CreatePool(ctx, &pool); err != nil {
		return nil, stacktrace.Propagate(err, "failed to create ipam pool")
	}

	logger.Info("IPAM pool created successfully",
		zap.String("pool_id", pool.ID.String()))

	return &pool, nil
}

func (d *ipamDomain) GetPool(ctx context.Context, poolID uuid.UUID) (*model.IPAMPool, error) {
	tenantID, err := contextutil.GetTenantID(ctx)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get tenant id from context")
	}

	pool, err := d.db.GetPool(ctx, poolID)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get ipam pool")
	}

	if pool.TenantID != tenantID.String() {
		return nil, stacktrace.NewError("pool not found or access denied")
	}

	return &pool, nil
}

func (d *ipamDomain) ListPools(ctx context.Context, tenantID string) ([]model.IPAMPool, error) {
	pools, err := d.db.ListPools(ctx, tenantID)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to list ipam pools")
	}
	return pools, nil
}

func (d *ipamDomain) UpdatePool(ctx context.Context, poolID uuid.UUID, pool model.IPAMPoolUpdateRequest) (*model.IPAMPool, error) {
	tenantID, err := contextutil.GetTenantID(ctx)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get tenant id from context")
	}

	existingPool, err := d.db.GetPool(ctx, poolID)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get existing pool")
	}

	if existingPool.TenantID != tenantID.String() {
		return nil, stacktrace.NewError("pool not found or access denied")
	}

	if pool.Name != "" {
		existingPool.Name = pool.Name
	}
	if pool.Description != "" {
		existingPool.Description = pool.Description
	}
	if pool.Network != "" {
		existingPool.Network = pool.Network
	}
	if pool.Netmask != "" {
		existingPool.Netmask = pool.Netmask
	}
	if pool.Gateway != "" {
		existingPool.Gateway = pool.Gateway
	}
	existingPool.UpdatedAt = time.Now()

	if err := d.db.UpdatePool(ctx, &existingPool); err != nil {
		return nil, stacktrace.Propagate(err, "failed to update ipam pool")
	}

	logger.Info("IPAM pool updated",
		zap.String("pool_id", poolID.String()))

	return &existingPool, nil
}

func (d *ipamDomain) DeletePool(ctx context.Context, poolID uuid.UUID) error {
	_, err := contextutil.GetTenantID(ctx)
	if err != nil {
		return stacktrace.Propagate(err, "failed to get tenant id from context")
	}

	if err := d.db.DeletePool(ctx, poolID); err != nil {
		return stacktrace.Propagate(err, "failed to delete ipam pool")
	}

	logger.Info("IPAM pool deleted",
		zap.String("pool_id", poolID.String()))

	return nil
}

func (d *ipamDomain) AllocateIP(ctx context.Context, req model.IPAllocationRequest) (*model.IPAllocation, error) {
	tenantID, err := contextutil.GetTenantID(ctx)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get tenant id from context")
	}

	pool, err := d.db.GetPool(ctx, req.PoolID)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get ipam pool")
	}

	if pool.TenantID != tenantID.String() {
		return nil, stacktrace.NewError("pool not found or access denied")
	}

	availableIP, err := d.findAvailableIP(ctx, &pool, req)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to find available ip")
	}

	now := time.Now()
	leaseStart := now
	allocation := &model.IPAllocation{
		ID:             uuid.New(),
		TenantID:       tenantID.String(),
		CustomerID:     req.CustomerID,
		PoolID:         req.PoolID,
		IPAddress:      availableIP.IPAddress,
		Status:         model.IPAllocationStatusActive,
		AllocatedAt:    now,
		LeaseStartTime: &leaseStart,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if req.MacAddress != nil {
		allocation.MacAddress = req.MacAddress
	}
	if req.Hostname != nil {
		allocation.Hostname = req.Hostname
	}

	if req.LeaseDuration > 0 {
		leaseEnd := time.Now().AddDate(0, 0, req.LeaseDuration)
		allocation.LeaseEndTime = &leaseEnd
	}

	if err := d.db.CreateAllocation(ctx, allocation); err != nil {
		return nil, stacktrace.Propagate(err, "failed to create ip allocation")
	}

	pool.UsedIPs++
	pool.UpdatedAt = time.Now()
	if err := d.db.UpdatePool(ctx, &pool); err != nil {
		logger.Error("failed to update pool used ip count",
			zap.String("pool_id", pool.ID.String()),
			zap.Error(err))
	} else {
		logger.Info("Pool updated",
			zap.String("pool_id", pool.ID.String()),
			zap.Int("used_ips", pool.UsedIPs))
	}

	logger.Info("IP allocated successfully",
		zap.String("allocation_id", allocation.ID.String()),
		zap.String("ip_address", availableIP.IPAddress),
		zap.String("customer_id", req.CustomerID))

	return allocation, nil
}

func (d *ipamDomain) ReleaseIP(ctx context.Context, allocationID uuid.UUID) error {
	tenantID, err := contextutil.GetTenantID(ctx)
	if err != nil {
		return stacktrace.Propagate(err, "failed to get tenant id from context")
	}

	allocation, err := d.db.GetAllocation(ctx, allocationID)
	if err != nil {
		return stacktrace.Propagate(err, "failed to get allocation")
	}

	if allocation.TenantID != tenantID.String() {
		return stacktrace.NewError("allocation not found or access denied")
	}

	if allocation.Status == model.IPAllocationStatusInactive || allocation.ReleasedAt != nil {
		return stacktrace.NewError("allocation already released")
	}

	now := time.Now()
	allocation.Status = model.IPAllocationStatusInactive
	allocation.ReleasedAt = &now
	allocation.UpdatedAt = now

	pool, err := d.db.GetPool(ctx, allocation.PoolID)
	if err != nil {
		return stacktrace.Propagate(err, "failed to get pool for update")
	}

	pool.UsedIPs--
	pool.UpdatedAt = now
	if err := d.db.UpdatePool(ctx, &pool); err != nil {
		logger.Error("failed to update pool used ip count",
			zap.String("pool_id", pool.ID.String()),
			zap.Error(err))
	} else {
		logger.Info("Pool updated",
			zap.String("pool_id", pool.ID.String()),
			zap.Int("used_ips", pool.UsedIPs))
	}

	logger.Info("IP released successfully",
		zap.String("allocation_id", allocation.ID.String()),
		zap.String("ip_address", allocation.IPAddress),
		zap.String("customer_id", allocation.CustomerID))

	if err := d.db.UpdateAllocation(ctx, &allocation); err != nil {
		return stacktrace.Propagate(err, "failed to update allocation")
	}

	return nil
}

func (d *ipamDomain) findAvailableIP(ctx context.Context, pool *model.IPAMPool, req model.IPAllocationRequest) (*model.IPAllocation, error) {
	if req.IPAddress != "" {
		existing, err := d.db.GetAllocationByIP(ctx, req.IPAddress)
		if err != nil {
			return nil, stacktrace.Propagate(err, "failed to check ip allocation")
		}
		if existing == nil {
			return &model.IPAllocation{
				IPAddress:  req.IPAddress,
				Status:     model.IPAllocationStatusActive,
				PoolID:     pool.ID,
				CustomerID: req.CustomerID,
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
			}, nil
		}
	}

	if pool.UsedIPs >= pool.TotalIPs {
		return nil, stacktrace.NewError("no available IPs in pool")
	}

	allocations, err := d.db.ListAllocations(ctx, model.IPAllocationFilter{
		PoolID: &pool.ID,
		Status: model.IPAllocationStatusInactive,
	})

	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to list allocations")
	}

	var availableIP model.IPAllocation
	for _, alloc := range allocations {
		availableIP = alloc
		break
	}

	if availableIP.IPAddress == "" {
		return nil, stacktrace.NewError("no available IPs in pool")
	}

	logger.Info("Found available IP from pool",
		zap.String("ip_address", availableIP.IPAddress))

	return &availableIP, nil
}

func (d *ipamDomain) GetAllocation(ctx context.Context, allocationID uuid.UUID) (*model.IPAllocation, error) {
	allocation, err := d.db.GetAllocation(ctx, allocationID)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get allocation")
	}
	return &allocation, nil
}

func (d *ipamDomain) ListAllocations(ctx context.Context, filters model.IPAllocationFilter) ([]model.IPAllocation, error) {
	allocations, err := d.db.ListAllocations(ctx, filters)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to list allocations")
	}
	return allocations, nil
}

func (d *ipamDomain) ReserveIP(ctx context.Context, poolID uuid.UUID, ipAddress string) (*model.IPAllocation, error) {
	tenantID, err := contextutil.GetTenantID(ctx)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get tenant id from context")
	}

	allocation, err := d.db.GetAllocationByIP(ctx, ipAddress)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to check ip allocation")
	}

	if allocation != nil && allocation.Status == model.IPAllocationStatusActive {
		return allocation, nil
	}

	pool, err := d.db.GetPool(ctx, poolID)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get pool")
	}

	if pool.UsedIPs >= pool.TotalIPs {
		return nil, stacktrace.NewError("no available IPs in pool")
	}

	newAllocation := &model.IPAllocation{
		ID:          uuid.New(),
		TenantID:    tenantID.String(),
		CustomerID:  "",
		PoolID:      poolID,
		IPAddress:   ipAddress,
		Status:      model.IPAllocationStatusReserved,
		AllocatedAt: time.Now(),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := d.db.CreateAllocation(ctx, newAllocation); err != nil {
		return nil, stacktrace.Propagate(err, "failed to reserve ip allocation")
	}

	logger.Info("IP reserved successfully",
		zap.String("allocation_id", newAllocation.ID.String()),
		zap.String("ip_address", ipAddress))

	return newAllocation, nil
}

func (d *ipamDomain) DeallocateIP(ctx context.Context, allocationID uuid.UUID) error {
	return d.ReleaseIP(ctx, allocationID)
}

func (d *ipamDomain) GetStats(ctx context.Context, tenantID string) (*model.IPAMStats, error) {
	pools, err := d.db.ListPools(ctx, tenantID)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to list pools")
	}

	stats := &model.IPAMStats{
		TotalPools: len(pools),
	}

	var totalIPs, usedIPs, reservedIPs int
	for _, pool := range pools {
		totalIPs += pool.TotalIPs
		usedIPs += pool.UsedIPs

		allocations, err := d.db.ListAllocations(ctx, model.IPAllocationFilter{
			PoolID: &pool.ID,
			Status: model.IPAllocationStatusReserved,
		})
		if err != nil {
			logger.Error("failed to count reserved ips", zap.Error(err))
		} else {
			reservedIPs = len(allocations)
		}
	}

	stats.TotalIPs = totalIPs
	stats.UsedIPs = usedIPs
	stats.AvailableIPs = totalIPs - usedIPs - reservedIPs

	if stats.TotalIPs > 0 {
		stats.AllocationRate = float64(usedIPs) / float64(len(pools)) / 30.0
		stats.AverageLeaseDays = 30.0
	}

	return stats, nil
}

func (d *ipamDomain) Reconcile(ctx context.Context, poolID uuid.UUID) error {
	logger.Info("Starting IPAM reconciliation",
		zap.String("pool_id", poolID.String()))

	pool, err := d.db.GetPool(ctx, poolID)
	if err != nil {
		return stacktrace.Propagate(err, "failed to get pool for reconciliation")
	}

	allocations, err := d.db.ListAllocations(ctx, model.IPAllocationFilter{
		PoolID: &pool.ID,
		Status: model.IPAllocationStatusActive,
	})
	if err != nil {
		return stacktrace.Propagate(err, "failed to list allocations for reconciliation")
	}

	pool.UsedIPs = len(allocations)
	pool.UpdatedAt = time.Now()

	if err := d.db.UpdatePool(ctx, &pool); err != nil {
		logger.Error("failed to update pool during reconciliation",
			zap.String("pool_id", pool.ID.String()),
			zap.Error(err))
	}

	logger.Info("IPAM reconciliation completed",
		zap.String("pool_id", poolID.String()),
		zap.Int("used_ips", pool.UsedIPs))

	return nil
}

func (d *ipamDomain) CleanupExpiredAllocations(ctx context.Context) error {
	logger.Info("Starting IPAM cleanup of expired allocations")

	expiredCount := 0
	err := d.db.CleanupExpiredAllocations(ctx)
	if err != nil {
		return stacktrace.Propagate(err, "failed to cleanup expired allocations")
	}

	logger.Info("IPAM cleanup completed",
		zap.Int("expired_allocations_cleaned", expiredCount))

	return nil
}

func (d *ipamDomain) ReallocateIP(ctx context.Context, allocationID uuid.UUID, newCustomerID string) (*model.IPAllocation, error) {
	_, err := contextutil.GetTenantID(ctx)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get tenant id from context")
	}

	allocation, err := d.db.GetAllocation(ctx, allocationID)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get allocation for reallocation")
	}

	if allocation.Status != model.IPAllocationStatusActive {
		return nil, stacktrace.NewError("allocation not active, cannot reallocate")
	}

	if allocation.CustomerID == newCustomerID {
		logger.Info("Allocation already assigned to customer",
			zap.String("allocation_id", allocationID.String()),
			zap.String("customer_id", newCustomerID))
		return &allocation, nil
	}

	oldCustomerID := allocation.CustomerID
	allocation.CustomerID = newCustomerID
	allocation.UpdatedAt = time.Now()

	if err := d.db.UpdateAllocation(ctx, &allocation); err != nil {
		return nil, stacktrace.Propagate(err, "failed to update allocation for reallocation")
	}

	logger.Info("IP reallocated successfully",
		zap.String("allocation_id", allocationID.String()),
		zap.String("old_customer_id", oldCustomerID),
		zap.String("new_customer_id", newCustomerID))

	return &allocation, nil
}
