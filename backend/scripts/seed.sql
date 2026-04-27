-- Seed data for development and testing
-- 1 verified account for US2+ test scenarios
-- Password: "TestPassword123!" (bcrypt cost 12)

INSERT OR IGNORE INTO accounts (
    id, email, email_canonical, password_hash, status,
    first_name, last_name, date_of_birth, created_at, updated_at
) VALUES (
    'seed-acc-001',
    'contact@komunumo.fr',
    'contact@komunumo.fr',
    '$2a$12$s7aSGQxOW4bgV45m7BW8Lun0vrV/V.7eSV9Ttf06L6.AXcPgWXcMi',
    'verified',
    'Anne',
    'Dupont',
    '1985-06-15',
    datetime('now', 'subsec'),
    datetime('now', 'subsec')
);
