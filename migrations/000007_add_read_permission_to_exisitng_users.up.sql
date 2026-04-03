INSERT INTO users_permissions
SELECT id, (SELECT id FROM permissions WHERE code = 'university:read') 
FROM users;
