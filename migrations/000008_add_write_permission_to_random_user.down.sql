DELETE FROM users_permissions
WHERE user_id = (SELECT id FROM users WHERE email = 'Freya@example.com')
AND permission_id = (SELECT id FROM permissions WHERE code = 'university:write');
