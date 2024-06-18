-- 20240315002_create_a2i_incidents_table.down.sql
DROP TRIGGER IF EXISTS a2i_incidents_event_trigger ON a2i_incidents;
DROP FUNCTION IF EXISTS notify_a2i_incidents_event();
DROP TABLE IF EXISTS a2i_incidents;
