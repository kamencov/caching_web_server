-- +goose Up
-- +goose StatementBegin
create table users
(
    id            bigserial                 not null
        constraint users_pk
            primary key,
    login         text                      not null
        constraint login_unique
            unique,
    password_hash text                      not null,
    created_at    timestamptz default now() not null
);

create table documents
(
    id           uuid                      not null
        constraint documents_pk
            primary key,
    owner_id     bigint                    not null references users (id) on delete cascade,
    name         text                      not null,
    mime         text                      not null,
    hash_file    boolean                   not null,
    public       boolean     default false not null,
    json_data    bytea,
    storage_path text,
    create_at    timestamptz default now() not null,
    is_deleted   boolean     default false not null
);

Create index documents_owner_id_idx
    on documents (owner_id);

create table grants
(
    doc_id  uuid   not null references documents (id) on delete cascade,
    user_id bigint not null references users (id) on delete cascade,
    primary key (doc_id, user_id)
);
-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin
drop table users;
drop table documents;
drop index documents_owner_id_idx;
drop table grants;
-- +goose StatementEnd
