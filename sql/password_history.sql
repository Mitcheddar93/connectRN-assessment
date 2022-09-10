CREATE TABLE password_history (
  	user_id INT FOREIGN KEY REFERENCES users(user_id) ON DELETE CASCADE,
  	password NVARCHAR(MAX) NOT NULL,
  	change_date DATETIME NOT NULL,
  	currently_active BIT
);