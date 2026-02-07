# API Documentation - MikroBill API

Dokumentasi lengkap untuk semua endpoint API MikroBill dengan penjelasan model, request/response, dan logika bisnis.

## Base URL
```
http://localhost:8080/v1
```

## Response Format
Semua response menggunakan format standar:
```json
{
  "success": true|false,
  "data": <response_data>,
  "metadata": {
    "total": 100,
    "limit": 10,
    "offset": 0
  },
  "error": "<error_message>"
}
```

---

## 1. Public Routes (`/v1/public`)

### 1.1 Get Tenant Info
**Endpoint:** `GET /v1/public/tenant/:slug`

**Deskripsi:** Mendapatkan informasi publik tenant berdasarkan slug untuk halaman registrasi

**Parameters:**
- `slug` (path) - Tenant slug

**Response Model:** `TenantPublicResponse`
```json
{
  "name": "PT Example",
  "slug": "example",
  "company_name": "PT Example Indonesia",
  "is_active": true,
  "status": "active",
  "features": {},
  "allows_signup": true
}
```

**Logic:**
- Validasi tenant aktif dan status = "active"
- Check fitur `public_registration` enabled
- Return informasi branding untuk registration page

---

### 1.2 Public Registration
**Endpoint:** `POST /v1/public/register/:tenant_slug`

**Middleware:** PublicRegistrationRateLimit

**Request Model:** `PublicRegistrationRequest`
```json
{
  "portal_email": "customer@example.com",
  "portal_password": "password123",
  "name": "John Doe",
  "phone": "081234567890",
  "email": "john@example.com",
  "address": "Jl. Example No. 123",
  "profile_id": "uuid",
  "service_type": "pppoe|hotspot|static_ip",
  "preferred_service_username": "john123"
}
```

**Response:** `CustomerResponse` dengan status `prospect`

**Logic:**
1. Validasi tenant exists dan allows public registration
2. Hash portal password (bcrypt)
3. Generate service credentials (auto untuk PPPoE, preferred untuk Hotspot)
4. Create customer dengan status `prospect`
5. Customer menunggu approval dari admin
6. Tidak langsung provision ke MikroTik

---

## 2. Auth Routes (`/v1/auth`)

### 2.1 Login
**Endpoint:** `POST /v1/auth/login`

**Request:**
```json
{
  "email": "user@example.com",
  "username": "username",
  "password": "password123"
}
```

**Response Model:** `EnhancedLoginResponse`
```json
{
  "access_token": "jwt_token",
  "refresh_token": "refresh_token",
  "token_type": "Bearer",
  "expires_in": 900,
  "refresh_expires_in": 2592000,
  "absolute_expires_in": 7776000,
  "user": { }
}
```

**Logic:**
1. Cari user by email OR username
2. Verify password (bcrypt)
3. Check user can login (status, lock)
4. Generate access token (15 min) dan refresh token (30 days)
5. Store refresh token metadata di Redis dengan absolute expiry 90 days
6. Update last_login_at dan last_ip
7. Reset failed_login_attempts

---

### 2.2 Register
**Endpoint:** `POST /v1/auth/register`

**Request Model:** `RegisterRequest`
```json
{
  "username": "newuser",
  "email": "new@example.com",
  "password": "password123",
  "fullname": "New User",
  "phone": "081234567890"
}
```

**Response:** `UserResponse`

**Logic:**
1. Validate unique username & email
2. Hash password (bcrypt)
3. Create user dengan role default `viewer`
4. Auto-assign tenant jika ada tenant context

---

### 2.3 Logout
**Endpoint:** `POST /v1/auth/logout`

**Middleware:** InternalAuth (optional)

**Request Model:** `LogoutRequest`
```json
{
  "refresh_token": "token"
}
```

**Logic:**
1. Get user_id dari JWT context (jika authenticated)
2. Invalidate refresh token di Redis
3. Remove from token family untuk prevent reuse

---

### 2.4 Refresh Token
**Endpoint:** `POST /v1/auth/refresh`

**Request Model:** `RefreshTokenRequest`
```json
{
  "refresh_token": "token"
}
```

**Response Model:** `RefreshTokenResponse`
```json
{
  "access_token": "new_jwt",
  "refresh_token": "new_refresh_token",
  "token_type": "Bearer",
  "expires_in": 900,
  "refresh_expires_in": 2592000,
  "rotation": true,
  "user": { }
}
```

