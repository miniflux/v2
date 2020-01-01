ALTER TABLE users ALTER COLUMN theme SET DEFAULT 'light_serif';
UPDATE users SET theme='light_serif' WHERE theme='default';
UPDATE users SET theme='light_sans_serif' WHERE theme='sansserif';
UPDATE users SET theme='dark_serif' WHERE theme='black';
