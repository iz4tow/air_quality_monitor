ALTER TABLE sensor_data
ADD COLUMN pm25 REAL DEFAULT 0;

ALTER TABLE sensor_data
ADD COLUMN pm10 REAL DEFAULT 0;
