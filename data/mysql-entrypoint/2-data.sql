USE cubit_db;

INSERT INTO users (username, password, last_presence, last_presence_at, updated_at, created_at) VALUES
('admin',  'password', '', NOW(), NOW(), NOW()),
('room01', 'password', '', NOW(), NOW(), NOW());
