box.cfg{
  listen = 3301, -- Или ваш порт
  wal_mode = 'none', -- Для разработки, в продакшене использовать wal_mode = 'on'
  memtx_memory = 128 * 1024 * 1024 -- 128MB
}

-- polls (Voting model)
box.schema.space.create('votings', { if_not_exists = true })
box.space.votings:format({
  { name = 'id',         type = 'string' },
  { name = 'creator_id',  type = 'string' },
  { name = 'question',   type = 'string' },
  { name = 'channel_id', type = 'string' },
  { name = 'options',    type = 'array', array_type = 'string' }, -- Changed to array of strings
  { name = 'created_at', type = 'number' }, -- Store as timestamp (seconds since epoch)
  { name = 'closed_at',  type = 'number' }, -- Store as timestamp (seconds since epoch)
  { name = 'results',    type = 'array', array_type = 'unsigned' }, -- Array of unsigned integers (results)
  { name = 'is_active',  type = 'boolean' },
})

box.space.votings:create_index('primary', {
  parts = {'id'},
  unique = true,
  if_not_exists = true
})

if not box.space.votings.index.channel_id then
  box.space.votings:create_index('channel_id', {
      parts = {'channel_id'},
      unique = false,
      if_not_exists = true
  })
end