DROP TABLE versions;
DROP TABLE statistics;
DROP TABLE repositories;

--

TRUNCATE repositories CASCADE;

--

CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE repositories (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  url         VARCHAR(256) NOT NULL,
  description VARCHAR(512) NOT NULL,
  updated     TIMESTAMP        DEFAULT now()
);
CREATE UNIQUE INDEX repositories_url_uindex
  ON repositories (url);

CREATE TABLE statistics (
  repository_id UUID,
  name          VARCHAR(64) NOT NULL,
  value         INT DEFAULT 0,
  url           VARCHAR     NOT NULL,
  CONSTRAINT statistics_repositories_id_fk FOREIGN KEY (repository_id) REFERENCES repositories (id) ON DELETE CASCADE ON UPDATE CASCADE
);

CREATE TABLE versions (
  repository_id UUID        NOT NULL,
  name          VARCHAR(64) NOT NULL,
  sort_order    INT         NOT NULL,
  published     TIMESTAMP,
  CONSTRAINT versions_repositories_id_fk FOREIGN KEY (repository_id) REFERENCES repositories (id) ON DELETE CASCADE ON UPDATE CASCADE
);
CREATE UNIQUE INDEX versions_name_uindex
  ON versions (repository_id, name);
