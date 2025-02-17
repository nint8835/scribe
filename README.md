# Scribe

Discord bot for recording & recalling quotes

## Setup

- Install Go 1.24 or above
- [Install the Tailwind CLI](https://tailwindcss.com/docs/installation).
  - Ensure the `tailwindcss` command is accessible from `$PATH`.
  - If you have homebrew installed, use: `brew install tailwindcss`.

### Setup the config

```shell-session
cp .env.dist .env
```

Create an application on [https://discord.com/developers/applications/](https://discord.com/developers/applications/) and fill out at least the `SCRIBE_TOKEN`, `SCRIBE_GUILD_ID`, and `SCRIBE_APP_ID` fields in `.env`

### Running the app

Just run `air`

```shell-session
go tool air
```

## Production data

If you'd like to have the production data to play around with,

- Use `/db` in the server you are in with Scribe.
- Download the `quotes.sqlite` file.
- Move that file to your `SQLITE_DB_PATH` location, the default is the root of the project directory.

## Website

To develop the website, you'll need to ensure its setup to let you log into it using Discord OAuth2

- Head to OAuth2, grab your client id & secret, set them up in your `.env`.
- Set `SCRIBE_COOKIE_SECRET` in `.env` to be a random value of your choice.
- Setup a callback URL, both in your bot's OAuth2 tab, and your `.env` file
  - For local development, it should probably be: http://localhost:8000/auth/callback

Then, head to [http://localhost:8000/](http://localhost:8000/), you should be able to log in, and view the web UI.
