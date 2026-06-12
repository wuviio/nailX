-- ロールバック: 全テーブルを逆順で削除
DROP TABLE IF EXISTS notifications;
DROP TABLE IF EXISTS reviews;
DROP TABLE IF EXISTS royalty_distributions;
DROP TABLE IF EXISTS payments;
DROP TABLE IF EXISTS bookings;
DROP TABLE IF EXISTS bids;
DROP TABLE IF EXISTS booking_requests;
DROP TABLE IF EXISTS salons;
DROP TABLE IF EXISTS design_royalty_nodes;
DROP TABLE IF EXISTS design_ips;
DROP TABLE IF EXISTS materials;
DROP TABLE IF EXISTS ar_sessions;
DROP TABLE IF EXISTS nail_profiles;
DROP TABLE IF EXISTS users;

DROP FUNCTION IF EXISTS update_updated_at();
DROP EXTENSION IF EXISTS vector;
DROP EXTENSION IF EXISTS "uuid-ossp";
