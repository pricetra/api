update "ai_prompt_template"
set prompt = 'This OCR string represents a product with a UPC: "{{ocr_string}}". Extract the following:
- Brand.
- Product name (extract important information while keeping it under 150 characters, include brand, size, & count).
- Weight (optional - only for products with volume / weight units shown clearly. ex: ''20 fl oz'', ''1.1 lb'', ''2 l'').
- Category (Google product taxonomy format).
- Quantity (optional).
Respond with a single JSON object only, using this schema:
`{"brand":string,"productName":string,"weight"?:string,"quantity"?:number,"category":string}`.'
where "type" = 'PRODUCT_DETAILS'::"ai_prompt_type"
