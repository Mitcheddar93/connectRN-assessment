SELECT u.first_name, u.last_name, p.password
FROM users AS u
INNER JOIN password_history AS p
ON u.user_id = p.user_id AND p.currently_active = 1;