**Logic:**
1. Validate refresh token di Redis
2. Check token belum expired (sliding window 30 days)
3. Check absolute expiry belum terlewati (90 days)
4. Detect reuse attack - jika parent token digunakan lagi, invalidate seluruh token family
5. **Token rotation:** Generate new refresh token setiap kali digunakan
6. Invalidate old refresh token
7. Generate new access token
8. Update metadata di Redis

---

### 2.5 Get Profile
**Endpoint:** `GET /v1/auth/profile`

**Middleware:** InternalAuth

**Response:** `UserResponse`

**Logic:**
1. Get user_id dari JWT context
2. Fetch user details dari database
3. Return user profile tanpa sensitive data

---

## 3. Tenant Management (`/v1/internal/tenant`)

**Middleware:** InternalAuth + RequireRole(superadmin)

### 3.1 Create Tenant
**Endpoint:** `POST /v1/internal/tenant`

**Request Model:** `CreateTenantRequest`
```json
{
  "name": "PT Example",
  "slug": "example",
  "subdomain": "example",
  "company_name": "PT Example Indonesia",
  "phone": "021xxx",
  "address": "Jakarta",
  "timezone": "Asia/Jakarta",
  "max_mikrotiks": 3,
  "max_network_users": 50,
  "max_staff_users": 5,
  "features": {},
  "metadata": {}
}
```

**Response:** `TenantResponse`

**Logic:**
1. Validate unique slug & subdomain
2. Set default limits dan features
3. Create tenant
4. Auto-create default roles untuk tenant
5. Setup tenant isolation

---

### 3.2 List Tenants
**Endpoint:** `GET /v1/internal/tenant/list`

**Query Params:** `TenantFilter`
- `status` - Filter by status
- `is_active` - Filter by active status
- `search` - Search name/company
- `limit` - Pagination limit
- `offset` - Pagination offset

**Response:** `TenantListResponse`

---

### 3.3 Get Tenant
**Endpoint:** `GET /v1/internal/tenant/:id`

**Response:** `TenantResponse`

---

### 3.4 Update Tenant
**Endpoint:** `PUT /v1/internal/tenant/:id`

**Request Model:** `UpdateTenantRequest`

**Logic:**
1. Validate tenant exists
2. Update fields yang disediakan
3. Log perubahan di audit log

---

### 3.5 Delete Tenant
**Endpoint:** `DELETE /v1/internal/tenant/:id`

**Logic:**
1. Soft delete tenant (set deleted_at)
2. Cascade soft delete ke semua resources terkait
3. Revoke semua user access ke tenant

---

### 3.6 Get Tenant Stats
**Endpoint:** `GET /v1/internal/tenant/:id/stats`

**Response Model:** `TenantStatsResponse`
```json
{
  "tenant_id": "uuid",
  "mikrotiks_count": 2,
  "network_users_count": 35,
  "staff_users_count": 3,
  "active_sessions_count": 20,
  "max_mikrotiks": 3,
  "max_network_users": 50,
  "max_staff_users": 5,
  "mikrotiks_usage_percent": 66.67,
  "network_users_usage_percent": 70.0,
  "staff_users_usage_percent": 60.0
}
```

**Logic:**
1. Count resources per tenant
2. Calculate usage percentage vs limits
3. Count active PPP sessions

---

## 4. User Management (`/v1/internal/users`)

**Middleware:** InternalAuth

### 4.1 Create User
**Endpoint:** `POST /v1/internal/users`

**Request Model:** `CreateUserRequest`

**Logic:**
1. Validate unique username & email
2. Hash password
3. Create user dengan tenant context (jika ada)
4. Assign role
5. Set created_by audit field

---

### 4.2 List Users
**Endpoint:** `GET /v1/internal/users/list`

**Response:** `UserListResponse`

**Logic:**
- Filter by tenant context (jika bukan superadmin)
- Pagination support

---

### 4.3 Get User
**Endpoint:** `GET /v1/internal/users/:id`

**Response:** `UserResponse`

---

### 4.4 Update User
**Endpoint:** `PUT /v1/internal/users/:id`

**Request Model:** `UpdateUserRequest`

---

### 4.5 Delete User
**Endpoint:** `DELETE /v1/internal/users/:id`

**Logic:** Soft delete

---

### 4.6 Assign Role
**Endpoint:** `POST /v1/internal/users/:id/assign-role`

