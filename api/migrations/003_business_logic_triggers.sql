-- Migration: 003_business_logic_triggers.sql
-- Description: Add business logic triggers for TT Stock Backend API
-- Created: 2024-01-01
-- Author: TT Stock Development Team

-- Function to update product quantity based on stock movements
CREATE OR REPLACE FUNCTION update_product_quantity()
RETURNS TRIGGER AS $$
BEGIN
    -- Update product quantity based on movement type
    IF NEW.movement_type = 'In' THEN
        UPDATE products 
        SET quantity_on_hand = quantity_on_hand + NEW.quantity,
            updated_at = CURRENT_TIMESTAMP
        WHERE id = NEW.product_id;
    ELSIF NEW.movement_type = 'Out' THEN
        UPDATE products 
        SET quantity_on_hand = quantity_on_hand - NEW.quantity,
            updated_at = CURRENT_TIMESTAMP
        WHERE id = NEW.product_id;
    ELSIF NEW.movement_type = 'Adjustment' THEN
        -- For adjustments, quantity represents the new total
        UPDATE products 
        SET quantity_on_hand = NEW.quantity,
            updated_at = CURRENT_TIMESTAMP
        WHERE id = NEW.product_id;
    ELSIF NEW.movement_type = 'Return' THEN
        UPDATE products 
        SET quantity_on_hand = quantity_on_hand + NEW.quantity,
            updated_at = CURRENT_TIMESTAMP
        WHERE id = NEW.product_id;
    END IF;
    
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Function to create low stock alerts
CREATE OR REPLACE FUNCTION create_low_stock_alert()
RETURNS TRIGGER AS $$
DECLARE
    product_record RECORD;
    alert_count INTEGER;
BEGIN
    -- Get product information
    SELECT * INTO product_record FROM products WHERE id = NEW.product_id;
    
    -- Check if quantity is below threshold
    IF NEW.quantity_on_hand <= product_record.low_stock_threshold THEN
        -- Check if there's already an active low stock alert for this product
        SELECT COUNT(*) INTO alert_count 
        FROM alerts 
        WHERE product_id = NEW.product_id 
        AND alert_type = 'LowStock' 
        AND is_active = true;
        
        -- Create alert if none exists
        IF alert_count = 0 THEN
            INSERT INTO alerts (
                product_id,
                alert_type,
                priority,
                title,
                message,
                is_read,
                is_active
            ) VALUES (
                NEW.product_id,
                'LowStock',
                CASE 
                    WHEN NEW.quantity_on_hand = 0 THEN 'Critical'
                    WHEN NEW.quantity_on_hand <= (product_record.low_stock_threshold / 2) THEN 'High'
                    ELSE 'Medium'
                END,
                'Low Stock Alert: ' || product_record.brand || ' ' || product_record.model,
                'Product ' || product_record.sku || ' has only ' || NEW.quantity_on_hand || ' units remaining. Threshold: ' || product_record.low_stock_threshold,
                false,
                true
            );
        END IF;
    ELSE
        -- If quantity is above threshold, deactivate any existing low stock alerts
        UPDATE alerts 
        SET is_active = false, updated_at = CURRENT_TIMESTAMP
        WHERE product_id = NEW.product_id 
        AND alert_type = 'LowStock' 
        AND is_active = true;
    END IF;
    
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Function to clean up expired sessions
CREATE OR REPLACE FUNCTION cleanup_expired_sessions()
RETURNS TRIGGER AS $$
BEGIN
    -- Deactivate expired sessions
    UPDATE sessions 
    SET is_active = false, updated_at = CURRENT_TIMESTAMP
    WHERE expires_at < CURRENT_TIMESTAMP AND is_active = true;
    
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Function to validate stock movement quantities
CREATE OR REPLACE FUNCTION validate_stock_movement()
RETURNS TRIGGER AS $$
DECLARE
    current_quantity INTEGER;
BEGIN
    -- Get current product quantity
    SELECT quantity_on_hand INTO current_quantity FROM products WHERE id = NEW.product_id;
    
    -- Validate that we don't go below zero for 'Out' movements
    IF NEW.movement_type = 'Out' AND (current_quantity - NEW.quantity) < 0 THEN
        RAISE EXCEPTION 'Insufficient stock. Current quantity: %, Requested: %', current_quantity, NEW.quantity;
    END IF;
    
    -- Validate positive quantities
    IF NEW.quantity <= 0 THEN
        RAISE EXCEPTION 'Stock movement quantity must be positive. Got: %', NEW.quantity;
    END IF;
    
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create triggers (drop first if exists)
DROP TRIGGER IF EXISTS trigger_update_product_quantity ON stock_movements;
CREATE TRIGGER trigger_update_product_quantity
    AFTER INSERT ON stock_movements
    FOR EACH ROW EXECUTE FUNCTION update_product_quantity();

DROP TRIGGER IF EXISTS trigger_create_low_stock_alert ON products;
CREATE TRIGGER trigger_create_low_stock_alert
    AFTER UPDATE OF quantity_on_hand ON products
    FOR EACH ROW EXECUTE FUNCTION create_low_stock_alert();

DROP TRIGGER IF EXISTS trigger_cleanup_expired_sessions ON sessions;
CREATE TRIGGER trigger_cleanup_expired_sessions
    AFTER INSERT ON sessions
    FOR EACH ROW EXECUTE FUNCTION cleanup_expired_sessions();

DROP TRIGGER IF EXISTS trigger_validate_stock_movement ON stock_movements;
CREATE TRIGGER trigger_validate_stock_movement
    BEFORE INSERT ON stock_movements
    FOR EACH ROW EXECUTE FUNCTION validate_stock_movement();

-- Create a function to get product statistics
CREATE OR REPLACE FUNCTION get_product_statistics()
RETURNS TABLE (
    total_products BIGINT,
    active_products BIGINT,
    low_stock_products BIGINT,
    out_of_stock_products BIGINT,
    total_inventory_value DECIMAL(15,2)
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        COUNT(*) as total_products,
        COUNT(*) FILTER (WHERE is_active = true) as active_products,
        COUNT(*) FILTER (WHERE is_active = true AND quantity_on_hand <= low_stock_threshold) as low_stock_products,
        COUNT(*) FILTER (WHERE is_active = true AND quantity_on_hand = 0) as out_of_stock_products,
        COALESCE(SUM(quantity_on_hand * cost_price), 0) as total_inventory_value
    FROM products;
END;
$$ language 'plpgsql';

-- Create a function to get stock movement summary
CREATE OR REPLACE FUNCTION get_stock_movement_summary(
    start_date TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP - INTERVAL '30 days',
    end_date TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
)
RETURNS TABLE (
    movement_type VARCHAR(20),
    total_quantity BIGINT,
    movement_count BIGINT
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        sm.movement_type,
        SUM(sm.quantity) as total_quantity,
        COUNT(*) as movement_count
    FROM stock_movements sm
    WHERE sm.movement_date BETWEEN start_date AND end_date
    GROUP BY sm.movement_type
    ORDER BY sm.movement_type;
END;
$$ language 'plpgsql';
