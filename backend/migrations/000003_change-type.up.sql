ALTER TABLE challenges DROP CONSTRAINT challenges_user_id_fkey;
ALTER TABLE challenges ALTER COLUMN user_id TYPE varchar (255);
