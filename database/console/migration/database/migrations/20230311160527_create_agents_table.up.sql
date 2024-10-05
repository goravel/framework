CREATE TABLE agents (
  id integer PRIMARY KEY AUTOINCREMENT NOT NULL,
  name varchar(255) NOT NULL,
  created_at datetime NOT NULL,
  updated_at datetime NOT NULL
);
INSERT INTO agents (name, created_at, updated_at) VALUES ('goravel', '2023-03-11 16:07:41', '2023-03-11 16:07:45');
