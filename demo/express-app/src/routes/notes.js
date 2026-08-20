const router = require('express').Router();
const { PrismaClient } = require('@prisma/client');
const auth = require('../middleware/auth');

const prisma = new PrismaClient();

router.use(auth);

router.get('/', async (req, res) => {
  const { search, language, pinned } = req.query;
  const where = { userId: req.user.sub };

  if (search) {
    where.OR = [
      { title: { contains: search } },
      { content: { contains: search } },
    ];
  }
  if (language) where.language = language;
  if (pinned !== undefined) where.pinned = pinned === 'true';

  const notes = await prisma.note.findMany({
    where,
    orderBy: [{ pinned: 'desc' }, { updatedAt: 'desc' }],
  });
  res.json(notes);
});

router.post('/', async (req, res) => {
  const { title, content, language, pinned } = req.body;
  if (!title || !content) {
    return res.status(422).json({ error: 'title and content are required' });
  }

  const note = await prisma.note.create({
    data: { title, content, language: language || 'text', pinned: pinned || false, userId: req.user.sub },
  });
  res.status(201).json(note);
});

router.get('/:id', async (req, res) => {
  const note = await prisma.note.findFirst({
    where: { id: Number(req.params.id), userId: req.user.sub },
  });
  if (!note) return res.status(404).json({ error: 'Note not found' });
  res.json(note);
});

router.put('/:id', async (req, res) => {
  const existing = await prisma.note.findFirst({
    where: { id: Number(req.params.id), userId: req.user.sub },
  });
  if (!existing) return res.status(404).json({ error: 'Note not found' });

  const note = await prisma.note.update({
    where: { id: existing.id },
    data: req.body,
  });
  res.json(note);
});

router.delete('/:id', async (req, res) => {
  const existing = await prisma.note.findFirst({
    where: { id: Number(req.params.id), userId: req.user.sub },
  });
  if (!existing) return res.status(404).json({ error: 'Note not found' });

  await prisma.note.delete({ where: { id: existing.id } });
  res.status(204).send();
});

module.exports = router;
