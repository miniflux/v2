alter table entries add column document_vectors tsvector;
update entries set document_vectors = setweight(to_tsvector(coalesce(title, '')), 'A') || setweight(to_tsvector(coalesce(content, '')), 'B');
create index document_vectors_idx on entries using gin(document_vectors);