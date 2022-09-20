--Example of how a password can be safely inserted into the password_history table
--The hash provided is the password provided that is salted using the creation date formatted as a unix timestamp
--Note that the currently_active column is not set due to the trigger that automatically sets the password to active
--A query that sets the other rows for the user as inactive is unnecessary as that is being performed by another trigger
INSERT INTO password_history (user_id, password, change_date)
VALUES (1, HASHBYTES('SHA2_256', 'supercalifragilisticexpialidocious' + CAST(DATEDIFF(SECOND, '1970-01-01', GETDATE()) AS VARCHAR)), GETDATE())
