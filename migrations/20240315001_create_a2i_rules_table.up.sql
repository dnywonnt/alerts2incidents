-- 20240315001_create_a2i_rules_table.up.sql
CREATE TABLE a2i_rules (
    id                                      VARCHAR(255) PRIMARY KEY,
    is_muted                                BOOLEAN NOT NULL,
    description                             TEXT NOT NULL,
    alerts_summary_conditions               VARCHAR(255)[] NOT NULL,
    alerts_activity_interval_conditions     INTERVAL[] NOT NULL,
    incident_life_time                      INTERVAL NOT NULL,
    set_incident_summary                    VARCHAR(255) NOT NULL,
    set_incident_description                TEXT NOT NULL,
    set_incident_departament                VARCHAR(255) NOT NULL,
    set_incident_client_affect              TEXT NOT NULL,
    set_incident_is_manageable              VARCHAR(255) NOT NULL,
    set_incident_sale_channels              VARCHAR(255)[] NOT NULL,
    set_incident_trouble_services           VARCHAR(255)[] NOT NULL,
    set_incident_failure_type               VARCHAR(255) NOT NULL,
    set_incident_is_downtime                BOOLEAN NOT NULL,
    set_incident_labels                     VARCHAR(255)[],
    created_at                              TIMESTAMP NOT NULL,
    updated_at                              TIMESTAMP NOT NULL
);

CREATE OR REPLACE FUNCTION notify_a2i_rules_event()
RETURNS TRIGGER AS $$
BEGIN
    IF (TG_OP = 'INSERT') THEN
        PERFORM pg_notify('a2i_rules_channel', 'INSERT:' || NEW.id);
    ELSIF (TG_OP = 'UPDATE') THEN
        PERFORM pg_notify('a2i_rules_channel', 'UPDATE:' || NEW.id);
    ELSIF (TG_OP = 'DELETE') THEN
        PERFORM pg_notify('a2i_rules_channel', 'DELETE:' || OLD.id);
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER a2i_rules_event_trigger
AFTER INSERT OR UPDATE OR DELETE ON a2i_rules
FOR EACH ROW EXECUTE FUNCTION notify_a2i_rules_event();
