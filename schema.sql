CREATE TABLE IF NOT EXISTS links (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    url text NOT NULL,
    slug text UNIQUE,
    created_at timestamptz NOT NULL DEFAULT now()
);