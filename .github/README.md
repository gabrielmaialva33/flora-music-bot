# 🎵 **WinxMusic** 🎶

[**WinxMusic**](https://github.com/gabrielmaialva33/flora-music-bot) is a powerful, enhanced version of the original [*
*WinxMusicBot**](https://github.com/gabrielmaialva33/flora-music-bot), designed for seamless, high-quality music
streaming in
Telegram voice chats. Built with **Python** and **Pyrogram**, it offers a robust and user-friendly experience for music
lovers and bot developers alike. 🚀

## ⚙️ Configuration

Need help setting up? Check out our detailed configuration guide: [**Configuration Instructions
**](https://github.com/gabrielmaialva33/flora-music-bot/blob/master/config/README.md).

> [!TIP]
> **Looking to use cookies for authentication?**  
> See: [**Using Cookies for Authentication
**](https://github.com/gabrielmaialva33/flora-music-bot/blob/master/config/README.md#using-cookies-for-authentication)

## Quick Deployment Options

## Deploy on Heroku

Get started quickly by deploying to Heroku with just one click:

<a href="https://dashboard.heroku.com/new?template=https://github.com/gabrielmaialva33/flora-music-bot">
  <img src="https://img.shields.io/badge/Deploy%20To%20Heroku-red?style=for-the-badge&logo=heroku" width="200"/>
</a>

### 🖥️ VPS Deployment Guide

- **Update System and Install Dependencies**:
  ```bash
  sudo apt update && sudo apt upgrade -y && sudo apt install -y ffmpeg git python3-pip tmux nano
  ```

- **Install uv for Efficient Dependency Management**:
  ```bash
  pip install --upgrade uv
  ```


- **Clone the Repository:**
  ```bash
  git clone https://github.com/gabrielmaialva33/flora-music-bot && cd WinxMusic
  ```


- **Create and Activate a Virtual Environment:**
    - You can create and activate the virtual Environment before cloning the repo.
  ```bash
  uv venv .venv && source .venv/bin/activate
  ```

- Install Python Requirements:
  ```bash
  uv pip install -e .
  ```

- Copy and Edit Environment Variables:
  ```bash
  cp .env.exemple .env && nano .env
  ```
  After editing, press `Ctrl+X`, then `Y`, and press **Enter** to save the changes.

- Start a tmux Session to Keep the Bot Running:
  ```bash
  tmux
  ```

- Run the Bot:
  ```bash
  winxmusic
  ```

- Detach from the **tmux** Session (Bot keeps running):  
  Press `Ctrl+b`, then `d`

## 🤝 Get Support

We're here to help you every step of the way! Reach out through:

- **📝 GitHub Issues**: Report bugs or ask questions by [**opening an issue
  **](https://github.com/gabrielmaialva33/flora-music-botissues/new?assignees=&labels=question&title=support).

- **💬 Telegram Support**: Connect with us on [**Telegram**](https://t.me/TheTeamVk).

- **👥 Support Channel**: Join our community at[**Gabriel Maia**](https://t.me/mrootx).

## ⭐ Support the Original

Show your love for the project that started it all! If you're using or forking **WinxMusic**, please **star** the
original repository: [**⭐ WinxMusicBot**](https://github.com/gabrielmaialva33/winx-music-bot)

## ❣️ Show Your Support

Love WinxMusic? Help us grow the project with these simple actions:

- **⭐ Star the Original:** Give a star to [**WinxMusicBot**](https://github.com/gabrielmaialva33/flora-music-bot).

- **🍴 Fork & Contribute**: Dive into the code and contribute to [**WinxMusic
  **](https://github.com/gabrielmaialva33/flora-music-bot).

- **📢 Spread the Word**: Share your experience on [**Dev.to**](https://dev.to/), [**Medium**](https://medium.com/), or
  your personal blog.

Together, we can make **WinxMusic** and **WinxMusicBot** even better!

## 🙏 Acknowledgments

A huge thank you to [**Team Winx**](https://t.me/canalclubdaswinx) for creating the original [**WinxMusicBot
**](https://github.com/gabrielmaialva33/flora-music-bot), the foundation of this project. Though the original is now
inactive, its
legacy lives on.

Special gratitude to [**Pranav-Saraswat**](https://github.com/Pranav-Saraswat) for reviving the project with [*
*WinxMusicFork**](https://github.com/Pranav-Saraswat/WinxMusicFork) (now deleted), which inspired WinxMusic.

**WinxMusic** is an imported and enhanced version of the now-deleted **WinxMusicFork**, with ongoing improvements to
deliver the best music streaming experience.
