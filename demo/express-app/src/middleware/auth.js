const jwt = require('jsonwebtoken');
const { jwtSecret } = require('../config');

module.exports = function auth(req, res, next) {
  const header = req.headers.authorization;
  if (!header || !header.startsWith('Bearer ')) {
    return res.status(401).json({ error: 'Missing or invalid authorization header' });
  }

  try {
    req.user = jwt.verify(header.slice(7), jwtSecret);
    next();
  } catch {
    res.status(401).json({ error: 'Token expired or invalid' });
  }
};
