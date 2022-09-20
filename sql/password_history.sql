--The DELETE CASCADE command on the foreign key on the user_id column ensures that all rows with the connected user_id are deleted when the users row with the same user_id is deleted
--Unsure if a similar command exists in MySQL
CREATE TABLE password_history (
  	user_id INT FOREIGN KEY REFERENCES users(user_id) ON DELETE CASCADE,
  	password VARBINARY(40) NOT NULL,
  	change_date DATETIME NOT NULL,
  	currently_active BIT
);