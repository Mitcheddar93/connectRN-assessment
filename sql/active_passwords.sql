--The password and currently_active columns are present in the password_history table, so an inner join is used with user_id as the common column 
SELECT u.first_name, u.last_name, p.password
FROM users AS u
INNER JOIN password_history AS p
ON u.user_id = p.user_id AND p.currently_active = 1;