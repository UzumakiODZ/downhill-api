-- Create "companies" table
CREATE TABLE "public"."companies" (
  "id" bigserial NOT NULL,
  "company_name" character varying(255) NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "uni_companies_company_name" UNIQUE ("company_name")
);
-- Create "users" table
CREATE TABLE "public"."users" (
  "id" bigserial NOT NULL,
  "username" character varying(255) NOT NULL,
  "email" character varying(255) NOT NULL,
  "reg_id" character varying(255) NOT NULL,
  "password" character varying(255) NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "uni_users_email" UNIQUE ("email"),
  CONSTRAINT "uni_users_reg_id" UNIQUE ("reg_id"),
  CONSTRAINT "uni_users_username" UNIQUE ("username")
);
-- Create "posts" table
CREATE TABLE "public"."posts" (
  "id" bigserial NOT NULL,
  "title" character varying(255) NULL,
  "content" text NULL,
  "user_id" bigint NULL,
  "company_id" bigint NULL,
  "created_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_posts_company" FOREIGN KEY ("company_id") REFERENCES "public"."companies" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "fk_posts_user" FOREIGN KEY ("user_id") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create "question_banks" table
CREATE TABLE "public"."question_banks" (
  "id" bigserial NOT NULL,
  "question" text NULL,
  "company_id" bigint NULL,
  "years" bigint NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_question_banks_company" FOREIGN KEY ("company_id") REFERENCES "public"."companies" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create "roles" table
CREATE TABLE "public"."roles" (
  "id" bigserial NOT NULL,
  "role_name" character varying(255) NULL,
  "year" bigint NULL,
  "company_id" bigint NULL,
  "offer_type" character varying(255) NULL,
  "cgpa" numeric NULL,
  "other" character varying(255) NULL,
  "ctc" numeric NULL,
  "base" numeric NULL,
  "hired" bigint NULL,
  "converted" bigint NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_roles_company" FOREIGN KEY ("company_id") REFERENCES "public"."companies" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
