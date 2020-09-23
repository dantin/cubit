USE cubit_db;


INSERT INTO roles (id, name) VALUES 
(1, 'admin'),
(2, 'user');

INSERT INTO users (username, password, last_presence, last_presence_at, updated_at, created_at) VALUES
('admin',  'password', '', NOW(), NOW(), NOW()),
('room01', 'password', '', NOW(), NOW(), NOW()),
('room02', 'password', '', NOW(), NOW(), NOW()),
('room03', 'password', '', NOW(), NOW(), NOW()),
('room04', 'password', '', NOW(), NOW(), NOW());

INSERT INTO user_role (username, role_id) VALUES
('admin',  1),
('room01', 2),
('room02', 2),
('room03', 2),
('room04', 2);