**Request:**
```json
{
  "role_id": "uuid"
}
```

---

### 4.7 Assign to Tenant
**Endpoint:** `POST /v1/internal/users/:id/assign-tenant`

**Request:**
```json
{
  "tenant_id": "uuid",
  "role_id": "uuid",
  "is_primary": true
}
```

**Logic:**
1. Create TenantUser relationship
2. Set primary tenant jika is_primary=true
3. Assign role dalam konteks tenant

---

## 5. Client Routes (`/v1/internal`)

**Middleware:** InternalAuth + TenantContext + RequireTenantAccess

### 5.1 Client Upsert
**Endpoint:** `POST /v1/internal/client-upsert`

**Logic:** Insert or update client entity (legacy support)

---

### 5.2 Client Find
**Endpoint:** `POST /v1/internal/client-find`

---

### 5.3 Client Delete
**Endpoint:** `DELETE /v1/internal/client-delete`

---

## 6. Mikrotik Routes (`/v1/internal/mikrotik`)

**Middleware:** InternalAuth + TenantContext + RequireTenantAccess

### 6.1 Create Mikrotik
**Endpoint:** `POST /v1/internal/mikrotik`

**Request Model:** `CreateMikrotikRequest`
```json
{
  "name": "MikroTik Office",
  "host": "192.168.1.1",
  "port": 8728,
  "api_username": "admin",
  "api_password": "password",
  "keepalive": true,
  "timeout": 300000,
  "location": "Office",
  "description": "Main router"
}
```

**Response:** `MikrotikResponse`

**Logic:**
1. Check tenant limit untuk max_mikrotiks
2. Encrypt API password
3. Test connection ke MikroTik
4. Create mikrotik record
5. Status awal = "offline"
6. Jika connection success, update status = "online"

---

### 6.2 List Mikrotik
**Endpoint:** `POST /v1/internal/mikrotik/list`

**Response:** `MikrotikListResponse`

**Logic:**
- Filter by tenant_id
- Count by status (online/offline/error)

---

### 6.3 Get Active Mikrotik
**Endpoint:** `GET /v1/internal/mikrotik/active`

**Response:** `MikrotikResponse`

**Logic:**
- Return mikrotik dengan is_active=true untuk tenant
- Hanya boleh ada 1 active mikrotik per tenant

---

### 6.4 Get Mikrotik by ID
**Endpoint:** `GET /v1/internal/mikrotik/:id`

**Response:** `MikrotikResponse`

---

### 6.5 Update Mikrotik
**Endpoint:** `PUT /v1/internal/mikrotik/:id`

**Request Model:** `UpdateMikrotikRequest`

**Logic:**
1. Update fields
2. Jika password diupdate, re-encrypt
3. Test connection jika credentials berubah

---

### 6.6 Delete Mikrotik
**Endpoint:** `DELETE /v1/internal/mikrotik/:id`

**Logic:**
1. Check tidak ada customer aktif
2. Soft delete
3. Cleanup associated resources

---

### 6.7 Update Status
**Endpoint:** `PATCH /v1/internal/mikrotik/:id/status`

**Request:**
```json
{
  "status": "online|offline|error"
}
```

---

### 6.8 Set Active
**Endpoint:** `PATCH /v1/internal/mikrotik/:id/activate`

**Logic:**
1. Set is_active=false untuk semua mikrotik lain di tenant
2. Set is_active=true untuk mikrotik ini
3. **Business Rule:** Hanya 1 mikrotik yang boleh active per tenant

---

## 7. PPP Routes (`/v1/internal/ppp`)

**Middleware:** InternalAuth + TenantContext + RequireTenantAccess

### 7.1 PPP Secret Routes

#### Create Secret
**Endpoint:** `POST /v1/internal/ppp/secret`

**Request Model:** `PPPSecretInput`
```json
{
  "name": "user123",
  "password": "pass123",
  "profile": "default",
  "service": "pppoe",
  "local_address": "10.10.10.1",
  "remote_address": "10.10.10.2",
  "comment": "Customer XYZ"
}
```

**Logic:**
1. Get active mikrotik untuk tenant
2. Create PPP secret di MikroTik via API
3. Return response dengan MikroTik object ID

#### Get Secret
**Endpoint:** `GET /v1/internal/ppp/secret/:id`

**Logic:** Fetch secret dari MikroTik by ID

