const router = require('express').Router();
const bcrypt = require('bcryptjs');
const jwt = require('jsonwebtoken');
const { PrismaClient } = require('@prisma/client');
const { jwtSecret, jwtExpiresIn, resendApiKey, emailFrom } = require('../config');

const prisma = new PrismaClient();

router.post('/register', async (req, res) => {
  const { email, password, name } = req.body;

  if (!email || !password || !name) {
    return res.status(422).json({ error: 'email, password, and name are required' });
  }

  const exists = await prisma.user.findUnique({ where: { email } });
  if (exists) {
    return res.status(409).json({ error: 'Email already registered' });
  }

  const hashed = await bcrypt.hash(password, 12);
  const user = await prisma.user.create({
    data: { email, password: hashed, name },
    select: { id: true, email: true, name: true, createdAt: true },
  });

  if (resendApiKey) {
    const { Resend } = require('resend');
    const resend = new Resend(resendApiKey);
    await resend.emails.send({
      from: emailFrom,
      to: email,
      subject: 'Welcome to DevNotes!',
      html: `<h1>Hi ${name}!</h1><p>Your account has been created. Happy coding!</p>`,
    }).catch(() => {});
  }

  const token = jwt.sign({ sub: user.id, email: user.email }, jwtSecret, { expiresIn: jwtExpiresIn });
  res.status(201).json({ user, token });
});

router.post('/login', async (req, res) => {
  const { email, password } = req.body;

  const user = await prisma.user.findUnique({ where: { email } });
  if (!user || !(await bcrypt.compare(password, user.password))) {
    return res.status(401).json({ error: 'Invalid credentials' });
  }

  const token = jwt.sign({ sub: user.id, email: user.email }, jwtSecret, { expiresIn: jwtExpiresIn });
  res.json({
    user: { id: user.id, email: user.email, name: user.name },
    token,
  });
});

module.exports = router;
