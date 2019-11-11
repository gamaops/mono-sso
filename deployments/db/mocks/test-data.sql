INSERT INTO sso.account (
  id,
  name,
  activation_method,
  password,
  created_at,
  updated_at,
  deleted_at
) VALUES (
  '06g9vjujs4p0qbvseegugtba5qkq',
  'Victor França Lopes',
  0,
  '$2y$10$Wo3x3hIVvNzvUfunPvMI2eAaROZrtB3CUWWWlzEDWkROW0a5aB9/G',
  now(),
  now(),
  null
), (
  '9u18u5hc7trf1o12mu0q20sb5qkq',
  'Marcos Jr',
  1,
  '$2y$10$Wo3x3hIVvNzvUfunPvMI2eAaROZrtB3CUWWWlzEDWkROW0a5aB9/G',
  now(),
  now(),
  null
);


INSERT INTO sso.tenant (
  id,
  name,
  created_at,
  updated_at,
  deleted_at
) VALUES (
  'pb6e3l41j17g2046a9bq5fek5qkq',
  'SSO Administration',
  now(),
  now(),
  null
);

INSERT INTO sso.account_tenant (
  tenant_id,
  account_id,
  created_at,
  updated_at,
  deleted_at
) VALUES (
  'pb6e3l41j17g2046a9bq5fek5qkq',
  '06g9vjujs4p0qbvseegugtba5qkq',
  now(),
  now(),
  null
),(
  'pb6e3l41j17g2046a9bq5fek5qkq',
  '9u18u5hc7trf1o12mu0q20sb5qkq',
  now(),
  now(),
  null
);

INSERT INTO sso.account_identifier (
  account_id,
	identifier,
	created_at
) VALUES (
  '06g9vjujs4p0qbvseegugtba5qkq',
  'vflopes',
  now()
), (
  '9u18u5hc7trf1o12mu0q20sb5qkq',
  'codermarcos',
  now()
);

INSERT INTO sso.client (
  id,
  name,
  type,
  redirect_uris,
  secret,
  created_at,
  updated_at,
  deleted_at
) VALUES (
  'g5e9p32jtjp0qbs3vugugtba5qkq',
  'Super App',
  1,
  ARRAY['https://localhost:3230/sign-in'],
  '$2y$10$Wo3x3hIVvNzvUfunPvMI2eAaROZrtB3CUWWWlzEDWkROW0a5aB9/G',
  now(),
  now(),
  null
);

INSERT INTO sso.scope(
  id,
  client_id,
  scope,
  created_at,
  updated_at,
  deleted_at
) VALUES (
  'aog0bburrjp0qb1tr7fugtba5qkq',
  'g5e9p32jtjp0qbs3vugugtba5qkq',
  'profile:write',
  now(),
  now(),
  null
), (
  '38je3ohlrbp0qb1e27fegtba5qkq',
  'g5e9p32jtjp0qbs3vugugtba5qkq',
  'profile:read',
  now(),
  now(),
  null
);