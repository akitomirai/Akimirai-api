ALTER TABLE usage_logs
    ADD COLUMN IF NOT EXISTS upstream_connection_reused BOOLEAN,
    ADD COLUMN IF NOT EXISTS upstream_connection_ready_ms INTEGER,
    ADD COLUMN IF NOT EXISTS upstream_dns_lookup_ms INTEGER,
    ADD COLUMN IF NOT EXISTS upstream_tcp_connect_ms INTEGER,
    ADD COLUMN IF NOT EXISTS upstream_tls_handshake_ms INTEGER,
    ADD COLUMN IF NOT EXISTS upstream_request_headers_written_ms INTEGER,
    ADD COLUMN IF NOT EXISTS upstream_response_headers_received_ms INTEGER,
    ADD COLUMN IF NOT EXISTS upstream_response_body_first_byte_ms INTEGER,
    ADD COLUMN IF NOT EXISTS upstream_first_event_ms INTEGER,
    ADD COLUMN IF NOT EXISTS request_first_output_character_ms INTEGER;
