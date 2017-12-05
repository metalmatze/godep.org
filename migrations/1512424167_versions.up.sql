CREATE TABLE versions (
  repository_id UUID        NOT NULL,
  name          VARCHAR(64) NOT NULL,
  sort_order    INT         NOT NULL,
  published     TIMESTAMP,
  CONSTRAINT versions_repositories_id_fk FOREIGN KEY (repository_id) REFERENCES repositories (id) ON DELETE CASCADE ON UPDATE CASCADE
);
CREATE UNIQUE INDEX versions_name_uindex
  ON versions (repository_id, name);
