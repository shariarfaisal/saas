-- Create a function to get the next order number using restaurant's order_sequence
CREATE OR REPLACE FUNCTION next_order_number(p_restaurant_id UUID, p_prefix TEXT)
RETURNS TEXT AS $$
DECLARE
    v_seq BIGINT;
BEGIN
    UPDATE restaurants
    SET order_sequence = order_sequence + 1
    WHERE id = p_restaurant_id
    RETURNING order_sequence INTO v_seq;

    RETURN CONCAT(p_prefix, '-', LPAD(v_seq::TEXT, 6, '0'));
END;
$$ LANGUAGE plpgsql;
