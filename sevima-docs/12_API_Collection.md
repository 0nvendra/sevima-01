# API Collection & Specification
# SEVIMA-PAY: RESTful API Endpoints

> **Version**: 1.0.0  
> **Base URL**: `http://localhost:8080/api/v1`  
> **Format**: JSON (`Content-Type: application/json`)  
> **Authentication**: Bearer Token (`Authorization: Bearer <access_token>`)  
> **Downloadable Collection**: [SEVIMA-PAY.postman_collection.json](file:///d:/Project/sevima/sevima-docs/SEVIMA-PAY.postman_collection.json) (Siap diimpor langsung ke Postman / Thunder Client)  

---

## 1. Overview & Standard Responses

### 1.1 Success Response Envelope
```json
{
  "success": true,
  "message": "Operasi berhasil",
  "data": { ... },
  "meta": {
    "page": 1,
    "per_page": 10,
    "total_records": 100,
    "total_pages": 10
  }
}
```

### 1.2 Error Response Envelope
```json
{
  "success": false,
  "error": {
    "code": "INVALID_SIGNATURE",
    "message": "Verifikasi HMAC-SHA256 signature gagal",
    "details": null
  }
}
```

---

## 2. Authentication Endpoints

### 2.1 User Login
- **Endpoint**: `POST /auth/login`
- **Access**: Public
- **Request Body**:
```json
{
  "email": "admin@ugm.ac.id",
  "password": "Password123!"
}
```
- **Response `200 OK`**:
```json
{
  "success": true,
  "message": "Login berhasil",
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsIn...",
    "refresh_token": "7a9b8c0d-1e2f-3a4b...",
    "expires_in": 900,
    "user": {
      "id": "e4b2d1c0-0000-0000-0000-000000000001",
      "email": "admin@ugm.ac.id",
      "full_name": "Budi Santoso",
      "role": "admin_keuangan",
      "institusi_id": "c3a1b0b0-1111-2222-3333-444455556666",
      "institution_code": "UGM",
      "institution_name": "Universitas Gadjah Mada"
    }
  }
}
```
- **cURL Example**:
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@ugm.ac.id","password":"Password123!"}'
```

### 2.2 Refresh Token
- **Endpoint**: `POST /auth/refresh`
- **Access**: Public
- **Request Body**:
```json
{
  "refresh_token": "7a9b8c0d-1e2f-3a4b..."
}
```
- **Response `200 OK`**:
```json
{
  "success": true,
  "message": "Token berhasil diperbarui",
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsIn...",
    "refresh_token": " नए-uuid-refresh-token...",
    "expires_in": 900
  }
}
```

### 2.3 User Logout
- **Endpoint**: `POST /auth/logout`
- **Access**: Authenticated
- **Headers**: `Authorization: Bearer <token>`
- **Request Body**:
```json
{
  "refresh_token": "7a9b8c0d-1e2f-3a4b..."
}
```
- **Response `200 OK`**:
```json
{
  "success": true,
  "message": "Logout berhasil"
}
```

---

## 3. Institution Endpoints (Super Admin Only)

### 3.1 List All Institutions
- **Endpoint**: `GET /institutions`
- **Access**: `super_admin`
- **Response `200 OK`**:
```json
{
  "success": true,
  "data": [
    {
      "id": "c3a1b0b0-1111-2222-3333-444455556666",
      "code": "UGM",
      "name": "Universitas Gadjah Mada",
      "email": "info@ugm.ac.id",
      "phone": "0274-588688",
      "is_active": true,
      "created_at": "2026-01-01T08:00:00Z"
    }
  ]
}
```

### 3.2 Create Institution
- **Endpoint**: `POST /institutions`
- **Access**: `super_admin`
- **Request Body**:
```json
{
  "code": "ITS",
  "name": "Institut Teknologi Sepuluh Nopember",
  "email": "humas@its.ac.id",
  "phone": "031-5994251",
  "address": "Kampus ITS Sukolilo, Surabaya"
}
```
- **Response `201 Created`**:
```json
{
  "success": true,
  "message": "Institusi berhasil ditambahkan",
  "data": {
    "id": "d5f6a7b8-9999-8888-7777-666655554444",
    "code": "ITS",
    "name": "Institut Teknologi Sepuluh Nopember",
    "hmac_secret": "e2a9b3c4d5e6f708192a3b4c5d6e7f80..."
  }
}
```

---

## 4. Invoice Endpoints (Admin Keuangan & Mahasiswa)

### 4.1 List Invoices
- **Endpoint**: `GET /invoices`
- **Access**: `admin_keuangan` (semua invoice di institusinya), `mahasiswa` (hanya invoice miliknya)
- **Query Params**:
  - `status`: `unpaid` | `paid` | `expired`
  - `semester`: `ganjil` | `genap`
  - `academic_year`: `2026/2027`
  - `page`: `1`
  - `per_page`: `10`
- **Response `200 OK`**:
```json
{
  "success": true,
  "data": [
    {
      "id": "b1a2c3d4-4444-5555-6666-777788889999",
      "invoice_number": "INV-2026-UGM-0001",
      "student": {
        "id": "a1b2c3d4-...",
        "nim": "21/475123/TK/52111",
        "full_name": "Ahmad Fauzi",
        "study_program": "Teknik Informatika"
      },
      "amount": 7500000,
      "status": "unpaid",
      "semester": "ganjil",
      "academic_year": "2026/2027",
      "due_date": "2026-11-15T23:59:59Z",
      "created_at": "2026-10-01T09:00:00Z"
    }
  ],
  "meta": {
    "page": 1,
    "per_page": 10,
    "total_records": 1,
    "total_pages": 1
  }
}
```

### 4.2 Create Invoice
- **Endpoint**: `POST /invoices`
- **Access**: `admin_keuangan`
- **Request Body**:
```json
{
  "student_id": "a1b2c3d4-1111-2222-3333-444455556666",
  "amount": 6000000,
  "semester": "ganjil",
  "academic_year": "2026/2027",
  "due_date": "2026-12-01T23:59:59Z",
  "notes": "Tagihan UKT Golongan 4"
}
```
- **Response `201 Created`**:
```json
{
  "success": true,
  "message": "Tagihan berhasil dibuat",
  "data": {
    "id": "b2c3d4e5-6666-7777-8888-999900001111",
    "invoice_number": "INV-2026-UGM-0002",
    "amount": 6000000,
    "status": "unpaid",
    "due_date": "2026-12-01T23:59:59Z"
  }
}
```

### 4.3 Get Student KRS Eligibility Status
- **Endpoint**: `GET /students/krs-status`
- **Access**: `mahasiswa`
- **Response `200 OK`**:
```json
{
  "success": true,
  "data": {
    "nim": "21/475123/TK/52111",
    "full_name": "Ahmad Fauzi",
    "is_eligible_krs": true,
    "unpaid_invoices_count": 0,
    "active_semester": "ganjil",
    "active_academic_year": "2026/2027",
    "message": "Seluruh kewajiban pembayaran telah lunas. Anda berhak mengisi KRS."
  }
}
```

---

## 5. Webhook Notification Endpoint

### 5.1 Process Payment Webhook
- **Endpoint**: `POST /webhook/payment`
- **Access**: Public / Payment Gateway Caller
- **Headers**:
  - `Content-Type`: `application/json`
  - `X-Signature`: `<HMAC-SHA256 hex digest of raw request body>`
- **Request Body**:
```json
{
  "institution_code": "UGM",
  "event_type": "payment.success",
  "invoice_number": "INV-2026-UGM-0001",
  "amount": 7500000,
  "payment_method": "BCA_VA",
  "paid_at": "2026-10-30T10:00:00Z",
  "transaction_id": "TRX-PG-992817293"
}
```
- **Response `200 OK` (Processed)**:
```json
{
  "success": true,
  "message": "Webhook processed successfully",
  "data": {
    "invoice_number": "INV-2026-UGM-0001",
    "status": "paid"
  }
}
```
- **Response `200 OK` (Idempotent Duplicate)**:
```json
{
  "success": true,
  "message": "Event already processed (idempotent no-op)",
  "data": {
    "invoice_number": "INV-2026-UGM-0001",
    "status": "paid"
  }
}
```
- **Response `401 Unauthorized` (Invalid Signature)**:
```json
{
  "success": false,
  "error": {
    "code": "INVALID_SIGNATURE",
    "message": "HMAC signature verification failed"
  }
}
```

---

## 6. Mock Payment Gateway Endpoints

### 6.1 Trigger Mock Payment (Automated HMAC generation)
- **Endpoint**: `POST /mock-gateway/pay`
- **Access**: Public / Development & Evaluation Tool
- **Request Body**:
```json
{
  "institution_code": "UGM",
  "invoice_number": "INV-2026-UGM-0001",
  "payment_method": "MANDIRI_VA",
  "simulate_tamper": false
}
```
- **Behavior**:
  1. Backend mencari secret key kampus `UGM`.
  2. Menyusun payload webhook standar `payment.success`.
  3. Menghitung HMAC-SHA256 signature yang valid (atau sengaja dirusak jika `simulate_tamper: true`).
  4. Melakukan internal dispatch ke handler webhook.
  5. Mengembalikan hasil eksekusi lengkap dengan signature yang dikalkulasikan.
- **Response `200 OK`**:
```json
{
  "success": true,
  "message": "Mock payment triggered and verified",
  "data": {
    "generated_signature": "4f53cda18c2bee0d...",
    "webhook_response_code": 200,
    "invoice_status_after": "paid"
  }
}
```

---

## 7. HMAC-SHA256 Signature Generation Guide

Bagi penguji atau integrator eksternal yang ingin menembakkan webhook manual via cURL, berikut formula kalkulasi signature:

### Python 3
```python
import hmac, hashlib

secret = b"your_institution_secret"
payload = b'{"institution_code":"UGM","event_type":"payment.success","invoice_number":"INV-2026-UGM-0001","amount":7500000,"payment_method":"BCA_VA","paid_at":"2026-10-30T10:00:00Z","transaction_id":"TRX-123"}'

signature = hmac.new(secret, payload, hashlib.sha256).hexdigest()
print("X-Signature:", signature)
```

### Node.js
```javascript
const crypto = require('crypto');

const secret = 'your_institution_secret';
const payload = JSON.stringify({ ... });

const signature = crypto.createHmac('sha256', secret).update(payload).digest('hex');
console.log('X-Signature:', signature);
```
