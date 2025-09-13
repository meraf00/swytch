-- Create enum type "task_status"
CREATE TYPE "task_status" AS ENUM ('pending', 'processing', 'completed', 'failed');
-- Create "files" table
CREATE TABLE "files" (
  "id" serial NOT NULL,
  "original_name" character varying(255) NOT NULL,
  "original_format" character varying(50) NOT NULL,
  "created_at" timestamptz NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamptz NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY ("id")
);
-- Create "jobs" table
CREATE TABLE "jobs" (
  "id" serial NOT NULL,
  "created_at" timestamptz NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamptz NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY ("id")
);
-- Create "tasks" table
CREATE TABLE "tasks" (
  "id" serial NOT NULL,
  "file_id" integer NULL,
  "job_id" integer NULL,
  "converted_file_name" character varying(255) NULL,
  "target_format" character varying(50) NOT NULL,
  "status" "task_status" NULL DEFAULT 'pending',
  "started_at" timestamptz NULL,
  "completed_at" timestamptz NULL,
  "error_message" text NULL,
  "created_at" timestamptz NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamptz NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY ("id"),
  CONSTRAINT "tasks_file_id_fkey" FOREIGN KEY ("file_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "tasks_job_id_fkey" FOREIGN KEY ("job_id") REFERENCES "jobs" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "idx_tasks_file_id" to table: "tasks"
CREATE INDEX "idx_tasks_file_id" ON "tasks" ("file_id");
-- Create index "idx_tasks_job_id" to table: "tasks"
CREATE INDEX "idx_tasks_job_id" ON "tasks" ("job_id");
-- Create index "idx_tasks_status" to table: "tasks"
CREATE INDEX "idx_tasks_status" ON "tasks" ("status");

-- Trigger to update updated_at on row modification
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
   NEW.updated_at = clock_timestamp();
   RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_files_updated_at
BEFORE UPDATE ON files
FOR EACH ROW
EXECUTE PROCEDURE update_updated_at_column();

CREATE TRIGGER update_tasks_updated_at
BEFORE UPDATE ON tasks
FOR EACH ROW
EXECUTE PROCEDURE update_updated_at_column();

CREATE TRIGGER update_jobs_updated_at
BEFORE UPDATE ON jobs
FOR EACH ROW
EXECUTE PROCEDURE update_updated_at_column();