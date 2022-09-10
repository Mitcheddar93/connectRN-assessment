CREATE TRIGGER no_longer_active
ON password_history
AFTER INSERT
AS
IF EXISTS (SELECT u.* FROM users AS u INNER JOIN password_history AS p ON u.user_id = p.user_id)
BEGIN
	SET NOCOUNT ON
	
  	UPDATE p
  	SET p.currently_active = 0
	FROM password_history AS p
  	LEFT JOIN INSERTED as i
	ON p.user_id = i.user_id
	WHERE p.change_date < i.change_date
END;