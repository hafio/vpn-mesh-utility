-- Policy schedules (amendment A43): a policy may apply only on some days
-- of the week (days: mon..sun, comma separated; none is every day) and
-- between two times of day (HH:MM; none is the whole day). All three empty
-- is no schedule. The time zone they are read in is the setting policy_tz,
-- else server_tz, which the server writes at every start.
ALTER TABLE seg_policies ADD COLUMN days TEXT NOT NULL DEFAULT '';
ALTER TABLE seg_policies ADD COLUMN start_time TEXT NOT NULL DEFAULT '';
ALTER TABLE seg_policies ADD COLUMN end_time TEXT NOT NULL DEFAULT '';
