#!/usr/bin/env bash
set -e

NS="devnotes"

echo "Storing DevNotes secrets into svault..."
echo

svault unlock
svault use $NS

# Replace the values below according to your environment
svault set DATABASE_URL "file:./dev.db"
svault set JWT_SECRET "$(openssl rand -hex 32)"
svault set RESEND_API_KEY "your-resend-api-key"

echo
echo "Secrets saved in namespace [$NS]."
echo
echo "Run the application (secrets injected directly from svault):"
echo "  svault exec --ns $NS -- npm start"
echo
echo "Run with Docker:"
echo "  svault exec --ns $NS -- docker compose up"
