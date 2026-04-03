INSERT INTO users_permissions
VALUES (
    (SELECT id FROM users WHERE email = 'Freya@example.com'),
    (SELECT id FROM permissions WHERE  code = 'university:write')
);
