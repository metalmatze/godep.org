CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE repositories (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  url         VARCHAR(256) NOT NULL,
  description VARCHAR(512) NOT NULL,
  updated     TIMESTAMP        DEFAULT now()
);
CREATE UNIQUE INDEX repositories_url_uindex
  ON repositories (url);
