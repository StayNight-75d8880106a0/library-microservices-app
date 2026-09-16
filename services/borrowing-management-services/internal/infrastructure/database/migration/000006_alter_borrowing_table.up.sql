ALTER TABLE borrowings
    ADD COLUMN active_slot VARCHAR(80)
    GENERATED ALWAYS AS (
        CASE
            WHEN status IN ('PENDING', 'BORROWED') THEN CONCAT(user_id, ':', book_id)
            ELSE NULL
        END
    ) STORED;

ALTER TABLE borrowings ADD UNIQUE KEY uq_active_borrowing (active_slot);
