// internal/service/mikrotik_service.go
package usecase

import (
	"context"
	"errors"
	"fmt"
	"mikrobill/internal/entity"
	"mikrobill/internal/infrastructure/mikrotik"
	"mikrobill/internal/model"
	"mikrobill/internal/port/repository"
	pkg_logger "mikrobill/pkg/logger"
	"mikrobill/pkg/utils"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type MikrotikUseCase interface {
	// CRUD Operations
	Create(ctx context.Context, req model.CreateMikrotikRequest) (*entity.Mikrotik, error)
	GetByID(ctx context.Context, id string) (*model.MikrotikResponse, error)
	List(ctx context.Context, req model.PaginationRequest) (*model.PaginationResponse, error)
	Update(ctx context.Context, id string, req model.UpdateMikrotikRequest) error
	Delete(ctx context.Context, id string) error

	// Connection & Status
	TestConnectionByID(ctx context.Context, id string) error
	UpdateStatus(ctx context.Context, id string, status string) error

	// Active Management
	SetActiveMikrotik(ctx context.Context, id string) error
	GetActiveMikrotik(ctx context.Context) (*model.MikrotikStatusResponse, error)

	// Client Management
	GetMikrotikClient() (*mikrotik.Client, error)
	GetClientByID(ctx context.Context, id string) (*mikrotik.Client, error)
	TestMikrotikConnection(ctx context.Context) error
}

type mikrotikUseCase struct {
	mikrotikRepo  repository.MikrotikRepository
	encryptionKey string
}

func NewMikrotikUseCase(mikrotikRepo repository.MikrotikRepository, encryptionKey string) MikrotikUseCase {
	return &mikrotikUseCase{
		mikrotikRepo:  mikrotikRepo,
		encryptionKey: encryptionKey,
	}
}

// Create creates new Mikrotik configuration
func (s *mikrotikUseCase) Create(ctx context.Context, req model.CreateMikrotikRequest) (*entity.Mikrotik, error) {
	// Standard assignment (Plain Text)
	encryptedPassword := req.APIPassword

	mikrotik := &entity.Mikrotik{
		Name:                 req.Name,
		Host:                 req.Host,
		Port:                 req.Port,
		APIUsername:          req.APIUsername,
		APIEncryptedPassword: encryptedPassword,
		Keepalive:            req.Keepalive,
		Timeout:              req.Timeout,
		Location:             req.Location,
		Description:          req.Description,
		IsActive:             req.IsActive,
		Status:               entity.MikrotikStatusOffline,
	}

	// Set default timeout if not provided
	if mikrotik.Timeout == 0 {
		mikrotik.Timeout = 300000 // Default 5 minutes
	}

	if err := s.mikrotikRepo.Create(ctx, mikrotik); err != nil {
		return nil, fmt.Errorf("failed to create mikrotik: %w", err)
	}

	pkg_logger.Info("Mikrotik created",
		zap.String("id", mikrotik.ID),
		zap.String("name", mikrotik.Name),
		zap.String("host", mikrotik.Host),
	)

	return mikrotik, nil
}

// GetByID retrieves Mikrotik by ID
func (s *mikrotikUseCase) GetByID(ctx context.Context, id string) (*model.MikrotikResponse, error) {
	mk, err := s.mikrotikRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrMikrotikNotFound
		}
		return nil, fmt.Errorf("failed to get mikrotik: %w", err)
	}

	return s.toMikrotikResponse(mk), nil
}

// List retrieves paginated list of Mikrotiks
func (s *mikrotikUseCase) List(ctx context.Context, req model.PaginationRequest) (*model.PaginationResponse, error) {
	// Set default pagination
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 {
		req.PageSize = 10
	}

	mikrotiks, total, err := s.mikrotikRepo.List(ctx, req.Page, req.PageSize, req.Search)
	if err != nil {
		return nil, fmt.Errorf("failed to list mikrotiks: %w", err)
	}

	// Convert to response DTOs
	mkResponses := make([]model.MikrotikResponse, len(mikrotiks))
	for i, mk := range mikrotiks {
		mkResponses[i] = *s.toMikrotikResponse(&mk)
	}

	totalPages := int((total + int64(req.PageSize) - 1) / int64(req.PageSize))

	return &model.PaginationResponse{
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalItems: int(total),
		TotalPages: totalPages,
		Data:       mkResponses,
	}, nil
}

