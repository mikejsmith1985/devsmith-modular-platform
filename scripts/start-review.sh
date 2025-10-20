#!/bin/sh
# Wait for Postgres, then start the review service
/wait-for-postgres.sh postgres:5432 30
exec ./review
