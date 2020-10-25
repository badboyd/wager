CREATE TABLE "wagers" (
  "id" SERIAL PRIMARY KEY,
  "odds" int,
  "total_wager_value" int,
  "selling_percentage" int,
  "selling_price" numeric,
  "current_selling_price" numeric,
  "percentage_sold" int DEFAULT null,
  "amount_sold" int DEFAULT null,
  "placed_at" timestamp NOT NULL DEFAULT NOW()
);

CREATE TABLE "purchases" (
  "id" SERIAL PRIMARY KEY,
  "wager_id" int,
  "buying_price" numeric,
  "bought_at" timestamp NOT NULL DEFAULT NOW()
);

ALTER TABLE "purchases" ADD FOREIGN KEY ("wager_id") REFERENCES "wagers" ("id");
