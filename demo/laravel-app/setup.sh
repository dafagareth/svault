#!/usr/bin/env bash
set -e

NS="quicktask"

echo "Storing QuickTask secrets into svault..."
echo

svault unlock
svault use $NS

# Replace the values below according to your environment
svault set APP_KEY "$(php artisan key:generate --show 2>/dev/null || openssl rand -base64 32)"
svault set DB_HOST "127.0.0.1"
svault set DB_DATABASE "quicktask"
svault set DB_USERNAME "root"
svault set DB_PASSWORD "your-db-password"
svault set MAIL_HOST "sandbox.smtp.mailtrap.io"
svault set MAIL_USERNAME "your-mailtrap-username"
svault set MAIL_PASSWORD "your-mailtrap-password"

echo
echo "Secrets saved in namespace [$NS]."
echo
echo "Run the application (secrets injected directly from svault):"
echo "  svault exec --ns $NS -- php artisan serve"
echo
echo "Run with Docker:"
echo "  svault exec --ns $NS -- docker compose up"
