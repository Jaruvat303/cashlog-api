-- คำสั่งสำหรับเติมข้อมูลบัญชีเริ่มต้น (Master Data)

INSERT INTO accounts (name, account_type, opening_balance, matching_keywords, icon_key, color_hex, is_active, created_at, updated_at) VALUES
('SCB', 'bank', 5000, '["SCB","ไทยพาณิชย์"]'::jsonb, '', '', true, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)

ON CONFLICT (name)
DO UPDATE SET
    account_type = EXCLUDED.account_type,
    matching_keywords = EXCLUDED.matching_keywords,
    updated_at = CURRENT_TIMESTAMP; -- อัปเดตเวลาแก้ไขล่าสุดเมื่อข้อมูลซ้ำ (ไม่แตะ opening_balance/icon_key/color_hex/is_active ที่ user อาจแก้เองแล้ว)
