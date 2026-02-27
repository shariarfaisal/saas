-- 0002_create_enums.up.sql
-- All platform ENUM types (28 total)

CREATE TYPE tenant_status   AS ENUM ('pending','active','suspended','cancelled');
CREATE TYPE tenant_plan     AS ENUM ('starter','growth','enterprise');
CREATE TYPE user_role       AS ENUM ('customer','tenant_owner','tenant_admin','restaurant_manager','restaurant_staff','rider','platform_admin','platform_support','platform_finance');
CREATE TYPE user_status     AS ENUM ('active','suspended','deleted');
CREATE TYPE restaurant_type AS ENUM ('restaurant','cloud_kitchen','store','dark_store');
CREATE TYPE product_avail   AS ENUM ('available','unavailable','out_of_stock');
CREATE TYPE price_type      AS ENUM ('flat','variant');
CREATE TYPE order_status    AS ENUM ('pending','created','confirmed','preparing','ready','picked','delivered','cancelled','rejected');
CREATE TYPE pickup_status   AS ENUM ('new','confirmed','preparing','ready','picked','rejected');
CREATE TYPE payment_status  AS ENUM ('unpaid','paid','refunded','partially_refunded');
CREATE TYPE payment_method  AS ENUM ('cod','bkash','aamarpay','sslcommerz','wallet','card');
CREATE TYPE txn_status      AS ENUM ('pending','success','failed','refunded','cancelled');
CREATE TYPE promo_type      AS ENUM ('fixed','percent');
CREATE TYPE promo_apply_on  AS ENUM ('all_items','category','specific_restaurant','delivery_charge');
CREATE TYPE promo_funder    AS ENUM ('vendor','platform','restaurant');
CREATE TYPE invoice_status  AS ENUM ('draft','finalized','paid');
CREATE TYPE issue_type      AS ENUM ('wrong_item','missing_item','quality_issue','late_delivery','other');
CREATE TYPE issue_status    AS ENUM ('open','resolved','closed');
CREATE TYPE accountable     AS ENUM ('restaurant','rider','platform');
CREATE TYPE refund_status   AS ENUM ('pending','approved','rejected','processed');
CREATE TYPE rider_subject   AS ENUM ('attendance_in','attendance_out','picked','delivered','in_hub','location_update');
CREATE TYPE wallet_type     AS ENUM ('credit','debit');
CREATE TYPE wallet_source   AS ENUM ('cashback','referral','welcome','refund','order_payment','admin_adjustment');
CREATE TYPE platform_source AS ENUM ('web','ios','android','pos');
CREATE TYPE discount_type   AS ENUM ('fixed','percent');
CREATE TYPE vehicle_type    AS ENUM ('bicycle','motorcycle','car');
CREATE TYPE penalty_status  AS ENUM ('pending','cleared','appealed');
CREATE TYPE delivery_model  AS ENUM ('zone_based','distance_based');
