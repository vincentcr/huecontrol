
INSERT INTO users(id, email,password) VALUES
  ('86eb1856a155497aac7fd7ef50e7d2df', 'vincentcr@gmail.com', crypt('abcdefg', gen_salt('bf', 8)))
;

VACUUM ANALYZE;
