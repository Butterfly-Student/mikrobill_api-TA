# Task: Enhance Multi-Tenant ISP System - Public Registration & Customer Portal

## Overview
Implement public registration and customer portal features leveraging existing Redis (caching) and RabbitMQ (worker) infrastructure. Excludes invoicing and payment gateways.

## Task Breakdown

### [/] Phase 1: Infrastructure Audit & Planning
- [x] Analyze existing Redis implementation
- [x] Analyze existing RabbitMQ implementation  
- [x] Review current "listening" logic for traffic data
- [x] Review Customer domain and prospect flow
- [x] Create implementation plan
- [ ] Get user approval on plan

### [x] Phase 2: Database Schema & Models  
- [x] Create migration `14_customer_portal_credentials.go`
  - [x] Add `portal_email` and `portal_password_hash` to customers table
  - [x] Rename `username` → `service_username` (clarity)
  - [x] Add `service_password_encrypted` and `service_password_visible` columns
  - [x] Create UNIQUE constraint on (`tenant_id`, `portal_email`, `deleted_at`)
  - [x] Create `customer_sessions` table
  - [x] Add provisioning status fields
- [x] Update Customer model with separated credentials
- [x] Create CustomerLoginRequest, CustomerPortalProfileResponse models
- [x] Create encryption utility service (AES-256-GCM)

### [/] Phase 3: Public Registration Module (Portal Credentials Only)
- [x] Repository layer updates
  - [x] Add GetByPortalEmail to CustomerDatabasePort
  - [x] Add GetByServiceUsername to CustomerDatabasePort  
  - [x] Add UpdatePortalPassword to CustomerDatabasePort
  - [x] Add UpdateServiceCredentials to CustomerDatabasePort
  - [x] Add UpdateProvisioningStatus to CustomerDatabasePort
  - [x] Implement all new methods in postgres adapter
  - [x] Update CreateProspect to use portal credentials
- [x] Domain layer updates
  - [x] Implement RegisterProspect() with bcrypt password hashing
  - [x] Implement ApproveProspect() with service credential generation
  - [x] Generate PPPoE username (auto, hidden) & password (16-char strong)
  - [x] Generate Hotspot username (custom/auto, visible) & password (8-char simple)
  - [x] Encrypt service passwords with AES-256
- [x] HTTP layer (Gin handlers & routes)
  - [x] GET `/public/tenant/:slug` - tenant branding info
  - [x] POST `/public/register/:slug` - public registration endpoint
  - [x] Add routes without authentication middleware

### [x] Phase 4: Provisioning Flow (Service Credential Generation + RabbitMQ)
- [x] Update ApproveProspect domain method
  - [x] Generate service_username (auto or based on pattern)
  - [x] Generate service_password (strong for PPPoE, simpler for Hotspot)
  - [x] Set service_password_visible (FALSE for PPPoE, TRUE for Hotspot)
  - [x] Encrypt service_password and store in database
  - [x] Publish RabbitMQ message with service credentials
  - [x] Do NOT call MikroTik API directly
- [x] Create provisioning worker (RabbitMQ consumer)
  - [x] Subscribe to `customer.provisioning` queue
  - [x] Receive: {customer_id, service_username, service_password, service_type, profile_id}
  - [x] Select optimal MikroTik (load balancing across 3)
  - [x] Execute RouterOS API `/ppp/secret/add` or `/ip/hotspot/user/add`
  - [x] Update customer status to ACTIVE on success
  - [x] Record mikrotik_id and mikrotik_object_id
  - [x] Handle provisioning errors (retry logic)
- [x] Add rabbitmq route for provisioning worker
- [x] Update app.go to start provisioning worker

### [x] Phase 5: Customer Portal Authentication (Portal Credentials)
- [x] Create Customer Auth domain
  - [x] Implement CustomerLogin method (portal_email + portal_password)
  - [x] Verify against portal_password_hash (bcrypt)
  - [x] Implement CustomerLogout method
  - [x] Implement CustomerRefreshToken method
  - [x] Use Redis for session storage (similar to Admin auth)
- [x] Create Customer Auth repository
  - [x] GetByPortalEmail for login
  - [x] Validate password hash
- [x] Create Customer Auth HTTP handlers
  - [x] POST `/v1/customer/auth/login`
  - [x] POST `/v1/customer/auth/logout`
  - [x] POST `/v1/customer/auth/refresh`
  - [x] GET `/v1/customer/auth/profile`
- [x] Create CustomerAuth middleware
  - [x] Validate customer JWT token
  - [x] Inject customer context
  - [x] Ensure tenant_id isolation

### [x] Phase 6: Customer Portal API (Service Credential Visibility)
- [x] Create customer portal routes group
- [x] Implement profile endpoint with credential visibility
  - [x] GET `/v1/customer/portal/profile`
  - [x] Show service_username for all service types
  - [x] Decrypt and show service_password ONLY if service_password_visible=TRUE (Hotspot)
  - [x] Hide service_password for PPPoE (service_password_visible=FALSE)
  - [x] PUT `/v1/customer/portal/profile` (name, phone, address, portal_password only)
- [x] Implement traffic monitoring endpoint
  - [x] Create Redis cache aggregator for traffic data
  - [x] GET `/v1/customer/portal/traffic` (fetch from Redis)
  - [ ] WebSocket `/v1/customer/portal/traffic/stream` (placeholder - to be implemented)
- [x] Implement session reset endpoint  
  - [x] POST `/v1/customer/portal/session/reset`
  - [x] Publish RabbitMQ task with service_username (not portal_email)
  - [ ] Create disconnect worker (exists from Phase 4)
- [x] Implement usage history endpoint
  - [x] GET `/v1/customer/portal/usage/history`

### [x] Phase 7: Traffic Data Aggregation (Redis Cache Integration)
- [x] Leverage existing `StreamTraffic` listening mechanism
- [x] Integrate Redis storage in `publishTrafficData()`
  - [x] Store traffic data with key: `traffic:{tenant_id}:{service_username}`
  - [x] Set TTL to 24 hours for automatic cleanup
  - [x] Convert bits/sec to bytes for storage
- [x] Customer portal reads from Redis cache (no MikroTik query needed)
- [x] Uses MikroTik `/interface/monitor-traffic` LISTEN (not polling)

### [ ] Phase 8: Security & Middleware Audit
- [ ] Review CustomerAuth middleware for tenant isolation
- [ ] Ensure customers can only access own data
- [ ] Add rate limiting to public registration endpoint
- [ ] Add rate limiting to customer portal endpoints
- [ ] Review password hashing algorithm (bcrypt)
- [ ] Add CSRF protection for customer portal

### [ ] Phase 9: Testing & Verification
- [ ] Test public registration flow (all service types)
- [ ] Test admin approval workflow
- [ ] Test RabbitMQ provisioning worker
- [ ] Test customer portal login
- [ ] Test customer portal traffic monitoring
- [ ] Test session reset functionality
- [ ] Test tenant isolation for customers
- [ ] Performance test Redis traffic aggregation

### [ ] Phase 10: Documentation
- [ ] Update API.md with new endpoints
- [ ] Document RabbitMQ message formats
- [ ] Document Redis cache keys
- [ ] Create customer portal API documentation
- [ ] Update README with provisioning worker setup
