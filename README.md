# Scribe

Discord bot for recording & recalling quotes

## Setup

Compile the CSS:

```shell-session
npm install tailwindcss @tailwindcss/typography
npx tailwindcss build -o pkg/web/static/tailwind.css
# if you have the standalone cli, just the above command without `npx` will suffice
```

Setup config:

```shell-session
cp .env.dist .env
# fill out at least the bot token, guild id, and app id from discord.com/developers/applications/...
```

Run the bot:

```shell-session
go run ./main.go
```

_TODO: explain how to setup oauth stuff for a functioning web server_
