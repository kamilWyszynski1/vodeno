CREATE TABLE entry (
    id SERIAL PRIMARY KEY,
    email TEXT NOT NULL,
    title TEXT NOT NULL,
    content TEXT NOT NULL,
    mailing_id NUMERIC NOT NULL,
    insert_time timestamp with time zone NOT NULL
);

-- listing will be faster :)
CREATE INDEX entry_id ON entry(id);

CREATE UNIQUE INDEX entry_payload ON entry(title, content);