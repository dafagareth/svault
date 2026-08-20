const config = require('./config');
const express = require('express');

const app = express();
app.use(express.json());

app.get('/health', (_, res) => res.json({ status: 'ok', version: '1.0.0' }));
app.use('/auth', require('./routes/auth'));
app.use('/notes', require('./routes/notes'));

app.use((req, res) => res.status(404).json({ error: 'Not found' }));

app.use((err, req, res, next) => {
  console.error(err);
  res.status(500).json({ error: 'Internal server error' });
});

app.listen(config.port, () => {
  console.log(`DevNotes API running on http://localhost:${config.port}`);
});
