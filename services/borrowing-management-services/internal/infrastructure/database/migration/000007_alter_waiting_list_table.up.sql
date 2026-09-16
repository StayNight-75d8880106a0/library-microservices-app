ALTER TABLE waiting_lists
    ADD COLUMN active_slot VARCHAR(80)
    GENERATED ALWAYS AS (
        CASE
            WHEN status IN ('WAITING', 'NOTIFIED') THEN CONCAT(user_id, ':', book_id)
            ELSE NULL
        END
    ) STORED;

ALTER TABLE waiting_lists ADD UNIQUE KEY uq_active_waiting_list (active_slot);