#### Update Secret
**Endpoint:** `PUT /v1/internal/ppp/secret/:id`

**Request Model:** `PPPSecretUpdateInput`

**Logic:** Update secret di MikroTik

#### Delete Secret
**Endpoint:** `DELETE /v1/internal/ppp/secret/:id`

**Logic:** Remove secret dari MikroTik

#### List Secrets
**Endpoint:** `GET /v1/internal/ppp/secret/list`

**Logic:** Fetch semua PPP secrets dari active MikroTik

---

### 7.2 PPP Profile Routes

#### Create Profile
**Endpoint:** `POST /v1/internal/ppp/profile`

**Request Model:** `PPPProfileInput`
```json
{
  "name": "10Mbps",
  "local_address": "10.10.10.1",
  "remote_address": "pool-pppoe",
  "rate_limit": "10M/10M",
  "only_one": "yes",
  "comment": "10Mbps Package"
}
```

**Logic:**
1. Get active mikrotik
2. **Auto-inject on-up dan on-down scripts** untuk callback
3. Create profile di MikroTik dengan scripts included
4. Scripts akan POST ke `/v1/callbacks/pppoe/up` dan `/v1/callbacks/pppoe/down`

#### Get Profile
**Endpoint:** `GET /v1/internal/ppp/profile/:id`

#### Update Profile
**Endpoint:** `PUT /v1/internal/ppp/profile/:id`

#### Delete Profile
**Endpoint:** `DELETE /v1/internal/ppp/profile/:id`

#### List Profiles
**Endpoint:** `GET /v1/internal/ppp/profile/list`

---

### 7.3 PPP Active Routes

#### List Active
**Endpoint:** `GET /v1/internal/ppp/active/list`

**Response:** Array of `PPPActive`

**Logic:**
- Fetch active PPP connections dari MikroTik
- Real-time data

#### Get Active by ID
**Endpoint:** `GET /v1/internal/ppp/active/:id`

#### Stream Active (WebSocket)
**Endpoint:** `GET /v1/internal/ppp/active/stream`

**Logic:**
- WebSocket connection
- Stream real-time PPP active updates
- Gunakan Redis Pub/Sub untuk broadcast
- Update setiap ada perubahan connection

---

### 7.4 PPP Inactive Routes

#### List Inactive
**Endpoint:** `GET /v1/internal/ppp/inactive/list`

**Logic:** List PPP secrets yang tidak sedang connected

#### Stream Inactive (WebSocket)
**Endpoint:** `GET /v1/internal/ppp/inactive/stream`

---

## 8. Pool Routes (`/v1/internal/pool`)

**Middleware:** InternalAuth + TenantContext + RequireTenantAccess

### 8.1 Create Pool
**Endpoint:** `POST /v1/internal/pool`

**Logic:** Create IP pool di MikroTik

### 8.2 Get Pool
**Endpoint:** `GET /v1/internal/pool/:id`

### 8.3 Update Pool
**Endpoint:** `PUT /v1/internal/pool/:id`

### 8.4 Delete Pool
**Endpoint:** `DELETE /v1/internal/pool/:id`

### 8.5 List Pools
**Endpoint:** `GET /v1/internal/pool/list`

---

## 9. Queue Routes (`/v1/internal/queue`)

**Middleware:** InternalAuth + TenantContext + RequireTenantAccess

### 9.1 Create Queue
**Endpoint:** `POST /v1/internal/queue`

### 9.2 Get Queue
**Endpoint:** `GET /v1/internal/queue/:id`

### 9.3 Update Queue
**Endpoint:** `PUT /v1/internal/queue/:id`

### 9.4 Delete Queue
**Endpoint:** `DELETE /v1/internal/queue/:id`

### 9.5 List Queues
**Endpoint:** `GET /v1/internal/queue/list`

---

## 10. Profile Routes (`/v1/internal/profile`)

**Middleware:** InternalAuth + TenantContext + RequireTenantAccess

Profile = Billing/Network profile (bukan PPP profile)

### 10.1 Create Profile
**Endpoint:** `POST /v1/internal/profile`

**Request Model:** `CreateProfileRequest`
```json
{
  "name": "10Mbps Package",
  "type": "pppoe",
  "rate_limit": "10M/10M",
  "session_timeout": 86400,
  "idle_timeout": 3600,
  "price": 150000.00,
  "tax_rate": 11.00,
  "local_address": "10.10.10.1",
  "remote_address": "pool-pppoe",
  "address_pool": "pool-pppoe",
  "mtu": 1480,
  "mru": 1480
}
```

