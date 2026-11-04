-- 1. Ubah aturan idempotency: dari (outbox_id) menjadi (outbox_id, user_id)
-- Karena 1 event ubah kuota akan menghasilkan BANYAK email (ke tiap mahasiswa).
ALTER TABLE notifications DROP CONSTRAINT IF EXISTS notifications_outbox_unique;
ALTER TABLE notifications ADD CONSTRAINT notifications_outbox_user_unique UNIQUE (outbox_id, user_id);

-- 2. Buat tabel enrollments sederhana (karena kita build from scratch)
CREATE TABLE IF NOT EXISTS enrollments (
    id BIGSERIAL PRIMARY KEY,
    student_id BIGINT NOT NULL,
    course_id BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 3. Masukkan 5 mahasiswa dummy ke kelas ID 1 (buat ngetes nanti)
INSERT INTO enrollments (student_id, course_id) VALUES 
(1, 1), (2, 1), (3, 1), (4, 1), (5, 1)
ON CONFLICT DO NOTHING;