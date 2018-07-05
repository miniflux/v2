alter table entries add column document_vectors tsvector;
update entries set document_vectors = to_tsvector(title || ' ' || coalesce(content, ''));
create index document_vectors_idx on entries using gin(document_vectors);