**Response:** `ProfileResponse`

**Logic:**
1. Create base profile di database
2. Create type-specific details (PPPoE/Hotspot/StaticIP)
3. Profile ini digunakan untuk customer subscription
4. **Price dan tax_rate** disimpan di profile untuk billing

---

### 10.2 Get Profile
**Endpoint:** `GET /v1/internal/profile/:id`

**Logic:** Return profile dengan details sesuai type

### 10.3 List Profiles
**Endpoint:** `GET /v1/internal/profile/list`

**Response:** `ProfileListResponse`

### 10.4 Update Profile
**Endpoint:** `PUT /v1/internal/profile/:id`

### 10.5 Delete Profile
**Endpoint:** `DELETE /v1/internal/profile/:id`

**Logic:**
1. Check tidak ada service aktif menggunakan profile
2. Soft delete

---

## 11. Customer Routes (`/v1/internal/customer`)

**Middleware:** InternalAuth + TenantContext + RequireTenantAccess

### 11.1 Create Customer
**Endpoint:** `POST /v1/internal/customer`

**Request Model:** `CreateCustomerRequest`
```json
{
  "username": "customer123",
  "name": "John Doe",
  "phone": "081234567890",
  "email": "john@example.com",
  "address": "Jl. Example",
  "password": "servicepass123",
  "profile_id": "uuid",
  "service_type": "pppoe",
  "billing_day": 1,
  "auto_suspension": true,
  "start_date": "2024-01-01",
  "customer_notes": "VIP customer"
}
```

**Response:** `CustomerResponse`

**Logic:**
1. Get active mikrotik untuk tenant
2. Get profile details (untuk price, tax_rate, network config)
3. **Generate service credentials:**
   - PPPoE: Auto-generate username dan password
   - Hotspot: Bisa custom atau auto
   - Static IP: Generate identifier
4. **Encrypt service password**
5. Create customer record
6. Create CustomerService record dengan price dari profile
7. **Provisioning ke MikroTik:**
   - Create PPP secret jika PPPoE
   - Create Hotspot user jika Hotspot
   - Create IP binding jika Static IP
8. Set status = "active" jika provisioning sukses
9. Set provisioning_status = "completed"
10. Jika error, set provisioning_error

---

### 11.2 Get Customer
**Endpoint:** `GET /v1/internal/customer/:id`

**Response:** `CustomerResponse` dengan services dan current_service

---

### 11.3 List Customers
**Endpoint:** `GET /v1/internal/customer/list`

**Response:** `CustomerListResponse`

**Logic:**
- Filter by tenant
- Count by status (active/inactive/suspended/pending)

---

### 11.4 Update Customer
**Endpoint:** `PUT /v1/internal/customer/:id`

**Logic:**
1. Update customer info
2. Jika service_username atau password berubah, update di MikroTik
3. Update CustomerService jika ada perubahan billing

---

### 11.5 Delete Customer
**Endpoint:** `DELETE /v1/internal/customer/:id`

**Logic:**
1. Terminate service di MikroTik
2. Soft delete customer
3. End current service (set end_date)

---

### 11.6 List Prospects
**Endpoint:** `GET /v1/internal/customer/prospects/list`

**Logic:**
- List customers dengan status = "prospect"
- Customers yang mendaftar via public registration

---

### 11.7 Approve Prospect
**Endpoint:** `POST /v1/internal/customer/prospects/approve`

**Request Model:** `ApproveProspectRequest`
```json
{
  "customer_id": "uuid",
  "billing_day": 1,
  "auto_suspension": true,
  "start_date": "2024-01-01",
  "technician_note": "Installed on 2024-01-01",
  "preferred_service_username": "customuser"
}
```

**Logic:**
1. Get prospect customer
2. **Generate final service credentials** (jika belum ada)
3. Get profile untuk billing
4. Create CustomerService
5. **Provision ke MikroTik** (create PPP secret/Hotspot user)
6. Update status = "active"
7. Set provisioning_status = "completed"
8. Send welcome email/SMS dengan credentials

---

### 11.8 Reject Prospect
**Endpoint:** `DELETE /v1/internal/customer/prospects/:id/reject`

**Logic:**
1. Delete prospect customer
2. Send rejection notification

