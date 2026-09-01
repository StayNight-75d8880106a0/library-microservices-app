CREATE TABLE waiting_lists (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    book_id VARCHAR(36) NOT NULL,
    arrival_rate   DECIMAL(10,4)  NOT NULL,                -- λ (request/jam)
    service_rate   DECIMAL(10,4)  NOT NULL,                -- μ (request/jam per server)
    num_servers    INT            NOT NULL,                -- c
    utilization    DECIMAL(6,4)   NOT NULL,                -- ρ
    prob_wait      DECIMAL(6,4)   NOT NULL,                -- P(wait)
    avg_queue_len  DECIMAL(10,4)  NOT NULL,                -- Lq
    avg_wait_min   DECIMAL(10,4)  NOT NULL,                -- Wq (menit)
    optimal_servers INT           NULL,      
    queue_number INT NOT NULL,
    status ENUM('WAITING', 'NOTIFIED', 'CANCELLED', 'FULFILLED') NOT NULL DEFAULT 'WAITING',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    INDEX idx_book_queue (book_id, queue_number)
);