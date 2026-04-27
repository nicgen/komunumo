-- Seed data for development and testing
-- 1 verified account for US2+ test scenarios
-- Password: "TestPassword123!" (bcrypt cost 12)

INSERT OR IGNORE INTO accounts (
    id, email, email_canonical, password_hash, status,
    first_name, last_name, date_of_birth, created_at, updated_at
) VALUES (
    'seed-acc-001',
    'anne@example.com',
    'anne@example.com',
    '$2a$12$LQv3c1yqBwEHxPvonEe9XOIr9H1.H3QZY4kW.8.ZhQj2RFxMDpXiS',
    'verified',
    'Anne',
    'Dupont',
    '1985-06-15',
    datetime('now', 'subsec'),
    datetime('now', 'subsec')
);
