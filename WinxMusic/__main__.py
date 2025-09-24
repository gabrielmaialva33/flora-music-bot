import asyncio
import os

from pyrogram import idle
from pytgcalls.exceptions import NoActiveGroupCall

import config
from WinxMusic import HELPABLE, LOGGER, app, userbot
from WinxMusic.core.call import Winx
from WinxMusic.misc import sudo
from WinxMusic.utils.database import get_banned_users, get_gbanned
from config import BANNED_USERS

logger = LOGGER("WinxMusic")
loop = asyncio.get_event_loop()


async def init():
    if len(config.STRING_SESSIONS) == 0:
        logger.error("No Assistant Clients Vars Defined!.. Exiting Process.")
        return
    if not config.SPOTIFY_CLIENT_ID and not config.SPOTIFY_CLIENT_SECRET:
        logger.warning(
            "No Spotify Vars defined. Your bot won't be able to play spotify queries."
        )
    try:
        users = await get_gbanned()
        for user_id in users:
            BANNED_USERS.add(user_id)
        users = await get_banned_users()
        for user_id in users:
            BANNED_USERS.add(user_id)
    except Exception:
        pass
    await sudo()
    await app.start()
    for mod in app.load_plugins_from("WinxMusic/plugins"):
        if mod and hasattr(mod, "__MODULE__") and mod.__MODULE__:
            if hasattr(mod, "__HELP__") and mod.__HELP__:
                HELPABLE[mod.__MODULE__.lower()] = mod

    if config.EXTRA_PLUGINS:
        if os.path.exists("xtraplugins"):
            result = await app.run_shell_command(["git", "-C", "xtraplugins", "pull"])
            if result["returncode"] != 0:
                logger.error(
                    f"Error pulling updates for extra plugins: {result['stderr']}"
                )
                exit()
        else:
            result = await app.run_shell_command(
                ["git", "clone", config.EXTRA_PLUGINS_REPO, "xtraplugins"]
            )
            if result["returncode"] != 0:
                logger.error(f"Error cloning extra plugins: {result['stderr']}")
                exit()

        req = os.path.join("xtraplugins", "requirements.txt")
        if os.path.exists(req):
            result = await app.run_shell_command(
                ["uv", "pip", "install", "--system", "-r", req]
            )
            if result["returncode"] != 0:
                logger.error(f"Error installing requirements: {result['stderr']}")

        for mod in app.load_plugins_from("xtraplugins"):
            if mod and hasattr(mod, "__MODULE__") and mod.__MODULE__:
                if hasattr(mod, "__HELP__") and mod.__HELP__:
                    HELPABLE[mod.__MODULE__.lower()] = mod

    LOGGER("WinxMusic.plugins").info("Successfully Imported All Modules ")
    await userbot.start()
    await Winx.start()
    LOGGER("WinxMusic").info("Assistant Started Sucessfully")
    try:
        await Winx.stream_call(
            "http://docs.evostream.com/sample_content/assets/sintel1m720p.mp4"
        )
    except NoActiveGroupCall:
        LOGGER("WinxMusic").error(
            "Please ensure the voice call in your log group is active."
        )
        exit()

    await Winx.decorators()
    LOGGER("WinxMusic").info("WinxMusic Started Successfully")
    await idle()
    await app.stop()
    await userbot.stop()
    await Winx.stop()


def main():
    loop.run_until_complete(init())
    LOGGER("WinxMusic").info("Stopping WinxMusic! GoodBye")


if __name__ == "__main__":
    main()
