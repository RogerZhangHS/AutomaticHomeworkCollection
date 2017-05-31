CREATE TABLE `Students` (
	`id`	integer PRIMARY KEY AUTOINCREMENT,
	`Name`	text NOT NULL,
	`StuId`	text NOT NULL UNIQUE,
	`Submitted`	INTEGER DEFAULT 0
);

CREATE TABLE `SubRecords` (
	`id`	integer PRIMARY KEY AUTOINCREMENT,
	`StuId`	INTEGER NOT NULL,
	`SubTime`	datetime DEFAULT (datetime('now'))
);

CREATE TABLE `sqlite_sequence` (
	`name`	TEXT,
	`seq`	TEXT
);