DROP TABLE IF EXISTS "athletes";
DROP SEQUENCE IF EXISTS athetes_id_seq;
CREATE SEQUENCE athetes_id_seq INCREMENT 1 MINVALUE 1 MAXVALUE 2147483647 START 1 CACHE 1;

CREATE TABLE "public"."athletes" (
    "id" integer DEFAULT nextval('athetes_id_seq') NOT NULL,
    "strava_id" integer,
    "name" text,
    "code" text,
    "access_token" text,
    "refresh_token" text,
    "raw" jsonb,
    "lastupdated" timestamp NOT NULL,
    CONSTRAINT "athletes_pkey" PRIMARY KEY ("id")
) WITH (oids = false);
