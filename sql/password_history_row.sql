CREATE TRIGGER currently_active
ON password_history
AFTER INSERT
AS
BEGIN
	SET NOCOUNT ON
	
  	UPDATE password_history
  	SET currently_active = 1
  	FROM INSERTED
  	WHERE INSERTED.user_id = password_history.user_id
END;

INSERT INTO password_history (user_id, password, change_date)
VALUES (1, 'test1', GETDATE());