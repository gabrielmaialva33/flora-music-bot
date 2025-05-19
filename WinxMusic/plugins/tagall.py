import asyncio
import random
from pyrogram import filters
from pyrogram.enums import ChatMembersFilter
from pyrogram.errors import FloodWait

from WinxMusic import app  # Impor app sesuai struktur direktori

SPAM_CHATS = []

EMOJIS = ["🔥", "⚡", "🌟", "🚀", "🎯", "💥", "🎉", "💫", "❤️", "🌀"]
SYMBOLS = ["✦", "➤", "✪", "★", "❖", "✺"]
QUOTES = [
    "   ‌ ⑅         ⑅
    ૮  ˊ ˘ ˋ  ა
‌ପ ‌ /  つ❤️ぅ 
    ⊂、 / 
         U!"
]

async def is_admin(chat_id, user_id):
    admin_ids = [
        admin.user.id
        async for admin in app.get_chat_members(chat_id, filter=ChatMembersFilter.ADMINISTRATORS)
    ]
    return user_id in admin_ids

@app.on_message(filters.command(["all", "tagall"], prefixes=["/", "@"]))
async def tag_all_users(_, message):
    admin = await is_admin(message.chat.id, message.from_user.id)
    if not admin:
        return

    if message.chat.id in SPAM_CHATS:
        return await message.reply_text("<blockquote><b>Tag all sedang berjalan der, ketik /cancel untuk membatalkan der</b></blockquote>", parse_mode="HTML")

    replied = message.reply_to_message
    if len(message.command) < 2 and not replied:
        await message.reply_text("<blockquote><b>Kasih teks nya der\nContoh: /tagall Halo semuanya!</b></blockquote>", parse_mode="HTML")
        return

    try:
        SPAM_CHATS.append(message.chat.id)
        usernum = 0
        usertxt = ""

        async for m in app.get_chat_members(message.chat.id):
            if message.chat.id not in SPAM_CHATS:
                break
            if m.user.is_deleted or m.user.is_bot:
                continue

            emoji = random.choice(EMOJIS)
            symbol = random.choice(SYMBOLS)
            quote = random.choice(QUOTES)

            usernum += 1
            usertxt += f"\n<blockquote><b>{emoji} {quote} {symbol} {m.user.first_name}</b></blockquote>"

            if usernum == 7:
                await replied.reply_text(usertxt, disable_web_page_preview=True, parse_mode="HTML")
                await asyncio.sleep(1)
                usernum = 0
                usertxt = ""

        if usernum != 0:
            await replied.reply_text(usertxt, disable_web_page_preview=True, parse_mode="HTML")
    except FloodWait as e:
        await asyncio.sleep(e.value + 2)  # Tambahkan buffer waktu
    finally:
        SPAM_CHATS.remove(message.chat.id)

@app.on_message(filters.command(["cancel"], prefixes=["/", "@"]))
async def cancelcmd(_, message):
    chat_id = message.chat.id
    admin = await is_admin(chat_id, message.from_user.id)
    if not admin:
        return
    if chat_id in SPAM_CHATS:
        SPAM_CHATS.remove(chat_id)
        return await message.reply_text("<blockquote><b>Tag all sukses dihentikan der</b></blockquote>", parse_mode="HTML")
    else:
        await message.reply_text("<blockquote><b>Gak ada proses berjalan der</b></blockquote>", parse_mode="HTML")
