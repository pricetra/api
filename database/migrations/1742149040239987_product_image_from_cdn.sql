update "product"
set "image" = concat(
        'https://res.cloudinary.com/pricetra-cdn/image/upload/',
        "product"."code"
    );
