import asyncio
from pyrogram import filters
from pyrogram.enums import ChatMembersFilter
from pyrogram.errors import FloodWait

# Impor app sesuai struktur direktori di repositori
from WinxMusic import app

# Daftar untuk menyimpan chat yang sedang menjalankan tag
SPAM_CHATS = []

# Fungsi untuk memeriksa apakah pengguna adalah admin
async def is_admin(chat_id, user_id):
    admin_ids = [
        admin.user.id
        async for admin in app.get_chat_members(
            chat_id, filter=ChatMembersFilter.ADMINISTRATORS
        )
    ]
    return user_id in admin_ids

# Fungsi untuk men-tag semua member di grup
@app.on_message(
    filters.command(["all", "allmention", "mentionall", "tagall"], prefixes=["/", "@"])
)
async def tag_all_users(_, message):
    admin = await is_admin(message.chat.id, message.from_user.id)
    if not admin:
        return

    if message.chat.id in SPAM_CHATS:
        return await message.reply_text(
            "<blockquote><b>Tag all sedang berjalan der, ketik /cancel untuk membatalkan der</b></blockquote>"
        )
    replied = message.reply_to_message
    if len(message.command) < 2 and not replied:
        await message.reply_text(
            "<blockquote><b>Kasih teks nya der\nContoh: /tagall Halo semuanya!</b></blockquote>"
        )
        return
    if replied:
        usernum = 0
        usertxt = ""
        try:
            SPAM_CHATS.append(message.chat.id)
            async for m in app.get_chat_members(message.chat.id):
                if message.chat.id not in SPAM_CHATS:
                    break
                if m.user.is_deleted or m.user.is_bot:
                    continue
                usernum += 1
                usertxt += f"\n<blockquote><b> [{m.user.first_name}](tg://user?id={m.user.id})  </b></blockquote>"
                if usernum == 7:
                    await replied.reply_text(
                        usertxt,
                        disable_web_page_preview=True,
                    )
                    await asyncio.sleep(1)
                    usernum = 0
                    usertxt = ""

            if usernum != 0:
                await replied.reply_text(
                    usertxt,
                    disable_web_page_preview=True,
                )
        except FloodWait as e:
            await asyncio.sleep(e.value)
        try:
            SPAM_CHATS.remove(message.chat.id)
        except Exception:
            pass
    else:
        try:
            usernum = 0
            usertxt = ""
            text = message.text.split(None, 1)[1]
            SPAM_CHATS.append(message.chat.id)
            async for m in app.get_chat_members(message.chat.id):
                if message.chat.id not in SPAM_CHATS:
                    break
                if m.user.is_deleted or m.user.is_bot:
                    continue
                usernum += 1
                usertxt += f"[{m.user.first_name}](tg://user?id={m.user.id})  "
                if usernum == 7:
                    await app.send_message(
                        message.chat.id,
                        f"\n{text}\n{usertxt}",
                        disable_web_page_preview=True,
                    )
                    await asyncio.sleep(2)
                    usernum = 0
                    usertxt = ""
            if usernum != 0:
                await app.send_message(
                    message.chat.id,
                    f"{text}\n\n{usertxt}",
                    disable_web_page_preview=True,
                )
        except FloodWait as e:
            await asyncio.sleep(e.value)
        try:
            SPAM_CHATS.remove(message.chat.id)
        except Exception:
            pass

# Fungsi untuk men-tag semua admin di grup
@app.on_message(
    filters.command(["admin", "admins", "report"], prefixes=["/"]) & filters.group)
async def admintag_with_reporting(client, message):
    if not message.from_user:
        return
    chat_id = message.chat.id
    from_user_id = message.from_user.id
    admins = [
        admin.user.id
        async for admin in client.get_chat_members(
            chat_id, filter=ChatMembersFilter.ADMINISTRATORS
        )
    ]
    if message.command[0] == "report":
        if from_user_id in admins:
            return await message.reply_text(
                "<blockquote><b>Lu itu admin der, ngapain report?\nTindak aja langsung der</b></blockquote>"
            )

    if from_user_id in admins:
        return await tag_all_admins(client, message)

    if len(message.text.split()) <= 1 and not message.reply_to_message:
        return await message.reply_text("Reply to a message to report that user.")

    reply = message.reply_to_message or message
    reply_user_id = reply.from_user.id if reply.from_user else reply.sender_chat.id
    linked_chat = (await client.get_chat(chat_id)).linked_chat
    if reply_user_id == app.id:
        return await message.reply_text("<blockquote><b>Lu report siapa der? Angin?</b></blockquote>")
    if (
        reply_user_id in admins
        or reply_user_id == chat_id
        or (linked_chat and reply_user_id == linked_chat.id)
    ):
        return await message.reply_text(
            "<blockquote><b>Lu tau yang lu report admin der?</b></blockquote>"
        )

    user_mention = reply.from_user.mention if reply.from_user else "the user"
    text = f"<blockquote><b>Bocah ini: {user_mention}\nDilaporkan ke admin</b></blockquote>."

    for admin in admins:
        admin_member = await client.get_chat_member(chat_id, admin)
        if not admin_member.user.is_bot and not admin_member.user.is_deleted:
            text += f"[\u2063](tg://user?id={admin})"

    await reply.reply_text(text)

# Fungsi untuk membatalkan tag all
@app.on_message(
    filters.command(
        [
            "stopmention",
            "cancel",
            "cancelmention",
            "offmention",
            "mentionoff",
            "cancelall",
        ],
        prefixes=["/", "@"],
    )
)
async def cancelcmd(_, message):
    chat_id = message.chat.id
    admin = await is_admin(chat_id, message.from_user.id)
    if not admin:
        return
    if chat_id in SPAM_CHATS:
        try:
            SPAM_CHATS.remove(chat_id)
        except Exception:
            pass
        return await message.reply_text("<blockquote><b>Tag all sukses dihentikan der</b></blockquote>")

    else:
        await message.reply_text("<blockquote><b>Gak ada proses berjalan der</b></blockquote>")
        return

MODULE = "Tagall Extended"
HELP = """
<blockquote><b>Admin Only
/tagall - Tag all semua member grup lu der
/admins - Tag all semua dmin grup der
/report - Report member tengil der [khusus member] 
/cancel - Cancel tag all yang sedang berjalan der</b></blockquote>
"""
