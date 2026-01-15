# MikroBill API Documentation

## Table of Contents
1. [General Information](#general-information)
2. [Authentication](#authentication)
3. [Internal Management](#internal-management)
   - [Tenant Management](#tenant-management)
   - [Client Management](#client-management)
   - [MikroTik Devices](#mikrotik-devices)
   - [PPP Secrets Management](#ppp-secrets-management)
   - [PPP Profiles Management](#ppp-profiles-management)
   - [System Profiles Management](#system-profiles-management)
   - [Customer Management](#customer-management)
   - [Monitoring](#monitoring)
4. [Callbacks](#callbacks)
5. [V1 Client API](#v1-client-api)

---

## General Information

### Base URL
The API base URL is: `/v1`

### Authentication
- Internal endpoints require `X-Client-Key` header or session authentication.
- Authorization is enforced by Tenant ID and User Role.
- Client API endpoints require client authentication via `X-Client-Key`.

### Response Format
All responses follow a standard structure:
```json
{
  "success": true,
  "data": {},
  "metadata": {
    "total": 100,
    "limit": 10,
    "offset": 0
  },
  "error": null
}
```

### WebSocket Endpoints
- `/v1/internal/monitor/traffic/:interface` - Real-time traffic monitoring
- `/v1/internal/customer/:id/traffic/stream` - Customer-specific traffic streaming
- `/v1/internal/customer/:id/ping/stream` - Continuous ICMP ping monitoring

---

## Authentication

### POST `/v1/auth/login`
**Description:** Authenticate a user and receive a session token.
**No Authentication Required**
**Request Body:**
```json
{
  "username": "admin",
  "password": "password"
}
```
**Response Data:**
```json
{
  "id": "uuid",
  "username": "admin",
  "email": "admin@example.com",
  "role": "admin",
  "api_token": "token_string"
}
```

### POST `/v1/auth/register`
**Description:** Register a new user.
**No Authentication Required**
**Request Body:**
```json
{
  "username": "newuser",
  "email": "user@example.com",
  "password": "securepassword",
  "fullname": "New User",
  "phone": "08123456789",
  "role_id": "uuid"
}
```

---

## Internal Management

### Tenant Management

#### POST `/v1/internal/tenant`
**Description:** Create a new tenant.
**Authentication Required:** Internal Auth (Super Admin)

#### GET `/v1/internal/tenant/list`
**Description:** List all tenants with pagination.

#### GET `/v1/internal/tenant/:id`
**Description:** Get tenant details by ID.

#### PUT `/v1/internal/tenant/:id`
**Description:** Update tenant configuration and limits.

#### DELETE `/v1/internal/tenant/:id`
**Description:** Soft-delete a tenant.

#### GET `/v1/internal/tenant/:id/stats`
**Description:** Get usage statistics for a tenant (device count, user count, etc.).

### Client Management

#### POST `/v1/internal/client-upsert`
**Description:** Create or update a client (API consumer).

#### POST `/v1/internal/client-find`
**Description:** Search for clients by name or ID.

#### DELETE `/v1/internal/client-delete`
**Description:** Delete clients by list of IDs.

### MikroTik Devices

#### POST `/v1/internal/mikrotik`
**Description:** Add a new MikroTik device to the system.

#### POST `/v1/internal/mikrotik/list`
**Description:** List MikroTik devices with filters.

#### GET `/v1/internal/mikrotik/active`
**Description:** Get the currently active/selected MikroTik device for the browser session.

#### GET `/v1/internal/mikrotik/:id`
**Description:** Get device details.

#### PUT `/v1/internal/mikrotik/:id`
**Description:** Update device connection details.

#### DELETE `/v1/internal/mikrotik/:id`
**Description:** Remove device from system.

#### PATCH `/v1/internal/mikrotik/:id/status`
**Description:** Update device reachability status.

#### PATCH `/v1/internal/mikrotik/:id/activate`
**Description:** Set this device as the active gateway for subsequent commands.

### PPP Secrets Management

#### POST `/v1/internal/ppp/secret`
**Description:** Create a direct PPP secret on MikroTik.

#### GET `/v1/internal/ppp/secret/:id`
**Description:** Fetch secret directly from RouterOS.

#### PUT `/v1/internal/ppp/secret/:id`
**Description:** Update secret configuration.

#### DELETE `/v1/internal/ppp/secret/:id`
**Description:** Remove secret from MikroTik.

#### GET `/v1/internal/ppp/secret/list`
**Description:** List all secrets from the active MikroTik.

### PPP Profiles Management

#### POST `/v1/internal/ppp/profile`
**Description:** Create a direct PPP profile on MikroTik.

#### GET `/v1/internal/ppp/profile/:id`
**Description:** Fetch profile from RouterOS.

#### PUT `/v1/internal/ppp/profile/:id`
**Description:** Update profile configuration.

#### DELETE `/v1/internal/ppp/profile/:id`
**Description:** Remove profile from MikroTik.

#### GET `/v1/internal/ppp/profile/list`
**Description:** List all profiles from the active MikroTik.

### System Profiles Management
*Note: These are managed by the application and synchronized to MikroTik.*

#### POST `/v1/internal/profile`
**Description:** Create a system profile (Plan).

#### GET `/v1/internal/profile/list`
**Description:** List all available plans.

#### GET `/v1/internal/profile/:id`
**Description:** Get plan details.

#### PUT `/v1/internal/profile/:id`
**Description:** Update plan and sync change to MikroTik.

#### DELETE `/v1/internal/profile/:id`
**Description:** Delete plan and remove from MikroTik.

### Customer Management

#### POST `/v1/internal/customer`
**Description:** Create a new customer and provision on MikroTik.

#### GET `/v1/internal/customer/:id`
**Description:** Get customer profile and service status.

#### GET `/v1/internal/customer/list`
**Description:** List customers with pagination.

#### PUT `/v1/internal/customer/:id`
**Description:** Update customer info or plan.

#### DELETE `/v1/internal/customer/:id`
**Description:** Suspend or delete customer account.

#### GET `/v1/internal/customer/:id/ping`
**Description:** Trigger ICMP ping from RouterOS to customer IP.

### Monitoring

#### GET `/v1/internal/monitor/traffic/:interface` (WebSocket)
**Description:** Stream real-time PPS/BPS for a physical interface.

#### GET `/v1/internal/customer/:id/traffic/stream` (WebSocket)
**Description:** Stream real-time traffic specifically for a customer's dynamic interface.

---

## Callbacks
*Used by MikroTik scripts for event notifications.*

### POST `/v1/callbacks/pppoe/up`
**Description:** Notify application when a PPPoE session starts.

### POST `/v1/callbacks/pppoe/down`
**Description:** Notify application when a PPPoE session ends.

---

## V1 Client API

### GET `/v1/client/ping`
**Description:** Public health check for clients.
**Headers Required:** `X-Client-Key`