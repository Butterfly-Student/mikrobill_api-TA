# MikroOps API Documentation

Comprehensive API documentation for MikroOps Multi-Tenant Billing & Monitoring System.

## Table of Contents
1. [Base Configuration](#base-configuration)
2. [Authentication](#1-authentication)
3. [Tenant Management](#2-tenant-management)
4. [User Management](#3-user-management)
5. [Client Operations](#4-client-operations)
6. [MikroTik Device Management](#5-mikrotik-device-management)
7. [PPP Management](#6-ppp-management)
8. [Other Resouces](#7-other-resources)
9. [Billing Profiles](#8-billing-profiles)
10. [Customer Management](#9-customer-management)
11. [Monitoring & Real-time Stats](#10-monitoring--real-time-stats)
12. [Callbacks & System](#11-callbacks--system)

---

## Base Configuration
*   **Base URL:** `http://localhost:8000/v1`
*   **Format:** All responses follow the standard `model.Response` structure.

### Standard Response Structure
```json
{
  "success": true,
  "data": null,
  "metadata": {
    "total": 0,
    "limit": 0,
    "offset": 0
  },
  "error": null
}
```

### Required Headers
*   `Authorization`: `Bearer <access_token>` (Required for all restricted routes)
*   `Content-Type`: `application/json`
*   `X-Tenant-ID`: `<uuid>`
    *   **Required** for **Super Admins** to switch context to a specific tenant.
    *   Derived automatically for regular Users (Admins/Viewers) based on their assigned tenant.

---

## 1. Authentication
Public endpoints for identity management.

### Login
*   **Path:** `POST /auth/login`
*   **Request Body:**
    ```json
    {
      "email": "admin@example.com", // Optional if username is provided
      "username": "admin",         // Optional if email is provided
      "password": "password123"
    }
    ```
*   **Response (Data):**
    ```json
    {
      "access_token": "jwt_token_string",
      "refresh_token": "uuid_string",
      "token_type": "Bearer",
      "expires_in": 900,           // 15 minutes
      "refresh_expires_in": 2592000, // 30 days
      "user": {
        "id": "uuid",
        "username": "admin",
        "email": "admin@example.com",
        "fullname": "Administrator",
        "user_role": "super_admin",
        "tenant_id": null
      }
    }
    ```

### Refresh Token
*   **Path:** `POST /auth/refresh`
*   **Request Body:**
    ```json
    {
      "refresh_token": "uuid_string"
    }
    ```
*   **Response (Data):**
    ```json
    {
      "access_token": "new_jwt_token",
      "token_type": "Bearer",
      "expires_in": 900,
      "rotation": true,
      "refresh_token": "new_uuid_string", // Only if rotated
      "refresh_expires_in": 2592000
    }
    ```

### Logout
*   **Path:** `POST /auth/logout`
*   **Request Body:**
    ```json
    {
      "refresh_token": "uuid_string"
    }
    ```
*   **Response (Data):** `{"message": "logged out successfully"}`

### Get Profile
*   **Path:** `GET /auth/profile`
*   **Headers:** `Authorization: Bearer <access_token>`
*   **Response (Data):** Returns current authenticated `User` object.

### Register
*   **Path:** `POST /auth/register`
*   **Request Body (`model.RegisterRequest`):**
    ```json
    {
      "username": "newuser",
      "email": "user@example.com",
      "password": "strongpassword",
      "fullname": "New User",
      "phone": "08123456789"
    }
    ```
*   **Response (Data):** Returns the created `User` object.

---

## 2. Tenant Management
Super Admin routes for managing platform tenants.

### Create Tenant
*   **Path:** `POST /internal/tenant`
*   **Request Body (`model.CreateTenantRequest`):**
    ```json
    {
      "name": "ISP Bandung",
      "subdomain": "ispbdg",
      "company_name": "PT ISP Digital",
      "phone": "022-123456",
      "timezone": "Asia/Jakarta",
      "max_mikrotiks": 10,
      "max_network_users": 5000,
      "features": { "api_access": true }
    }
    ```

### List Tenants
*   **Path:** `GET /internal/tenant/list`
*   **Query Params:** `status`, `is_active`, `search`, `limit`, `offset`
*   **Response (Data):** `[]model.TenantResponse`

### Get Tenant
*   **Path:** `GET /internal/tenant/:id`
*   **Response (Data):** Full `model.TenantResponse`.

### Update Tenant
*   **Path:** `PUT /internal/tenant/:id`

### Delete Tenant
*   **Path:** `DELETE /internal/tenant/:id`

### Get Tenant Stats
*   **Path:** `GET /internal/tenant/:id/stats`

---

## 3. User Management
Routes for managing dashboard users (Admins, Operators, Viewers).
*   **Super Admin:** Can manage all users across all tenants.
*   **Tenant Admin:** Can manage users only within their own tenant.

### Create User
*   **Path:** `POST /internal/users`
*   **Request Body (`model.CreateUserRequest`):**
    ```json
    {
      "username": "tenant_admin",
      "email": "admin@isp.com",
      "password": "securePass123!",
      "fullname": "Tenant Administrator",
      "phone": "08123456789",
      "user_role": "admin",
      "role_id": "uuid_role" // Optional RBAC role
    }
    ```

### List Users
*   **Path:** `GET /internal/users/list`
*   **Query Params:** `limit`, `offset`, `tenant_id` (Super Admin only)
*   **Response (Data):** `{"users": [], "total": 10}`

### Get User
*   **Path:** `GET /internal/users/:id`

### Update User
*   **Path:** `PUT /internal/users/:id`
*   **Request Body:** `model.UpdateUserRequest` (Partial fields)

### Delete User
*   **Path:** `DELETE /internal/users/:id`

### Assign Role
*   **Path:** `POST /internal/users/:id/assign-role`
*   **Request Body:** `{"role_id": "uuid"}`

### Assign to Tenant
*   **Path:** `POST /internal/users/:id/assign-tenant`
*   **Note:** Super Admin only.
*   **Request Body:** `{"tenant_id": "uuid", "is_primary": true}`

---

## 4. Client Operations
Generic operations for managing linked clients/entities.

### Upsert Clients
*   **Path:** `POST /internal/client-upsert`
*   **Request Body:** `[]model.ClientInput`

### Find Clients
*   **Path:** `POST /internal/client-find`
*   **Request Body:** `model.ClientFilter`

### Delete Client
*   **Path:** `DELETE /internal/client-delete`

---

## 5. MikroTik Device Management
Requires Tenant Context. Traced via `MikrotikHttpPort`.

### Create Device
*   **Path:** `POST /internal/mikrotik`
*   **Request Body (`model.CreateMikrotikRequest`):**
    ```json
    {
      "name": "Router-01",
      "host": "103.11.22.33",
      "port": 8728,
      "api_username": "billing",
      "api_password": "securepassword",
      "location": "IDC Cyber"
    }
    ```

### List Devices
*   **Path:** `POST /internal/mikrotik/list`
*   **Response (Data):** `[]model.MikrotikResponse`

### Get Device By ID
*   **Path:** `GET /internal/mikrotik/:id`

### Update Device
*   **Path:** `PUT /internal/mikrotik/:id`

### Delete Device
*   **Path:** `DELETE /internal/mikrotik/:id`

### Update Status
*   **Path:** `PATCH /internal/mikrotik/:id/status`

### Set Active Device
*   **Path:** `PATCH /internal/mikrotik/:id/activate`

---

## 6. PPP Management
Direct MikroTik API calls for PPP Secrets, Profiles, and Active Connections.

### Create PPP Secret
*   **Path:** `POST /internal/ppp/secret`
*   **Request Body (`model.PPPSecretInput`):**
    ```json
    {
      "name": "customer_user",
      "password": "secretpassword",
      "profile": "Plan-10M",
      "service": "pppoe",
      "local_address": "10.0.0.1", // Optional
      "remote_address": "10.0.0.2", // Optional
      "comment": "John Doe"
    }
    ```

### Get PPP Secret
*   **Path:** `GET /internal/ppp/secret/:id`

### Update PPP Secret
*   **Path:** `PUT /internal/ppp/secret/:id`

### Delete PPP Secret
*   **Path:** `DELETE /internal/ppp/secret/:id`

### List PPP Secrets
*   **Path:** `GET /internal/ppp/secret/list`

### Create PPP Profile
*   **Path:** `POST /internal/ppp/profile`
*   **Request Body (`model.PPPProfileInput`):**
    ```json
    {
      "name": "Plan-10M",
      "local_address": "10.0.0.1",
      "remote_address": "pool-pppoe",
      "rate_limit": "10M/10M",
      "only_one": "yes"
    }
    ```

### Get PPP Profile
*   **Path:** `GET /internal/ppp/profile/:id`

### Update PPP Profile
*   **Path:** `PUT /internal/ppp/profile/:id`

### Delete PPP Profile
*   **Path:** `DELETE /internal/ppp/profile/:id`

### List PPP Profiles
*   **Path:** `GET /internal/ppp/profile/list`

### List PPP Active Connections
*   **Path:** `GET /internal/ppp/active/list`
*   **Response:** List of `model.PPPActive`

### Get PPP Active Connection
*   **Path:** `GET /internal/ppp/active/:id`

### WebSocket: Stream PPP Active Connections
*   **Path:** `GET /internal/ppp/active/stream`
*   **Protocol:** WebSocket
*   **Description:** Streams real-time updates of active PPP connections. Initial message contains full list, subsequent messages contain updates.
*   **Event Structure:**
    ```json
    {
      "type": "update",
      "data": {
        "action": "up", // or "down"
        "name": "user1",
        "caller_id": "AA:BB:CC:DD:EE:FF",
        "interface": "pppoe-in1",
        "remote_address": "10.0.0.50",
        "service": "pppoe"
      }
    }
    ```

### List PPP Inactive Connections
*   **Path:** `GET /internal/ppp/inactive/list`

### WebSocket: Stream PPP Inactive Connections
*   **Path:** `GET /internal/ppp/inactive/stream`
*   **Protocol:** WebSocket

---

## 7. Other Resources
Pool and Queue management.

### IP Pools
*   **List:** `GET /internal/pool/list`
*   **Get:** `GET /internal/pool/:id`
*   **Create:** `POST /internal/pool`
*   **Update:** `PUT /internal/pool/:id`
*   **Delete:** `DELETE /internal/pool/:id`

### Queues
*   **List:** `GET /internal/queue/list`
*   **Get:** `GET /internal/queue/:id`
*   **Create:** `POST /internal/queue`
*   **Update:** `PUT /internal/queue/:id`
*   **Delete:** `DELETE /internal/queue/:id`

---

## 8. Billing Profiles
System-level billing plans mapped to MikroTik configurations.

### Create Profile
*   **Path:** `POST /internal/profile`
*   **Request Body (`model.CreateProfileRequest`):**
    ```json
    {
      "name": "Premium 50M",
      "type": "pppoe",
      "price": 350000,
      "rate_limit": "50M/50M",
      "local_address": "10.0.0.1",
      "remote_address": "pool-pppoe"
    }
    ```

### List Profiles
*   **Path:** `GET /internal/profile/list`
*   **Response:** List of defined profiles available for customer assignment.

### Get, Update, Delete Profile
*   `GET /internal/profile/:id`
*   `PUT /internal/profile/:id`
*   `DELETE /internal/profile/:id`

---

## 9. Customer Management
End-user subscription management.

### Create Customer
*   **Path:** `POST /internal/customer`
*   **Request Body (`model.CreateCustomerRequest`):**
    ```json
    {
      "username": "johndoe",
      "name": "John Doe",
      "phone": "0812345678",
      "password": "p",
      "profile_id": "uuid", // Links to Billing Profile
      "service_type": "pppoe",
      "billing_day": 5
    }
    ```

### List Customers
*   **Path:** `GET /internal/customer/list`

### Get, Update, Delete Customer
*   `GET /internal/customer/:id`
*   `PUT /internal/customer/:id`
*   `DELETE /internal/customer/:id`

---

## 10. Monitoring & Real-time Stats
WebSocket and high-frequency polling routes.

### Real-time Traffic Stream
*   **Path:** `GET /internal/customer/:id/traffic/stream`
*   **Protocol:** WebSocket
*   **Description:** Live traffic bandwidth monitoring for specific customer interface.

### On-Demand Ping
*   **Path:** `GET /internal/customer/:id/ping`

### Ping Stream
*   **Path:** `GET /internal/customer/:id/ping/stream`
*   **Protocol:** WebSocket

### Direct Interface Traffic
*   **Path:** `GET /internal/monitor/traffic/:interface`
*   **Path:** `GET /internal/monitor/direct/traffic/:interface/stream`

### Direct Host Ping
*   **Path:** `POST /internal/monitor/direct/ping`
*   **Path:** `GET /internal/monitor/direct/ping/:ip/stream`

### Log Streaming
*   **Path:** `GET /internal/logs/stream`
*   **Protocol:** WebSocket
*   **Description:** Live stream of MikroTik logs.

---

## 11. Callbacks & System
Endpoints for external system integration (MikroTik Scripts).

### PPPoE On-Up Callback
*   **Path:** `POST /v1/callbacks/pppoe/up`
*   **Description:** Triggered by MikroTik when a PPPoE user connects.
*   **Request Body (`model.PPPoEEventInput`):**
    ```json
    {
      "name": "user1",
      "caller_id": "AA:BB:CC:DD:EE:FF",
      "interface": "pppoe-in1",
      "local_address": "10.0.0.1",
      "remote_address": "10.0.0.50",
      "service": "pppoe"
    }
    ```
*   **Effect:** Updates internal database, invalidates cache, and broadcasts to WebSocket listeners.

### PPPoE On-Down Callback
*   **Path:** `POST /v1/callbacks/pppoe/down`
*   **Description:** Triggered by MikroTik when a PPPoE user disconnects.
*   **Request Body:** Same as On-Up.

### System Resource Ping
*   **Path:** `GET /client/ping`