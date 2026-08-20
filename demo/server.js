const port = process.env.PORT || 3000;

if (!process.env.DB_URL || !process.env.JWT_SECRET) {
  console.error('error: DB_URL and JWT_SECRET are required');
  process.exit(1);
}

console.log(`server listening on port ${port}`);
