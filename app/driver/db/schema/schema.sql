-- Add new schema named "public"
CREATE SCHEMA IF NOT EXISTS "public";
-- Set comment to schema: "public"
COMMENT ON SCHEMA "public" IS 'standard public schema';
-- Create "articles" table
CREATE TABLE "public"."articles" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "slug" character varying(20) NOT NULL,
  "title" character varying(100) NOT NULL,
  "body" text NOT NULL,
  "status" character varying(20) NOT NULL DEFAULT 'draft',
  "published_at" timestamptz NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("id"),
  CONSTRAINT "articles_slug_key" UNIQUE ("slug"),
  CONSTRAINT "articles_status_check" CHECK ((status)::text = ANY ((ARRAY['draft'::character varying, 'published'::character varying])::text[]))
);
-- Create index "articles_published_recent_idx" to table: "articles"
CREATE INDEX "articles_published_recent_idx" ON "public"."articles" ("published_at" DESC) WHERE ((status)::text = 'published'::text);
-- Create "schema_migrations" table
CREATE TABLE "public"."schema_migrations" (
  "version" bigint NOT NULL,
  "dirty" boolean NOT NULL,
  PRIMARY KEY ("version")
);
-- Create "tags" table
CREATE TABLE "public"."tags" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "slug" character varying(20) NOT NULL,
  "name" character varying(30) NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("id"),
  CONSTRAINT "tags_slug_key" UNIQUE ("slug")
);
-- Create "article_tags" table
CREATE TABLE "public"."article_tags" (
  "article_id" uuid NOT NULL,
  "tag_id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("article_id", "tag_id"),
  CONSTRAINT "article_tags_article_id_fkey" FOREIGN KEY ("article_id") REFERENCES "public"."articles" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "article_tags_tag_id_fkey" FOREIGN KEY ("tag_id") REFERENCES "public"."tags" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "article_tags_tag_id_idx" to table: "article_tags"
CREATE INDEX "article_tags_tag_id_idx" ON "public"."article_tags" ("tag_id");
-- Create "series" table
CREATE TABLE "public"."series" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "slug" character varying(20) NOT NULL,
  "title" character varying(100) NOT NULL,
  "description" text NOT NULL DEFAULT '',
  "status" character varying(20) NOT NULL DEFAULT 'draft',
  "published_at" timestamptz NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("id"),
  CONSTRAINT "series_slug_key" UNIQUE ("slug"),
  CONSTRAINT "series_status_check" CHECK ((status)::text = ANY ((ARRAY['draft'::character varying, 'published_ongoing'::character varying, 'published_completed'::character varying])::text[]))
);
-- Create index "series_published_recent_idx" to table: "series"
CREATE INDEX "series_published_recent_idx" ON "public"."series" ("published_at" DESC) WHERE ((status)::text ~~ 'published_%'::text);
-- Create "series_articles" table
CREATE TABLE "public"."series_articles" (
  "series_id" uuid NOT NULL,
  "article_id" uuid NOT NULL,
  "position" integer NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("article_id"),
  CONSTRAINT "series_articles_series_id_position_key" UNIQUE ("series_id", "position"),
  CONSTRAINT "series_articles_article_id_fkey" FOREIGN KEY ("article_id") REFERENCES "public"."articles" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "series_articles_series_id_fkey" FOREIGN KEY ("series_id") REFERENCES "public"."series" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create "series_tags" table
CREATE TABLE "public"."series_tags" (
  "series_id" uuid NOT NULL,
  "tag_id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("series_id", "tag_id"),
  CONSTRAINT "series_tags_series_id_fkey" FOREIGN KEY ("series_id") REFERENCES "public"."series" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "series_tags_tag_id_fkey" FOREIGN KEY ("tag_id") REFERENCES "public"."tags" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "series_tags_tag_id_idx" to table: "series_tags"
CREATE INDEX "series_tags_tag_id_idx" ON "public"."series_tags" ("tag_id");
