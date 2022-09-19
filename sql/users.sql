--A user_id column was added that creates unique integers automatically to correspond with each user
--It is possible that people with the same first and last name living in the same city and ZIP code can exist, so no other columns were made pimrary keys
CREATE TABLE users (
  	user_id INT PRIMARY KEY IDENTITY,
  	first_name NVARCHAR(100) NOT NULL,
  	last_name NVARCHAR(100) NOT NULL,
  	city NVARCHAR(100) NOT NULL,
  	zip_code INT NOT NULL
);