---

### 11.9 Stream Customer Traffic (WebSocket)
**Endpoint:** `GET /v1/internal/customer/:id/traffic/stream`

**Logic:**
- WebSocket stream traffic data untuk specific customer
- Real-time upload/download bytes
- Session duration
- Gunakan Redis untuk caching traffic data

---

### 11.10 Ping Customer
**Endpoint:** `GET /v1/internal/customer/:id/ping`

**Response:**
```json
{
  "host": "10.10.10.100",
  "alive": true,
  "rtt_avg": "5ms",
  "packet_loss": 0
}
```

**Logic:**
- Ping customer IP dari MikroTik
- Return latency dan packet loss

---

### 11.11 Stream Customer Ping (WebSocket)
**Endpoint:** `GET /v1/internal/customer/:id/ping/stream`

**Logic:** Continuous ping stream via WebSocket

---

## 12. Monitor Routes (`/v1/internal/monitor` & `/v1/internal/customer/:id`)

**Middleware:** InternalAuth + TenantContext + RequireTenantAccess

### 12.1 Stream Traffic by Interface
**Endpoint:** `GET /v1/internal/monitor/traffic/:interface`

**Logic:**
- WebSocket stream traffic untuk specific interface
- Real-time bandwidth monitoring

---

### 12.2 Log Streaming
**Endpoint:** `GET /v1/internal/logs/stream`

**Logic:**
- WebSocket stream MikroTik logs
- Real-time log monitoring

---

### 12.3 Direct Monitor Routes

#### Stream Traffic by Interface
**Endpoint:** `GET /v1/internal/monitor/direct/traffic/:interface/stream`

**Logic:** Direct traffic monitoring tanpa customer entity

#### Ping Host
**Endpoint:** `POST /v1/internal/monitor/direct/ping`

**Request:**
```json
{
  "host": "192.168.1.100",
  "count": 4
}
```

#### Stream Ping Host
**Endpoint:** `GET /v1/internal/monitor/direct/ping/:ip/stream`

---

## 13. Callback Routes (`/v1/callbacks`)

### 13.1 PPPoE Up
**Endpoint:** `POST /v1/callbacks/pppoe/up`

**Request Model:** `PPPoEEventInput`
```json
{
  "name": "user123",
  "caller_id": "00:11:22:33:44:55",
  "interface": "pppoe-user123",
  "local_address": "10.10.10.1",
  "remote_address": "10.10.10.100",
  "service": "pppoe"
}
```

**Logic:**
1. Find customer by service_username
2. Update customer assigned_ip dan last_online
3. Set status = "active" jika inactive
4. Publish event ke Redis untuk real-time updates
5. Log activity

---

### 13.2 PPPoE Down
**Endpoint:** `POST /v1/callbacks/pppoe/down`

**Request Model:** `PPPoEEventInput`

**Logic:**
1. Find customer by service_username
2. Update last_ip
3. Clear assigned_ip
4. Publish event ke Redis
5. Log activity

---

## 14. Customer Portal Routes

### 14.1 Customer Auth (`/v1/customer/auth`)

#### Customer Login
**Endpoint:** `POST /v1/customer/auth/login`

**Middleware:** CustomerLoginRateLimit

**Request Model:** `CustomerLoginRequest`
```json
{
  "email": "customer@example.com",
  "password": "password123"
}
```

**Response Model:** `CustomerLoginResponse`

**Logic:**
1. Find customer by portal_email
2. Verify portal_password
3. Generate JWT access token (15 min)
4. Generate refresh token (30 days)
5. Store session di Redis
6. Return tokens

---

#### Refresh Token
**Endpoint:** `POST /v1/customer/auth/refresh`

**Request:**
```json
{
  "refresh_token": "token"
}
```

**Logic:** Sama seperti internal auth refresh

---

#### Customer Logout
**Endpoint:** `POST /v1/customer/auth/logout`

**Middleware:** CustomerAuth

**Logic:** Invalidate customer session di Redis

---

#### Get Profile
**Endpoint:** `GET /v1/customer/auth/profile`

**Middleware:** CustomerAuth

**Response:** `CustomerPortalProfileResponse`

---

### 14.2 Customer Portal (`/v1/customer/portal`)

**Middleware:** CustomerAuth + CustomerPortalRateLimit + OriginValidation

#### Get Profile
**Endpoint:** `GET /v1/customer/portal/profile`

