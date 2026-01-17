# MikroOps API Documentation

Comprehensive API documentation for MikroOps Multi-Tenant Billing & Monitoring System.

## Table of Contents
1. [Base Configuration](#base-configuration)
2. [Authentication](#1-authentication)
3. [Tenant Management](#2-tenant-management)
4. [Client Operations](#3-client-operations)
5. [MikroTik Device Management](#4-mikrotik-device-management)
6. [PPP Management](#5-ppp-management)
7. [Billing Profiles](#6-billing-profiles)
8. [Customer Management](#7-customer-management)
9. [Monitoring & Real-time Stats](#8-monitoring--real-time-stats)
10. [Callbacks & System](#9-callbacks--system)

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
*   `Authorization`: `Bearer <jwt_token>` (Required for all `/internal/*` routes)
*   `Content-Type`: `application/json`
*   `X-Tenant-ID`: `<uuid>`
    *   **Required** for **Super Admins** to switch context to a specific tenant.
    *   Derived automatically for regular Users.

---

## 1. Authentication
Public endpoints for identity management.

### Login
*   **Path:** `POST /auth/login`
*   **Request Body:**
    ```json
    {
      "email": "admin@example.com", // Optional
      "username": "admin",         // Required if email is empty
      "password": "password123"
    }
    ```
*   **Response (Data):**
    ```json
    {
      "id": "uuid",
      "username": "admin",
      "email": "admin@example.com",
      "fullname": "Administrator",
      "user_role": "admin",
      "role_id": "uuid",
      "api_token": "jwt_string"
    }
    ```

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
*   **Response (Data):** Full `model.TenantResponse` including limits and features.

---

## 3. Client Operations
Generic operations for managing linked clients/entities.

### Upsert Clients
*   **Path:** `POST /internal/client-upsert`
*   **Request Body:** `[]model.ClientInput`
*   **Response (Data):** List of upserted clients.

### Find Clients
*   **Path:** `POST /internal/client-find`
*   **Request Body:** `model.ClientFilter`
*   **Response (Data):** List of matching clients.

---

## 4. MikroTik Device Management
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
*   **Note:** Uses POST for potential complex filtering in the future, currently returns all.
*   **Response (Data):** `[]model.MikrotikResponse`

### Get Device By ID
*   **Path:** `GET /internal/mikrotik/:id`
*   **Response (Data):**
    ```json
    {
      "id": "uuid",
      "name": "Router-01",
      "host": "103.11.22.33",
      "status": "online",
      "total_profiles": 5,
      "total_customers": 120,
      "last_sync": "2024-01-01T12:00:00Z"
    }
    ```

### Set Active Device
*   **Path:** `PATCH /internal/mikrotik/:id/activate`
*   **Description:** Sets this router as the primary active device for the tenant.

---

## 5. PPP Management
Direct MikroTik API calls for PPP Secrets and Profiles.

### Create PPP Secret
*   **Path:** `POST /internal/ppp/secret`
*   **Request Body (`model.PPPSecretInput`):**
    ```json
    {
      "name": "customer_user",
      "password": "secretpassword",
      "profile": "Plan-10M",
      "service": "pppoe",
      "comment": "John Doe"
    }
    ```

### List PPP Secrets
*   **Path:** `GET /internal/ppp/secret/list`
*   **Response (Data):** `[]model.PPPSecret`

### List PPP Profiles
*   **Path:** `GET /internal/ppp/profile/list`
*   **Response (Data):** `[]model.PPPProfile`

---

## 6. Billing Profiles
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
*   **Response (Data):** `[]model.ProfileResponse`

---

## 7. Customer Management
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
      "profile_id": "uuid",
      "service_type": "pppoe",
      "billing_day": 5
    }
    ```

### Get Customer
*   **Path:** `GET /internal/customer/:id`
*   **Response (Data):** Full `model.CustomerResponse` including `mikrotik` info and `services` history.

---

## 8. Monitoring & Real-time Stats
WebSocket and high-frequency polling routes.

### Real-time Traffic Stream
*   **Path:** `GET /internal/customer/:id/traffic/stream`
*   **Protocol:** WebSocket
*   **Output (`model.CustomerTrafficData`):**
    ```json
    {
      "rx_bits_per_second": "5240000",
      "tx_bits_per_second": "1200000",
      "download_speed": "5.24 Mbps",
      "upload_speed": "1.20 Mbps",
      "timestamp": "..."
    }
    ```

### On-Demand Ping
*   **Path:** `GET /internal/customer/:id/ping`
*   **Response (Data):** Returns a one-time ping result summary.

### Ping Stream
*   **Path:** `GET /internal/customer/:id/ping/stream`
*   **Protocol:** WebSocket
*   **Output:** Continuous `model.PingResponse` frames until closure.

---

## 9. Callbacks & System

### PPPoE On-Up Callback
*   **Path:** `POST /callbacks/pppoe/up`
*   **Request Body (`model.PPPoEEventInput`):**
    ```json
    {
      "name": "johndoe",
      "interface": "pppoe-johndoe",
      "remote_address": "192.168.100.50"
    }
    ```

### System Resource Ping
*   **Path:** `GET /client/ping`
*   **Output:** Returns Backend CPU, RAM, and Core usage stats.

---
*Generated by tracing: Handler -> Port -> Domain -> Model.*