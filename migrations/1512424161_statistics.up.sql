CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE statistics (
  repository_id UUID,
  name          VARCHAR(64) NOT NULL,
  value         INT DEFAULT 0,
  url           VARCHAR     NOT NULL,
  CONSTRAINT statistics_repositories_id_fk FOREIGN KEY (repository_id) REFERENCES repositories (id) ON DELETE CASCADE ON UPDATE CASCADE
);
