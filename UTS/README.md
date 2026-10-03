# SIAKAD Mini — RESTful API

Backend API layanan akademik sederhana. Dibangun menggunakan Go, Fiber, pgx, dan PostgreSQL.

## Teknologi

| Komponen          | Teknologi      |
|-------------------|----------------|
| Bahasa            | Go 1.21+       |
| Framework         | Fiber v2       |
| Database Driver   | pgx v5         |
| Database          | PostgreSQL 15+ |
| Authentication    | JWT            |
| Password Hashing  | bcrypt         |

## Struktur Folder

```
UTS/
├── app/
│   ├── model/         # struct entity dan request/response
│   ├── repository/    # query SQL via pgx
│   └── service/       # business rules dan handler endpoint
├── config/            # pembacaan .env
├── database/          # koneksi PostgreSQL dan migration
├── helper/            # JWT, hashing password, response formatter
├── middleware/        # auth, role, logger, CORS
├── migrations/        # file SQL
├── route/             # pendaftaran route
├── seeder/            # data awal
└── main.go
```

## Setup

1. Siapkan PostgreSQL:

   ```sql
   CREATE USER siakad WITH PASSWORD 'siakad123';
   CREATE DATABASE siakad_mini_fiber OWNER siakad;
   ```

2. Copy `.env.example` menjadi `.env`, sesuaikan nilainya.

3. Install dependency dan jalankan:

   ```bash
   go mod tidy
   go run .
   ```

   Migrasi dan seeder akan otomatis dijalankan saat start.

## Akun Seed

| Role      | Email                                      | Password          |
|-----------|--------------------------------------------|-------------------|
| Admi      | `admin@siakad.test`                        | `admin12345`      |
| Mahasiswa | `mhs1@siakad.test` ... `mhs20@siakad.test` | NIM masing-masing |

## Daftar Endpoint

| No    | Method | Endpoint                   | Akses                           |
|---    |---     |---                         |---                              |
| 1     | POST   | `/api/v1/auth/login`       | Publik                          |
| 2     | GET    | `/api/v1/auth/me`          | Semua role                      |
| 3     | GET    | `/api/v1/students`         | Admin                           |
| 4     | POST   | `/api/v1/students`         | Admin                           |
| 5     | GET    | `/api/v1/students/{id}`    | Admin, mahasiswa (data sendiri) |
| 6     | PUT    | `/api/v1/students/{id}`    | Admin                           |
| 7     | DELETE | `/api/v1/students/{id}`    | Admin                           |
| 8     | GET    | `/api/v1/courses`          | Semua role                      |
| 9     | POST   | `/api/v1/enrollments`      | Mahasiswa                       |
| 10    | DELETE | `/api/v1/enrollments/{id}` | Mahasiswa                       |

## Aturan Bisnis

| Aturan            | Keterangan                                                                      |
|-------------------|---------------------------------------------------------------------------------|
| Batas SKS         | IPK ≥ 3.00 → maks 24 SKS; IPK 2.50–2.99 → maks 21 SKS; IPK < 2.50 → maks 18 SKS |
| Duplikasi MK      | Tidak boleh mengambil mata kuliah yang sama pada tahun akademik yang sama       |
| Kuota             | Mata kuliah dengan kuota penuh tidak dapat diambil                              |
| Kepemilikan KRS   | Mahasiswa hanya dapat mengakses KRS miliknya sendiri                            |

## Format Response

```json
{
  "success": true,
  "message": "...",
  "data": { }
}
```

## Author

Octavia Nuzulul Azmi — 434241040
D4 Teknik Informatika, Fakultas Vokasi, Universitas Airlangga