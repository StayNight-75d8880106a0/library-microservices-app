ALTER TABLE waiting_lists
    ADD UNIQUE KEY uq_book_queue (book_id, queue_number);