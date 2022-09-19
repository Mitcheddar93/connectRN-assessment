--This Transact-SQL query automatically sets the currently_active column for any new rows added to 1
--Research was done to see if the column can be made read-only after the new row was added, but time constraints and complexity prevented further implementation
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