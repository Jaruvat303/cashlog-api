-- คำสั่งสำหรับเติมข้อมูลประเภทรายรับ-รายจ่ายเริ่มต้น (Master Data) ทั้งหมด 30 รายการ

INSERT INTO categories (name, type, icon_key, color_hex, created_at, updated_at) VALUES
-- ==========================================
-- หน้าจอรายจ่าย (Expense) - 20 รายการ
-- ==========================================
('อาหารและเครื่องดื่ม', 'expense', 'restaurant-fill', '#EF4444', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
('ของทานเล่นและคาเฟ่', 'expense', 'cup-fill', '#F97316', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
('ค่าของชำและวัตถุดิบเข้าบ้าน', 'expense', 'shopping-bag-3-fill', '#FBBF24', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
('ค่าเดินทางและขนส่งสาธารณะ', 'expense', 'bus-fill', '#3B82F6', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
('ค่าน้ำมันและดูแลรักษารถ', 'expense', 'car-fill', '#1E3A8A', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
('ค่าที่อยู่อาศัย (ค่าเช่า/ผ่อนบ้าน)', 'expense', 'home-4-fill', '#78350F', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
('ค่าน้ำและค่าไฟฟ้า', 'expense', 'flashlight-fill', '#F59E0B', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
('ค่าอินเทอร์เน็ตและโทรศัพท์', 'expense', 'wifi-fill', '#60A5FA', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
('เสื้อแฟชั่นและเครื่องแต่งกาย', 'expense', 't-shirt-fill', '#EC4899', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
('สกินแคร์และเครื่องสำอาง', 'expense', 'sparkles-fill', '#F472B6', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
('สตรีมมิ่งและความบันเทิง', 'expense', 'clapperboard-fill', '#8B5CF6', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
('สังสรรค์และปาร์ตี้', 'expense', 'goblet-fill', '#D97706', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
('ท่องเที่ยวและโรงแรม', 'expense', 'plane-fill', '#14B8A6', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
('ค่ารักษาพยาบาลและยา', 'expense', 'capsule-fill', '#E11D48', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
('เบี้ยประกัน (ชีวิต/สุขภาพ/รถ)', 'expense', 'shield-check-fill', '#4B5563', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
('การศึกษาและพัฒนาตนเอง', 'expense', 'book-open-fill', '#6B7280', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
('อุปกรณ์ไอทีและแกดเจ็ต', 'expense', 'computer-fill', '#0F172A', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
('ของใช้สัตว์เลี้ยง', 'expense', 'footprint-fill', '#9A3412', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
('เงินบริการและทำบุญ', 'expense', 'heart-3-fill', '#F43F5E', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
('รายจ่ายเบ็ดเตล็ดอื่น ๆ', 'expense', 'archive-fill', '#9CA3AF', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),

-- ==========================================
-- หน้าจอรายรับ (Income) - 10 รายการ
-- ==========================================
('เงินเดือนประจำ', 'income', 'wallet-3-fill', '#22C55E', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
('โบนัสและเงินรางวัล', 'income', 'gift-fill', '#10B981', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
('งานฟรีแลนซ์และพาร์ทไทม์', 'income', 'tools-fill', '#84CC16', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
('ธุรกิจส่วนตัว / ค้าขาย', 'income', 'store-2-fill', '#06B6D4', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
('เงินปันผลจากการลงทุน', 'income', 'line-chart-fill', '#2563EB', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
('กำไรจากการขายสินทรัพย์', 'income', 'trophy-fill', '#1D4ED8', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
('ค่าเช่ารับ (อสังหาริมทรัพย์)', 'income', 'key-2-fill', '#EA580C', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
('เงินช่วยเหลือ / เงินจากครอบครัว', 'income', 'hand-heart-fill', '#A855F7', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
('เครดิตเงินคืน / แคชแบ็ก', 'income', 'coins-fill', '#EAB308', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
('รายรับอื่น ๆ', 'income', 'money-dollar-circle-fill', '#71717A', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)

ON CONFLICT (name) 
DO UPDATE SET 
    icon_key = EXCLUDED.icon_key,
    color_hex = EXCLUDED.color_hex,
    updated_at = CURRENT_TIMESTAMP; -- อัปเดตเวลาแก้ไขล่าสุดเมื่อข้อมูลซ้ำ