// Update updates existing Mikrotik configuration
func (s *mikrotikUseCase) Update(ctx context.Context, id string, req model.UpdateMikrotikRequest) error {
	mk, err := s.mikrotikRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return utils.ErrMikrotikNotFound
		}
		return fmt.Errorf("failed to get mikrotik: %w", err)
	}

	// Update fields if provided
	if req.Name != "" {
		mk.Name = req.Name
	}
	if req.Host != "" {
		mk.Host = req.Host
	}
	if req.Port > 0 {
		mk.Port = req.Port
	}
	if req.APIUsername != "" {
		mk.APIUsername = req.APIUsername
	}
	if req.APIPassword != "" {
		// Standard assignment (Plain Text)
		mk.APIEncryptedPassword = req.APIPassword
	}
	if req.Keepalive != nil {
		mk.Keepalive = *req.Keepalive
	}
	if req.Timeout > 0 {
		mk.Timeout = req.Timeout
	}
	if req.Location != "" {
		mk.Location = req.Location
	}
	if req.Description != "" {
		mk.Description = req.Description
	}
	if req.IsActive != nil {
		mk.IsActive = *req.IsActive
	}

	if err := s.mikrotikRepo.Update(ctx, mk); err != nil {
		return fmt.Errorf("failed to update mikrotik: %w", err)
	}

	pkg_logger.Info("Mikrotik updated",
		zap.String("id", mk.ID),
		zap.String("name", mk.Name),
	)

	return nil
}

// Delete deletes Mikrotik configuration
func (s *mikrotikUseCase) Delete(ctx context.Context, id string) error {
	if err := s.mikrotikRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete mikrotik: %w", err)
	}

	pkg_logger.Info("Mikrotik deleted", zap.String("id", id))
	return nil
}

// TestConnectionByID tests connection to specific Mikrotik by ID
func (s *mikrotikUseCase) TestConnectionByID(ctx context.Context, id string) error {
	mk, err := s.mikrotikRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return utils.ErrMikrotikNotFound
		}
		return fmt.Errorf("failed to get mikrotik: %w", err)
	}

	// Use password directly
	password := mk.APIEncryptedPassword

	pkg_logger.Debug("Decrypt TestConnectionByID", zap.String("password", password))

	// Create client and test connection
	client, err := mikrotik.NewClient(mikrotik.Config{
		Host:     mk.Host,
		Port:     mk.Port,
		Username: mk.APIUsername,
		Password: password,
		Timeout:  time.Duration(mk.Timeout) * time.Millisecond,
	})
	if err != nil {
		// Update status to offline on connection failure
		_ = s.mikrotikRepo.UpdateStatus(ctx, id, string(entity.MikrotikStatusOffline))
		return fmt.Errorf("failed to connect to mikrotik: %w", err)
	}
	defer client.Close()

	// Test connection by getting system resource
	_, err = client.Run("/system/resource/print")
	if err != nil {
		_ = s.mikrotikRepo.UpdateStatus(ctx, id, string(entity.MikrotikStatusError))
		return fmt.Errorf("failed to get system resource: %w", err)
	}

	// Update status to online on success
	_ = s.mikrotikRepo.UpdateStatus(ctx, id, string(entity.MikrotikStatusOnline))
	_ = s.mikrotikRepo.UpdateLastSync(ctx, id)

	pkg_logger.Info("Mikrotik connection test successful",
		zap.String("id", id),
		zap.String("host", mk.Host),
	)

	return nil
}

// UpdateStatus updates Mikrotik status
func (s *mikrotikUseCase) UpdateStatus(ctx context.Context, id string, status string) error {
	return s.mikrotikRepo.UpdateStatus(ctx, id, status)
}

// SetActiveMikrotik sets a Mikrotik as active (only one can be active at a time)
func (s *mikrotikUseCase) SetActiveMikrotik(ctx context.Context, id string) error {
	// Verify mikrotik exists
	_, err := s.mikrotikRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return utils.ErrMikrotikNotFound
		}
		return fmt.Errorf("failed to get mikrotik: %w", err)
	}

	// Deactivate all mikrotiks first
	if err := s.mikrotikRepo.DeactivateAll(ctx); err != nil {
		return fmt.Errorf("failed to deactivate all mikrotiks: %w", err)
	}

	// Activate the specified mikrotik
	if err := s.mikrotikRepo.SetActive(ctx, id, true); err != nil {
		return fmt.Errorf("failed to activate mikrotik: %w", err)
	}

	pkg_logger.Info("Mikrotik set as active", zap.String("id", id))
	return nil
}

