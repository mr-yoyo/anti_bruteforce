CREATE TABLE ip_whitelist (
    id bigserial NOT NULL PRIMARY KEY,
    address inet NOT NULL,
    added_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX ON ip_whitelist USING GIST(address inet_ops);

CREATE TABLE ip_blacklist (
    id bigserial NOT NULL PRIMARY KEY,
    address inet NOT NULL,
    added_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX ON ip_blacklist USING GIST(address inet_ops);