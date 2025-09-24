import asyncio
import importlib.util
import os
import traceback
from datetime import datetime
from functools import wraps

from pyrogram import Client, StopPropagation, errors
from pyrogram.enums import ChatMemberStatus
from pyrogram.errors import (
    ChatSendMediaForbidden,
    ChatSendPhotosForbidden,
    ChatWriteForbidden,
    FloodWait,
    MessageIdInvalid,
    MessageNotModified,
)
from pyrogram.handlers import MessageHandler
from pyrogram.types import (
    BotCommand,
    BotCommandScopeAllChatAdministrators,
    BotCommandScopeAllGroupChats,
    BotCommandScopeAllPrivateChats,
    BotCommandScopeChat,
    BotCommandScopeChatMember,
)

import config
from ..logging import LOGGER


class WinxBot(Client):
    def __init__(self, *args, **kwargs):
        LOGGER(__name__).info("Starting Bot...")

        super().__init__(
            "WinxMusic",
            api_id=config.API_ID,
            api_hash=config.API_HASH,
            bot_token=config.BOT_TOKEN,
            sleep_threshold=240,
            max_concurrent_transmissions=5,
            workers=50,
        )
        self.loaded_plug_counts = 0

    def on_message(self, filters=None, group=0):
        def decorator(func):
            @wraps(func)
            async def wrapper(client, message):
                try:
                    if asyncio.iscoroutinefunction(func):
                        await func(client, message)
                    else:
                        func(client, message)
                except FloodWait as e:
                    LOGGER(__name__).warning(
                        f"FloodWait: Sleeping for {e.value} seconds."
                    )
                    await asyncio.sleep(e.value)
                except (
                        ChatWriteForbidden,
                        ChatSendMediaForbidden,
                        ChatSendPhotosForbidden,
                        MessageNotModified,
                        MessageIdInvalid,
                ):
                    pass
                except StopPropagation:
                    raise
                except Exception as e:
                    date_time = datetime.now().strftime("%Y-%m-%d %H:%M:%S")
                    user_id = message.from_user.id if message.from_user else "Unknown"
                    chat_id = message.chat.id if message.chat else "Unknown"
                    chat_username = (
                        f"@{message.chat.username}"
                        if message.chat.username
                        else "Private Group"
                    )
                    command = message.text
                    error_trace = traceback.format_exc()
                    error_message = (
                        f"<b>Error:</b> {type(e).__name__}\n"
                        f"<b>Date:</b> {date_time}\n"
                        f"<b>Chat ID:</b> {chat_id}\n"
                        f"<b>Chat Username:</b> {chat_username}\n"
                        f"<b>User ID:</b> {user_id}\n"
                        f"<b>Command/Text:</b>\n<pre language='python'><code>{command}</code></pre>\n\n"
                        f"<b>Traceback:</b>\n<pre language='python'><code>{error_trace}</code></pre>"
                    )
                    await self.send_message(config.LOG_GROUP_ID, error_message)
                    try:
                        await self.send_message(config.OWNER_ID[0], error_message)
                    except Exception:
                        pass

            handler = MessageHandler(wrapper, filters)
            self.add_handler(handler, group)
            return func

        return decorator

    async def start(self):
        await super().start()
        get_me = await self.get_me()
        self.username = get_me.username
        self.id = get_me.id
        self.name = get_me.full_name
        self.mention = get_me.mention

        try:
            await self.send_message(
                config.LOG_GROUP_ID,
                text=(
                    f"<u><b>{self.mention} Bot Started :</b></u>\n\n"
                    f"Id : <code>{self.id}</code>\n"
                    f"Name : {self.name}\n"
                    f"Username : @{self.username}"
                ),
            )
        except (errors.ChannelInvalid, errors.PeerIdInvalid):
            LOGGER(__name__).error(
                "Bot failed to access the log group. Ensure the bot is added and promoted as admin."
            )
            LOGGER(__name__).error("Error details:", exc_info=True)
            exit()
        if config.SET_CMDS:
            try:
                await self._set_default_commands()
            except Exception:
                LOGGER(__name__).warning("Failed to set commands:", exc_info=True)

        try:
            a = await self.get_chat_member(config.LOG_GROUP_ID, "me")
            if a.status != ChatMemberStatus.ADMINISTRATOR:
                LOGGER(__name__).error("Please promote bot as admin in logger group")
                exit()
        except Exception:
            pass
        LOGGER(__name__).info(f"MusicBot started as {self.name}")

    async def _set_default_commands(self):
        private_commands = [
            BotCommand("start", "Iniciar o bot"),
            BotCommand("help", "Obter o menu de ajuda"),
            BotCommand("ping", "Verificar se o bot está ativo ou inativo"),
        ]
        group_commands = [BotCommand("play", "Começar a tocar a música solicitada")]
        admin_commands = [
            BotCommand("play", "Começar a tocar a música solicitada"),
            BotCommand("skip", "Ir para a próxima música na fila"),
            BotCommand("pause", "Pausar a música atual"),
            BotCommand("resume", "Retomar a música pausada"),
            BotCommand("end", "Limpar a fila e sair do chat de voz"),
            BotCommand("shuffle", "Embaralhar aleatoriamente a playlist na fila"),
            BotCommand("playmode", "Alterar o modo de reprodução padrão do seu chat"),
            BotCommand("settings", "Abrir as configurações do bot para o seu chat"),
        ]
        owner_commands = [
            BotCommand("update", "Atualizar o bot"),
            BotCommand("restart", "Reiniciar o bot"),
            BotCommand("logs", "Obter os registros"),
            BotCommand("export", "Exportar todos os dados do MongoDB"),
            BotCommand("import", "Importar todos os dados no MongoDB"),
            BotCommand("addsudo", "Adicionar um usuário como sudoer"),
            BotCommand("delsudo", "Remover um usuário dos sudoers"),
            BotCommand("sudolist", "Listar todos os usuários sudo"),
            BotCommand("log", "Obter os registros do bot"),
            BotCommand("getvar", "Obter uma variável de ambiente específica"),
            BotCommand("delvar", "Excluir uma variável de ambiente específica"),
            BotCommand("setvar", "Definir uma variável de ambiente específica"),
            BotCommand("usage", "Obter informações sobre o uso do Dyno"),
            BotCommand("maintenance", "Ativar ou desativar o modo de manutenção"),
            BotCommand("logger", "Ativar ou desativar o registro de atividades"),
            BotCommand("block", "Bloquear um usuário"),
            BotCommand("unblock", "Desbloquear um usuário"),
            BotCommand("blacklist", "Adicionar um chat à lista negra"),
            BotCommand("whitelist", "Remover um chat da lista negra"),
            BotCommand("blacklisted", "Listar todos os chats na lista negra"),
            BotCommand(
                "autoend", "Ativar ou desativar o término automático para transmissões"
            ),
            BotCommand("reboot", "Reiniciar o bot"),
            BotCommand("restart", "Reiniciar o bot"),
        ]

        await self.set_bot_commands(
            private_commands, scope=BotCommandScopeAllPrivateChats()
        )
        await self.set_bot_commands(
            group_commands, scope=BotCommandScopeAllGroupChats()
        )
        await self.set_bot_commands(
            admin_commands, scope=BotCommandScopeAllChatAdministrators()
        )

        LOG_GROUP_ID = (
            f"@{config.LOG_GROUP_ID}"
            if isinstance(config.LOG_GROUP_ID, str)
               and not config.LOG_GROUP_ID.startswith("@")
            else config.LOG_GROUP_ID
        )

        for owner_id in config.OWNER_ID:
            try:
                await self.set_bot_commands(
                    owner_commands,
                    scope=BotCommandScopeChatMember(
                        chat_id=LOG_GROUP_ID, user_id=owner_id
                    ),
                )
                await self.set_bot_commands(
                    private_commands + owner_commands,
                    scope=BotCommandScopeChat(chat_id=owner_id),
                )
            except Exception:
                pass

    def load_plugin(self, file_path: str, base_dir: str, utils=None):
        file_name = os.path.basename(file_path)
        module_name, ext = os.path.splitext(file_name)
        if module_name.startswith("__") or ext != ".py":
            return None

        relative_path = os.path.relpath(file_path, base_dir).replace(os.sep, ".")
        module_path = f"{os.path.basename(base_dir)}.{relative_path[:-3]}"

        spec = importlib.util.spec_from_file_location(module_path, file_path)
        module = importlib.util.module_from_spec(spec)
        module.logger = LOGGER(module_path)
        module.app = self
        module.Config = config

        if utils:
            module.utils = utils

        try:
            spec.loader.exec_module(module)
            self.loaded_plug_counts += 1
        except Exception as e:
            LOGGER(__name__).error(
                f"Failed to load {module_path}: {e}\n\n", exc_info=True
            )
            exit()

        return module

    def load_plugins_from(self, base_folder: str):
        base_dir = os.path.abspath(base_folder)
        utils_path = os.path.join(base_dir, "utils.py")
        utils = None

        if os.path.exists(utils_path) and os.path.isfile(utils_path):
            try:
                spec = importlib.util.spec_from_file_location("utils", utils_path)
                utils = importlib.util.module_from_spec(spec)
                spec.loader.exec_module(utils)
            except Exception as e:
                LOGGER(__name__).error(
                    f"Failed to load 'utils' module: {e}", exc_info=True
                )

        for root, _, files in os.walk(base_dir):
            for file in files:
                if file.endswith(".py") and not file == "utils.py":
                    file_path = os.path.join(root, file)
                    mod = self.load_plugin(file_path, base_dir, utils)
                    yield mod

    async def run_shell_command(self, command: list):
        process = await asyncio.create_subprocess_exec(
            *command,
            stdout=asyncio.subprocess.PIPE,
            stderr=asyncio.subprocess.PIPE,
        )

        stdout, stderr = await process.communicate()

        return {
            "returncode": process.returncode,
            "stdout": stdout.decode().strip() if stdout else None,
            "stderr": stderr.decode().strip() if stderr else None,
        }
