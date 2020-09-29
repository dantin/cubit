USE cubit_db;


INSERT INTO roles (id, name) VALUES 
(1, 'admin'),
(2, 'user');

INSERT INTO users (username, password, last_presence, last_presence_at, updated_at, created_at) VALUES
('admin',  'password', '', NOW(), NOW(), NOW()),
('room01', 'password', '', NOW(), NOW(), NOW()),
('room02', 'password', '', NOW(), NOW(), NOW()),
('room03', 'password', '', NOW(), NOW(), NOW()),
('room04', 'password', '', NOW(), NOW(), NOW()),
('room05', 'password', '', NOW(), NOW(), NOW());

INSERT INTO user_role (username, role_id) VALUES
('admin',  1),
('room01', 2),
('room02', 2),
('room03', 2),
('room04', 2),
('room05', 2);

INSERT INTO rooms (id, username, name, `type`, created_at, updated_at) VALUES
(1, 'admin',  'Room QC', 'qc',     NOW(), NOW()),
(2, 'room01', 'Room 01', 'normal', NOW(), NOW()),
(3, 'room02', 'Room 02', 'normal', NOW(), NOW()),
(4, 'room03', 'Room 03', 'normal', NOW(), NOW()),
(5, 'room04', 'Room 04', 'normal', NOW(), NOW()),
(6, 'room05', 'Room 05', 'normal', NOW(), NOW());

INSERT INTO room_video_streams (id, input, broadcast, routes, `type`, room_id) VALUES
(1,  '', '', '{"room01": "srt://10.189.153.255:39991", "room02": "srt://10.189.153.255:39992", "room03": "srt://10.189.153.255:39993", "room04": "srt://10.189.153.255:39994", "room05": "srt://10.189.153.255:39995"}', 'camera', 1),
(2,  '', '', '{"user": "srt://10.189.153.255:65101", "admin": "srt://10.189.153.255:65102"}', 'device', 2),
(3,  '', '', '{"user": "srt://10.189.153.255:65103", "admin": "srt://10.189.153.255:65104"}', 'camera', 2),
(4,  '', '', '{"user": "srt://10.189.153.255:65106", "admin": "srt://10.189.153.255:65107"}', 'device', 3),
(5,  '', '', '{"user": "srt://10.189.153.255:65108", "admin": "srt://10.189.153.255:65109"}', 'camera', 3),
(6,  '', '', '{"user": "srt://10.189.153.255:65111", "admin": "srt://10.189.153.255:65112"}', 'device', 4),
(7,  '', '', '{"user": "srt://10.189.153.255:65113", "admin": "srt://10.189.153.255:65114"}', 'camera', 4),
(8,  '', '', '{"user": "srt://10.189.153.255:65116", "admin": "srt://10.189.153.255:65117"}', 'device', 5),
(9,  '', '', '{"user": "srt://10.189.153.255:65118", "admin": "srt://10.189.153.255:65119"}', 'camera', 5),
(10, '', '', '{"user": "srt://10.189.153.255:20021", "admin": "srt://10.189.153.255:30021"}', 'device', 6),
(11, '', '', '{"user": "srt://10.189.153.255:20020", "admin": "srt://10.189.153.255:30020"}', 'camera', 6);
