alter table feeds 
    add column blocklist_rules text not null default '',
    add column keeplist_rules text not null default ''
;
