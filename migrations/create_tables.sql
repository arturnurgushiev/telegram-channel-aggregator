CREATE TABLE IF NOT EXISTS subscriptions (
    user_id BIGINT NOT NULL,
    channel_username TEXT NOT NULL,
    PRIMARY KEY (user_id, channel_username)
);

CREATE TABLE IF NOT EXISTS channel_states (
    channel_username TEXT PRIMARY KEY,
    last_post_id BIGINT DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_channel_username ON subscriptions(channel_username);
CREATE INDEX IF NOT EXISTS idx_user_id ON subscriptions(user_id);

