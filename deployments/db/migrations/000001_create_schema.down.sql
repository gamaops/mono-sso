
ALTER TABLE sso.token DROP CONSTRAINT pk_sso_token;
ALTER TABLE sso.token DROP CONSTRAINT pk_sso_token_id;
ALTER TABLE sso.token DROP CONSTRAINT fk_sso_token_account_id;
ALTER TABLE sso.token DROP CONSTRAINT fk_sso_token_client_id;

ALTER TABLE sso.grant DROP CONSTRAINT pk_sso_grant;
ALTER TABLE sso.grant DROP CONSTRAINT unq_sso_grant_scope_id_account_id;
ALTER TABLE sso.grant DROP CONSTRAINT fk_sso_grant_scope_id;
ALTER TABLE sso.grant DROP CONSTRAINT fk_sso_grant_account_id;

ALTER TABLE sso.scope_i18n DROP CONSTRAINT pk_sso_scope_i18n;
ALTER TABLE sso.scope_i18n DROP CONSTRAINT unq_sso_scope_i18n_scope_id_locale;
ALTER TABLE sso.scope_i18n DROP CONSTRAINT fk_sso_scope_i18n_scope_id;

ALTER TABLE sso.scope DROP CONSTRAINT pk_sso_scope;
ALTER TABLE sso.scope DROP CONSTRAINT unq_sso_scope_id;
ALTER TABLE sso.scope DROP CONSTRAINT unq_sso_scope_scope_client_id;
ALTER TABLE sso.scope DROP CONSTRAINT fk_sso_scope_client_id;
DROP INDEX IF EXISTS idx_sso_scope_scope;

ALTER TABLE sso.client DROP CONSTRAINT pk_sso_client;
ALTER TABLE sso.client DROP CONSTRAINT unq_sso_client_id;

ALTER TABLE sso.account_identifier DROP CONSTRAINT unq_sso_account_identifier_identifier;
ALTER TABLE sso.account_identifier DROP CONSTRAINT fk_sso_account_identifier_account;

ALTER TABLE sso.account DROP CONSTRAINT pk_sso_account;
ALTER TABLE sso.account DROP CONSTRAINT unq_sso_account_id;

DROP TABLE IF EXISTS sso.account;
DROP TABLE IF EXISTS sso.account_identifier;
DROP TABLE IF EXISTS sso.client;
DROP TABLE IF EXISTS sso.scope;
DROP TABLE IF EXISTS sso.scope_i18n;
DROP TABLE IF EXISTS sso.grant;
DROP TABLE IF EXISTS sso.token;

DROP SCHEMA IF EXISTS sso;