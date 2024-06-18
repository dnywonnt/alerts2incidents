-- 20240315001_create_a2i_rules_table.down.sql
DROP TRIGGER IF EXISTS a2i_rules_event_trigger ON a2i_rules;
DROP FUNCTION IF EXISTS notify_a2i_rules_event();
DROP TABLE IF EXISTS a2i_rules;
