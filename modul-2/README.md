REST API Students - Modul 2

Proyek ini adalah implementasi REST API untuk entitas Student menggunakan framework Go Fiber. Dibuat untuk memenuhi tugas praktikum Backend Lanjut Modul 2.

Kontrak API

Berikut adalah daftar endpoint beserta detail request dan response yang tersedia:

| Metode | Endpoint | Parameter | Contoh Body Permintaan | Status yang Mungkin | Contoh Respons |
|---|---|---|---|---|---|
| GET | /api/v1/students | Query: page, limit, search, sort, order, is_active | Kosong | 200 OK | {"success": true, "message": "daftar mahasiswa berhasil diambil", "data": [{"id": 2, "nim": "999888777", "name": "Budi Santoso", "grade": 88, "is_active": true}], "meta": {"page": 1, "limit": 5, "total": 1, "total_pages": 1}} |
| GET | /api/v1/students/:id | Path: id | Kosong | 200 OK, 400 Bad Request, 404 Not Found | {"success": true, "message": "data mahasiswa ditemukan", "data": {"id": 1, "nim": "434241040", "name": "Octavia", "grade": 90.5, "is_active": true}} |
| POST | /api/v1/students | Kosong | {"nim": "434241040", "name": "Octavia", "grade": 90.5} | 201 Created, 400 Bad Request, 409 Conflict, 415 Unsupported Media Type, 422 Unprocessable Entity | {"success": true, "message": "data mahasiswa berhasil ditambahkan", "data": {"id": 1, "nim": "434241040", "name": "Octavia", "grade": 90.5, "is_active": true}} |
| PUT | /api/v1/students/:id | Path: id | {"nim": "434241040", "name": "Octavia Nuzulul", "grade": 98.5, "is_active": false} | 200 OK, 400 Bad Request, 404 Not Found, 409 Conflict, 415 Unsupported Media Type, 422 Unprocessable Entity | {"success": true, "message": "data mahasiswa berhasil diganti seluruhnya", "data": {"id": 1, "nim": "434241040", "name": "Octavia Nuzulul", "grade": 98.5, "is_active": false}} |
| PATCH | /api/v1/students/:id | Path: id | {"is_active": true} | 200 OK, 400 Bad Request, 404 Not Found, 409 Conflict, 415 Unsupported Media Type, 422 Unprocessable Entity | {"success": true, "message": "data mahasiswa berhasil diperbarui sebagian", "data": {"id": 1, "nim": "434241040", "name": "Octavia Nuzulul", "grade": 98.5, "is_active": true}} |
| DELETE | /api/v1/students/:id | Path: id | Kosong | 204 No Content, 400 Bad Request, 404 Not Found | Kosong |