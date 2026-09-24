INSERT INTO permissions (name, description) VALUES
    ('student:list', 'Melihat daftar seluruh student'),
    ('student:read:any', 'Melihat data student mana pun'),
    ('student:create', 'Membuat data student baru'),
    ('student:update:any', 'Mengubah data student mana pun'),
    ('student:delete', 'Menghapus data student')
ON CONFLICT (name) DO NOTHING;

INSERT INTO role_permissions (role_name, permission_name) VALUES
    ('admin', 'student:list'),
    ('admin', 'student:read:any'),
    ('admin', 'student:create'),
    ('admin', 'student:update:any'),
    ('admin', 'student:delete'),
    ('staff', 'student:list'),
    ('staff', 'student:read:any'),
    ('staff', 'student:create')
ON CONFLICT DO NOTHING;