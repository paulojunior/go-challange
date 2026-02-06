-- Link products to their respective categories

-- Clothing: PROD001, PROD004, PROD007
UPDATE products
SET category_id = (SELECT id FROM categories WHERE code = 'CLOTHING')
WHERE code IN ('PROD001', 'PROD004', 'PROD007');

-- Shoes: PROD002, PROD006
UPDATE products
SET category_id = (SELECT id FROM categories WHERE code = 'SHOES')
WHERE code IN ('PROD002', 'PROD006');

-- Accessories: PROD003, PROD005, PROD008
UPDATE products
SET category_id = (SELECT id FROM categories WHERE code = 'ACCESSORIES')
WHERE code IN ('PROD003', 'PROD005', 'PROD008');
