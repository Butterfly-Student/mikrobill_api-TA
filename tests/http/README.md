# HTTP API Test Files

Folder ini berisi HTTP test files untuk testing semua API endpoints menggunakan REST Client extension di VS Code.

## Requirements

Install extension [REST Client](https://marketplace.visualstudio.com/items?itemName=humao.rest-client) di VS Code.

## Configuration

File `http-client.env.json` berisi environment variables untuk base URL dan credentials default.

## Test Files

### Authentication & Authorization
- **auth.http** - User authentication (login, register, refresh, logout)
- **public.http** - Public endpoints (tenant info, customer registration)

### Admin Management  
- **tenant.http** - Tenant management (SuperAdmin only)
- **user.http** - User management (CRUD, assign roles/tenants)

### Network Configuration
- **mikrotik.http** - MikroTik device management
- **ppp.http** - PPP secrets, profiles, active/inactive connections
- **pool.http** - IP pool management
- **queue.http** - Queue management
- **profile.http** - Billing profiles (PPPoE/Hotspot/Static IP)

### Customer Management
- **customer.http** - Customer CRUD, prospect approval/rejection, monitoring

### Monitoring & Callbacks
- **monitor.http** - Traffic monitoring, log streaming, direct monitoring
- **callback.http** - PPPoE connection callbacks (from MikroTik)

### Customer Portal
- **customer-portal.http** - Customer-facing portal endpoints

## Usage

1. **Start the server:**
   ```bash
   go run cmd/main.go
   ```

2. **Open any `.http` file** in VS Code

3. **Click "Send Request"** on the line above endpoint definition

4. **Variable chaining** akan otomatis capture tokens dan IDs dari responses

## Test Flow Example

```
1. auth.http
   └─> Login → Capture @token

2. tenant.http (SuperAdmin only)
   └─> Use @token → Create tenant → Capture @tenantId

3. mikrotik.http
   └─> Use @token → Create mikrotik → Activate

4. profile.http
   └─> Create billing profile → Capture @profileId

5. customer.http
   └─> Create customer using @profileId → Provisions to MikroTik

6. monitor.http
   └─> Monitor customer traffic
```

## Variable Chaining

Test files menggunakan variable chaining untuk mengurangi manual editing:

```http
# @name loginSuperAdmin
POST {{baseUrl}}/v1/auth/login
{
  "username": "superadmin@mikrobill.com",
  "password": "password123"
}

@token = {{loginSuperAdmin.response.body.data.access_token}}

### Subsequent requests use @token
GET {{baseUrl}}/v1/auth/profile
Authorization: Bearer {{token}}
```

## Notes

- **Authentication:** Hampir semua endpoints require authentication
- **Tenant Context:** Internal endpoints require tenant context dari user
- **Role-Based:** Tenant management requires SuperAdmin role
- **WebSocket:** Beberapa endpoints (traffic stream, log stream) menggunakan WebSocket
- **Error Cases:** Setiap file include test cases untuk error scenarios

## Cleanup

Old `.rest` files dapat dihapus setelah verify bahwa `.http` files berfungsi dengan baik.
