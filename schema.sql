CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE repositories (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  url         VARCHAR(256) NOT NULL,
  description VARCHAR(512) NOT NULL,
  updated     TIMESTAMP        DEFAULT now()
);
CREATE UNIQUE INDEX repositories_url_uindex
  ON public.repositories (url);

--

CREATE TABLE public.statistics (
  repository_id UUID,
  name          VARCHAR(64) NOT NULL,
  value         INT DEFAULT 0,
  url           VARCHAR     NOT NULL,
  CONSTRAINT statistics_repositories_id_fk FOREIGN KEY (repository_id) REFERENCES repositories (id) ON DELETE CASCADE
);
