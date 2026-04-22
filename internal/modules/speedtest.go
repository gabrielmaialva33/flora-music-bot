package modules

import (
	"fmt"

	"github.com/amarnathcjd/gogram/telegram"
	"github.com/showwin/speedtest-go/speedtest"

	"main/internal/locales"
	"main/internal/utils"
)

func init() {
	helpTexts["/speedtest"] = `<i>Roda um teste de velocidade de rede no servidor.</i>

<u>Uso:</u>
<b>/speedtest</b> ou <b>/spt</b> — Testa a velocidade da rede

<b>📊 Resultados incluem:</b>
• Velocidade de download (Mbps)
• Velocidade de upload (Mbps)
• Localização do servidor
• Latência (ms)
• Info do ISP

<b>🔒 Restrições:</b>
• Apenas <b>sudoers</b>

<b>⚠️ Observação:</b>
O teste pode levar de 30-60 segundos pra terminar.`
}

func sptHandle(m *telegram.NewMessage) error {
	chatID := m.ChannelID()

	user, err := speedtest.FetchUserInfo()
	if err != nil {
		m.Reply(F(chatID, "spt_fetch_fail", locales.Arg{
			"error": err,
		}))
		return nil
	}

	servers, err := speedtest.FetchServers()
	if err != nil {
		m.Reply(F(chatID, "spt_servers_fetch_fail", locales.Arg{
			"error": err,
		}))
		return nil
	}

	best, err := servers.FindServer([]int{})
	if err != nil || len(best) == 0 {
		m.Reply(F(chatID, "spt_best_server_fail", locales.Arg{
			"error": err,
		}))
		return nil
	}
	server := best[0]

	statusMsg, err := m.Reply(F(chatID, "spt_running_download"))
	if err != nil {
		return err
	}

	server.DownloadTest()

	utils.EOR(statusMsg, F(chatID, "spt_running_upload"))
	server.UploadTest()

	output := F(chatID, "spt_result", locales.Arg{
		"ip":          user.IP,
		"isp":         user.Isp,
		"lat":         user.Lat,
		"lon":         user.Lon,
		"server_name": server.Name,
		"country":     server.Country,
		"sponsor":     server.Sponsor,
		"distance_km": fmt.Sprintf("%.2f", server.Distance),
		"latency_ms": fmt.Sprintf(
			"%.2f",
			float64(server.Latency.Microseconds())/1000,
		),
		"dl_mbps": fmt.Sprintf("%.2f", server.DLSpeed/1024/1024),
		"ul_mbps": fmt.Sprintf("%.2f", server.ULSpeed/1024/1024),
	})

	utils.EOR(statusMsg, output)
	return nil
}
