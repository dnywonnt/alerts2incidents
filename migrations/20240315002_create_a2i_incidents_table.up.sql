-- 20240315002_create_a2i_incidents_table.up.sql
CREATE TABLE a2i_incidents (
    id                  VARCHAR(255) PRIMARY KEY,
    type                VARCHAR(255) NOT NULL,
    status              VARCHAR(255) NOT NULL,
    summary             VARCHAR(255) NOT NULL,
    description         TEXT NOT NULL,
    from_at             TIMESTAMP NOT NULL,
    to_at               TIMESTAMP NOT NULL,
    is_confirmed        BOOLEAN NOT NULL,
    confirmation_time   TIMESTAMP NOT NULL,
    quarter             INTEGER NOT NULL,
    departament         VARCHAR(255) NOT NULL,
    client_affect       TEXT NOT NULL,
    is_manageable       VARCHAR(255) NOT NULL,
    sale_channels       VARCHAR(255)[] NOT NULL,
    trouble_services    VARCHAR(255)[] NOT NULL,
    fin_losses          INTEGER NOT NULL,
    failure_type        VARCHAR(255) NOT NULL,
    is_deploy           BOOLEAN NOT NULL,
    deploy_link         VARCHAR(255) NOT NULL,
    labels              VARCHAR(255)[] NOT NULL,
    is_downtime         BOOLEAN NOT NULL NOT NULL,
    postmortem_link     VARCHAR(255) NOT NULL,
    creator             VARCHAR(255) NOT NULL,
    rule_id             VARCHAR(255),
    matching_count      INTEGER NOT NULL,
    last_matching_time  TIMESTAMP NOT NULL,
    alerts_data         JSONB NOT NULL,
    created_at          TIMESTAMP NOT NULL,
    updated_at          TIMESTAMP NOT NULL,
    CONSTRAINT fk_rule FOREIGN KEY(rule_id) REFERENCES a2i_rules(id) ON DELETE SET NULL
);

CREATE OR REPLACE FUNCTION notify_a2i_incidents_event()
RETURNS TRIGGER AS $$
BEGIN
    IF (TG_OP = 'INSERT') THEN
        PERFORM pg_notify('a2i_incidents_channel', 'INSERT:' || NEW.id);
    ELSIF (TG_OP = 'UPDATE') THEN
        PERFORM pg_notify('a2i_incidents_channel', 'UPDATE:' || NEW.id);
    ELSIF (TG_OP = 'DELETE') THEN
        PERFORM pg_notify('a2i_incidents_channel', 'DELETE:' || OLD.id);
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER a2i_incidents_event_trigger
AFTER INSERT OR UPDATE OR DELETE ON a2i_incidents
FOR EACH ROW EXECUTE FUNCTION notify_a2i_incidents_event();