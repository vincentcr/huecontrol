CREATE TABLE users(
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  email VARCHAR(256) UNIQUE NOT NULL CHECK(email ~ '^[a-zA-Z0-9_%+-]+@[a-zA-Z0-9-]+\.[a-zA-Z0-9][a-zA-Z0-9]+$'),
  password VARCHAR(128) NOT NULL
);

CREATE TABLE bridges(
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  user_id uuid REFERENCES users(id) NOT NULL,
  api_token VARCHAR(256) NOT NULL,
  device_id VARCHAR(256) NOT NULL
);

CREATE TYPE schedule_status AS ENUM ('enabled', 'disabled');
CREATE TABLE schedules(
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  bridge_id uuid REFERENCES bridges(id) NOT NULL,
  user_id uuid REFERENCES users(id) NOT NULL,
  name VARCHAR(32) NOT NULL,
  description VARCHAR(64) NOT NULL,
  command jsonb NOT NULL,
  local_time TEXT NOT NULL,
  status schedule_status NOT NULL
);
