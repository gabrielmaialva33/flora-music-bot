package modules

import (
	"fmt"

	tg "github.com/amarnathcjd/gogram/telegram"

	"main/internal/config"
)

func init() {
	helpTexts["/privacy"] = `<i>Mostra a política de privacidade do bot.</i>`
}

func privacyHandler(m *tg.NewMessage) error {
	privacyText := fmt.Sprintf(`<b>🛡 Política de Privacidade &amp; Tratamento de Dados</b>

Sua privacidade é importante pra gente. Esse bot foi feito com privacidade em mente.

<b>📊 Dados que coletamos</b>
<blockquote>Só guardamos o essencial pro bot funcionar:
• <b>IDs de usuário e chat:</b> Para identificar grupos e gerenciar configurações.
• <b>Preferências:</b> Idioma e configurações do bot.
• <b>Controle de acesso:</b> Lista de usuários autorizados do seu grupo.
• <b>Config RTMP:</b> Apenas se você usar o recurso de streaming RTMP.</blockquote>

<b>📩 Privacidade das mensagens</b>
<blockquote>• O bot <b>só</b> lê mensagens que começam com um comando (ex.: <code>/play</code>) ou interações com os próprios botões.
• Ele <b>não</b> lê, guarda ou monitora suas conversas privadas ou mensagens gerais do grupo.</blockquote>

<b>🌐 Serviços de terceiros</b>
<blockquote>• Usamos serviços externos como <b>YouTube</b> e <b>Spotify</b> só para buscar e transmitir a música que você pede.
• Nenhum dado pessoal é compartilhado com esses serviços além da própria busca.</blockquote>

<b>🤝 Compartilhamento de dados</b>
<blockquote>• <b>Nunca</b> vendemos, compartilhamos ou trocamos seus dados com terceiros.
• Todos os dados são usados estritamente para fornecer e melhorar os recursos de streaming de música.</blockquote>

<b>✨ Nosso compromisso</b>
Esse bot é um <a href="https://github.com/TheTeamVivek/YukkiMusic">projeto open-source</a> dedicado a oferecer uma experiência de streaming de alta qualidade respeitando a privacidade do usuário.

<i>Se tiver qualquer dúvida, entra no nosso <a href="%s">Chat de Suporte</a>.</i>`, config.SupportChat)

	_, err := m.Reply(privacyText)
	return err
}
