# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Flora Music Bot (also named WinxMusic in the code) is a Telegram music and video bot built with Python, using Pyrogram for Telegram API interaction and Py-Tgcalls for voice chat functionality. The bot can stream music and videos from various sources including YouTube, Spotify, Apple Music, SoundCloud, and local Telegram files.

## Repository Structure

- **WinxMusic/**: Main bot package
  - **core/**: Core bot functionality (bot setup, calls, etc.)
  - **platforms/**: Integrations with music platforms (YouTube, Spotify, etc.)
  - **plugins/**: Command handlers for bot features

## Setup and Running

### Prerequisites
- Python 3.8+
- MongoDB database
- Telegram API credentials
- Spotify API credentials (optional)

### Environment Setup
1. Clone the repository and navigate to the directory:
   ```bash
   git clone https://github.com/gabrielmaialva33/flora-music-bot && cd flora-music-bot
   ```

2. Run the setup script:
   ```bash
   bash setup
   ```

3. Configure the environment variables in a `.env` file based on the `sample.env`.

### Running the Bot
Run the bot with:
```bash
python3 -m WinxMusic
```

For production deployment, use tmux to keep the bot running:
```bash
tmux
python3 -m WinxMusic
# Detach with Ctrl+b, then d
```

## Core Architecture

1. **Bot Initialization (`__main__.py`)**: 
   - Loads plugins
   - Sets up assistants
   - Initializes voice calls

2. **Bot Class (`WinxMusic/core/bot.py`)**: 
   - Handles message processing
   - Plugin loading
   - Error handling

3. **Call Management (`WinxMusic/core/call.py`)**:
   - Manages Telegram voice chats
   - Handles stream playback

4. **Platform Integrations (`WinxMusic/platforms/`)**: 
   - Each file represents a different music platform integration
   - Handles URL parsing, metadata extraction, and stream retrieval

5. **Command Plugins (`WinxMusic/plugins/`)**: 
   - Command handlers that process user requests
   - Organized by feature (play, admin, etc.)

## Configuration

The bot is configured through environment variables. Required variables:
- `API_ID` and `API_HASH`: Telegram API credentials
- `BOT_TOKEN`: Bot token from BotFather
- `MONGO_DB_URI`: MongoDB connection string
- `LOG_GROUP_ID`: Telegram group ID for logging
- `OWNER_ID`: User ID of the bot owner
- `STRING_SESSIONS`: Pyrogram session strings for assistant accounts

Additional configuration options are available in `config/config.py`.

## Database Usage

The bot uses MongoDB to store:
- User preferences
- Chat settings
- Playlists
- Statistics

## Important Files

- `WinxMusic/__main__.py`: Entry point
- `WinxMusic/core/bot.py`: Main bot class
- `WinxMusic/core/call.py`: Voice chat management
- `WinxMusic/platforms/*.py`: Platform integrations
- `WinxMusic/plugins/play/*.py`: Core music playback functionality
- `config/config.py`: Configuration handling