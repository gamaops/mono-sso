CREATE SCHEMA IF NOT EXISTS sso AUTHORIZATION CURRENT_USER;

CREATE TABLE IF NOT EXISTS sso.account(
	id char(28) not null,
	name text not null,
	activation_method smallint not null default 0,
	password char(60),
	created_at timestamp with time zone not null,
	updated_at timestamp with time zone not null,
	deleted_at timestamp with time zone
) WITHOUT OIDS;

CREATE TABLE IF NOT EXISTS sso.account_identifier(
	account_id char(28) not null,
	identifier text not null,
	created_at timestamp with time zone not null
) WITHOUT OIDS;

CREATE TABLE IF NOT EXISTS sso.client(
	id char(28) not null,
	name text not null,
	type smallint not null default 0,
	redirect_uris text[] not null default '{}',
	secret char(60),
	created_at timestamp with time zone not null,
	updated_at timestamp with time zone not null,
	deleted_at timestamp with time zone
) WITHOUT OIDS;

CREATE TABLE IF NOT EXISTS sso.scope(
	id char(28) not null,
	client_id char(28) not null,
	scope text not null,
	created_at timestamp with time zone not null,
	updated_at timestamp with time zone not null,
	deleted_at timestamp with time zone
) WITHOUT OIDS;

CREATE TABLE IF NOT EXISTS sso.scope_i18n(
	scope_id char(28) not null,
	locale char(5) not null,
	description text not null,
	created_at timestamp with time zone not null,
	updated_at timestamp with time zone not null,
	deleted_at timestamp with time zone
) WITHOUT OIDS;

CREATE TABLE IF NOT EXISTS sso.grant(
	account_id char(28) not null,
	scope_id char(28) not null,
	created_at timestamp with time zone not null,
	updated_at timestamp with time zone not null,
	deleted_at timestamp with time zone
) WITHOUT OIDS;

CREATE TABLE IF NOT EXISTS sso.token(
	id char(36) not null,
	account_id char(28) not null,
	client_id char(28) not null,
	type smallint not null default 0,
	expires_at timestamp with time zone not null,
	created_at timestamp with time zone not null,
	updated_at timestamp with time zone not null,
	deleted_at timestamp with time zone
) WITHOUT OIDS;

ALTER TABLE sso.account ADD CONSTRAINT pk_sso_account PRIMARY KEY(id);
ALTER TABLE sso.account ADD CONSTRAINT unq_sso_account_id UNIQUE (id);

ALTER TABLE sso.account_identifier ADD CONSTRAINT unq_sso_account_identifier_identifier UNIQUE (identifier);
ALTER TABLE sso.account_identifier ADD CONSTRAINT fk_sso_account_identifier_account FOREIGN KEY (account_id) REFERENCES sso.account (id);

ALTER TABLE sso.client ADD CONSTRAINT pk_sso_client PRIMARY KEY(id);
ALTER TABLE sso.client ADD CONSTRAINT unq_sso_client_id UNIQUE (id);

ALTER TABLE sso.scope ADD CONSTRAINT pk_sso_scope PRIMARY KEY(id);
ALTER TABLE sso.scope ADD CONSTRAINT unq_sso_scope_id UNIQUE (id);
ALTER TABLE sso.scope ADD CONSTRAINT unq_sso_scope_scope_client_id UNIQUE (scope, client_id);
ALTER TABLE sso.scope ADD CONSTRAINT fk_sso_scope_client_id FOREIGN KEY (client_id) REFERENCES sso.client (id);
CREATE INDEX idx_sso_scope_scope ON sso.scope USING BTREE (scope);

ALTER TABLE sso.scope_i18n ADD CONSTRAINT pk_sso_scope_i18n PRIMARY KEY(scope_id, locale);
ALTER TABLE sso.scope_i18n ADD CONSTRAINT unq_sso_scope_i18n_scope_id_locale UNIQUE (scope_id, locale);
ALTER TABLE sso.scope_i18n ADD CONSTRAINT fk_sso_scope_i18n_scope_id FOREIGN KEY (scope_id) REFERENCES sso.scope (id);

ALTER TABLE sso.grant ADD CONSTRAINT pk_sso_grant PRIMARY KEY(scope_id, account_id);
ALTER TABLE sso.grant ADD CONSTRAINT unq_sso_grant_scope_id_account_id UNIQUE (scope_id, account_id);
ALTER TABLE sso.grant ADD CONSTRAINT fk_sso_grant_scope_id FOREIGN KEY (scope_id) REFERENCES sso.scope (id);
ALTER TABLE sso.grant ADD CONSTRAINT fk_sso_grant_account_id FOREIGN KEY (account_id) REFERENCES sso.account (id);

ALTER TABLE sso.token ADD CONSTRAINT pk_sso_token PRIMARY KEY(id);
ALTER TABLE sso.token ADD CONSTRAINT pk_sso_token_id UNIQUE (id);
ALTER TABLE sso.token ADD CONSTRAINT fk_sso_token_account_id FOREIGN KEY (account_id) REFERENCES sso.account (id);
ALTER TABLE sso.token ADD CONSTRAINT fk_sso_token_client_id FOREIGN KEY (client_id) REFERENCES sso.client (id);