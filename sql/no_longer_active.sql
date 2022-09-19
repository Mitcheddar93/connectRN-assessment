--This Transact-SQL query creates a trigger that automatically updates rows with older change_date values as inactive for a user_id in the password_history table when a new row is added for a new password
--A left join is used for the INSERTED row so that all rows except for the new one will be updated with currently_active = 0
--A trigger for setting a new row to the current date for the change_date column, but all of the rows were set to the same date in testing (used dbfiddle.uk as testing tool)
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