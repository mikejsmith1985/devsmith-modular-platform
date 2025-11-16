-- Migration: Add size constraints to prevent massive log entries
-- Date: 2025-11-10
-- Purpose: Prevent log entries with metadata/messages larger than 10MB from being inserted

-- Add check constraint for message size (max 10MB)
ALTER TABLE logs.entries 
ADD CONSTRAINT check_message_size 
CHECK (length(message) <= 10485760);

-- Add check constraint for metadata size (max 5MB when serialized)
ALTER TABLE logs.entries 
ADD CONSTRAINT check_metadata_size 
CHECK (length(metadata::text) <= 5242880);

-- Add comment explaining the constraints
COMMENT ON CONSTRAINT check_message_size ON logs.entries IS 
'Prevents log messages larger than 10MB to avoid performance issues';

COMMENT ON CONSTRAINT check_metadata_size ON logs.entries IS 
'Prevents log metadata larger than 5MB (when serialized) to avoid performance issues';