// GetActiveMikrotik retrieves the currently active Mikrotik
func (s *mikrotikUseCase) GetActiveMikrotik(ctx context.Context) (*model.MikrotikStatusResponse, error) {
	mk, err := s.mikrotikRepo.GetActiveMikrotik(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("no active mikrotik found")
		}
		return nil, fmt.Errorf("failed to get active mikrotik: %w", err)
	}

	return &model.MikrotikStatusResponse{
		ID:       mk.ID,
		Name:     mk.Name,
		Host:     mk.Host,
		IsActive: mk.IsActive,
		Status:   string(mk.Status),
	}, nil
}

// GetMikrotikClient creates a new Mikrotik client connection to the active Mikrotik
func (s *mikrotikUseCase) GetMikrotikClient() (*mikrotik.Client, error) {
	// Get active mikrotik from database
	mk, err := s.mikrotikRepo.GetActiveMikrotik(context.Background())
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("no active mikrotik configured")
		}
		return nil, fmt.Errorf("failed to get active mikrotik: %w", err)
	}

	if !mk.IsActive {
		return nil, errors.New("mikrotik is not active")
	}

	pkg_logger.Debug("Active MK", zap.String("username", mk.APIUsername), zap.String("encrypted", mk.APIEncryptedPassword))

	// Use password directly
	password := mk.APIEncryptedPassword

	pkg_logger.Debug("Decrypt GetMikrotikClient", zap.String("password", password))

	// Create client
	client, err := mikrotik.NewClient(mikrotik.Config{
		Host:     mk.Host,
		Port:     mk.Port,
		Username: mk.APIUsername,
		Password: password,
		Timeout:  time.Duration(mk.Timeout) * time.Millisecond,
	})
	if err != nil {
		// Update mikrotik status to offline
		_ = s.mikrotikRepo.UpdateStatus(context.Background(), mk.ID, string(entity.MikrotikStatusOffline))
		return nil, fmt.Errorf("failed to connect to mikrotik: %w", err)
	}

	// Update mikrotik status to online and last sync time
	_ = s.mikrotikRepo.UpdateStatus(context.Background(), mk.ID, string(entity.MikrotikStatusOnline))
	_ = s.mikrotikRepo.UpdateLastSync(context.Background(), mk.ID)

	pkg_logger.Debug("Mikrotik client created successfully",
		zap.String("id", mk.ID),
		zap.String("host", mk.Host),
	)

	return client, nil
}

// GetClientByID retrieves a Mikrotik client by ID
func (s *mikrotikUseCase) GetClientByID(ctx context.Context, id string) (*mikrotik.Client, error) {
	mk, err := s.mikrotikRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrMikrotikNotFound
		}
		return nil, fmt.Errorf("failed to get mikrotik: %w", err)
	}

	// Use password directly (Plain Text)
	password := mk.APIEncryptedPassword

	// Create client
	client, err := mikrotik.NewClient(mikrotik.Config{
		Host:     mk.Host,
		Port:     mk.Port,
		Username: mk.APIUsername,
		Password: password,
		Timeout:  time.Duration(mk.Timeout) * time.Millisecond,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to mikrotik: %w", err)
	}

	return client, nil
}

// TestMikrotikConnection tests connection to the active Mikrotik
func (s *mikrotikUseCase) TestMikrotikConnection(ctx context.Context) error {
	client, err := s.GetMikrotikClient()
	if err != nil {
		return err
	}
	defer client.Close()

	// Try to get system resource as connection test
	_, err = client.Run("/system/resource/print")
	if err != nil {
		return fmt.Errorf("connection test failed: %w", err)
	}

	pkg_logger.Info("Mikrotik connection test successful")
	return nil
}

// Helper function to convert entity model to response DTO
func (s *mikrotikUseCase) toMikrotikResponse(mk *entity.Mikrotik) *model.MikrotikResponse {
	lastSync := ""
	if mk.LastSync != nil {
		lastSync = mk.LastSync.Format("2006-01-02 15:04:05")
	}

	return &model.MikrotikResponse{
		ID:          mk.ID,
		Name:        mk.Name,
		Host:        mk.Host,
		Port:        mk.Port,
		APIUsername: mk.APIUsername,
		Keepalive:   mk.Keepalive,
		Timeout:     mk.Timeout,
		Location:    mk.Location,
		Description: mk.Description,
		IsActive:    mk.IsActive,
		Status:      string(mk.Status),
		Version:     "", // Dynamic field, not in DB
		Uptime:      "", // Dynamic field, not in DB
		LastSync:    lastSync,
		CreatedAt:   mk.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:   mk.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}