**Response:** `CustomerProfileResponse`

**Logic:**
- Return customer profile dengan conditional service password visibility
- Hotspot: show password
- PPPoE: hide password

---

#### Update Profile
**Endpoint:** `PUT /v1/customer/portal/profile`

**Request Model:** `UpdateCustomerProfileRequest`
```json
{
  "name": "New Name",
  "phone": "082xxx",
  "address": "New Address",
  "portal_password": "newpass123"
}
```

**Logic:**
1. Update basic info
2. Jika portal_password provided, hash dan update
3. NO update ke service credentials

---

#### Get Traffic
**Endpoint:** `GET /v1/customer/portal/traffic`

**Response:** `CustomerTrafficResponse`

**Logic:**
- Fetch current traffic stats dari MikroTik
- Cache di Redis (5 sec TTL)

---

#### Stream Traffic (WebSocket)
**Endpoint:** `GET /v1/customer/portal/traffic/stream`

**Logic:**
- Real-time traffic stream untuk authenticated customer
- Update interval: 2 seconds

---

#### Reset Session
**Endpoint:** `POST /v1/customer/portal/session/reset`

**Logic:**
1. Get customer service_username
2. Disconnect PPP/Hotspot session di MikroTik
3. Force re-authentication

---

#### Get Usage History
**Endpoint:** `GET /v1/customer/portal/usage`

**Response:** Array of `UsageHistoryResponse`

**Logic:**
- Return historical daily usage
- Data dari traffic snapshots stored di database

---

## 15. Legacy Client Route (`/v1/client`)

**Middleware:** ClientAuth

### Ping
**Endpoint:** `GET /v1/client/ping`

**Logic:** Simple ping endpoint untuk client compatibility check

---

## Middleware Explanation

### InternalAuth
- Validate JWT token
- Extract user_id, tenant_id, role
- Set context untuk request

### TenantContext
- Extract tenant_id dari user context atau header
- Validate tenant exists dan active
- Set tenant_id di context

### RequireTenantAccess
- Validate user has access ke tenant
- Check TenantUser relationship
- Enforce tenant isolation

### RequireRole(role)
- Validate user role
- Superadmin bypass semua checks

### CustomerAuth
- Validate customer JWT token
- Extract customer_id
- Set customer context

### Rate Limiting
- PublicRegistrationRateLimit: 5 req/hour per IP
- CustomerLoginRateLimit: 10 req/5min per IP
- CustomerPortalRateLimit: 100 req/min per customer

### OriginValidation
- Validate request origin dari allowed domains
- Prevent CSRF attacks

---

## Database Models Summary

### Core Entities
- **Tenant** - Multi-tenant isolation
- **User** - System users (admin/staff)
- **Customer** - Network customers
- **Mikrotik** - MikroTik devices
- **MikrotikProfile** - Billing profiles
- **CustomerService** - Service subscriptions
- **TenantUser** - User-Tenant relationships

### Service-Specific
- **MikrotikProfilePPPoE**
- **MikrotikProfileHotspot**
- **MikrotikProfileStaticIP**

### Audit & Sessions
- **ActivityLog**
- **UserSession** (Redis)
- **CustomerSession** (Redis)
- **RefreshTokenMetadata** (Redis)

---

## Business Rules

1. **Tenant Isolation:** Semua resources isolated per tenant
2. **Active Mikrotik:** Hanya 1 mikrotik active per tenant
3. **Profile Pricing:** Price & tax stored di profile, copied ke service saat creation
4. **Service Credentials:** Auto-generated untuk PPPoE, customizable untuk Hotspot
5. **Prospect Approval:** Public registration creates prospect, admin must approve
6. **Token Rotation:** Refresh tokens rotated setiap digunakan
7. **Reuse Detection:** Parent token reuse invalidates entire token family
8. **Absolute Expiry:** Session tidak bisa extend beyond 90 days
9. **Auto Suspension:** Customer auto-suspended jika billing overdue (jika enabled)
10. **Provisioning Status:** Track provisioning success/failure

---

## Error Handling

Semua error dikembalikan dengan format:
```json
{
  "success": false,
  "error": "Error message"
}
```

HTTP Status Codes:
- 200: Success
- 201: Created
- 400: Bad Request (validation error)
- 401: Unauthorized (auth required)
- 403: Forbidden (insufficient permission)
- 404: Not Found
- 500: Internal Server Error
