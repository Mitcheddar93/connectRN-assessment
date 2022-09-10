CREATE TABLE users (
  	user_id INT PRIMARY KEY IDENTITY,
  	first_name NVARCHAR(100) NOT NULL,
  	last_name NVARCHAR(100) NOT NULL,
  	city NVARCHAR(100) NOT NULL,
  	zip_code INT NOT NULL
);