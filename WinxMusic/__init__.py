import asyncio as _asyncio

import uvloop as _uvloop

_asyncio.set_event_loop_policy(_uvloop.EventLoopPolicy())  # noqa

from WinxMusic.core.bot import WinxBot
from WinxMusic.core.dir import dirr
from WinxMusic.core.git import git
from WinxMusic.core.userbot import Userbot
from WinxMusic.misc import dbb, heroku

from .logging import LOGGER

# Directories
dirr()

# Check Git Updates
git()

# Initialize Memory DB
dbb()

# Heroku APP
heroku()

app = WinxBot()
userbot = Userbot()

HELPABLE = {}
