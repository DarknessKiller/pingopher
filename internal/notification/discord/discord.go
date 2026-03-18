package discord

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/DarknessKiller/pingopher/internal/model"
	"github.com/DarknessKiller/pingopher/internal/util"
)

type DiscordNotification struct {
	WebhookURL    string
	Username      string
	PrefixMessage string
	DisableURL    bool
	ChannelType   string
	ThreadID      string
	PostName      string
}

const (
	colorDown = 0xFF0000
	colorUp   = 0x00FF00
	avatarURL = ""
)

type discordEmbedField struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline,omitempty"`
}

type discordEmbed struct {
	Title     string              `json:"title"`
	Color     int                 `json:"color"`
	Timestamp string              `json:"timestamp"`
	Fields    []discordEmbedField `json:"fields,omitempty"`
}

type discordPayload struct {
	Username   string         `json:"username,omitempty"`
	Content    string         `json:"content,omitempty"`
	AvatarURL  string         `json:"avatar_url,omitempty"`
	ThreadName string         `json:"thread_name,omitempty"`
	Embeds     []discordEmbed `json:"embeds,omitempty"`
}

func (d DiscordNotification) Send(host *model.Host, histories []*model.History) error {
	if len(histories) == 0 {
		return fmt.Errorf("no history provided")
	}

	isUp := host.Status == model.HostStatusUp
	color := colorUp
	statusText := "is UP!"
	if !isUp {
		color = colorDown
		statusText = "went DOWN"
	}

	title := fmt.Sprintf("%s %s", host.Name, statusText)

	// Pre-calculate capacity: base fields + per-history fields + separators
	capacity := 3 + len(histories)*(5)

	if !d.DisableURL {
		capacity++
	}

	fields := make([]discordEmbedField, 0, capacity)

	fields = append(fields, discordEmbedField{Name: "Host", Value: host.Name, Inline: true})

	if !d.DisableURL {
		url := host.Protocol + "://" + host.HostURL
		fields = append(fields, discordEmbedField{Name: "URL", Value: url, Inline: true})
	}

	if len(fields)%3 == 2 {
		fields = append(fields, discordEmbedField{Name: "\u200B", Value: "\u200B", Inline: true})
	}

	for i, host := range histories {
		if i > 0 {
			fields = append(fields, discordEmbedField{Value: "\u200B", Inline: false})
		}

		t := host.PingDateTime.Time.In(time.Local).Format("02/01/2006 15:04:05 (MST)")

		dnsAddr := host.DNS.IP
		if host.DNS.Port != 0 {
			dnsAddr += ":" + strconv.Itoa(int(host.DNS.Port))
		}

		fields = append(fields,
			discordEmbedField{Name: "Status Code", Value: strconv.Itoa(int(host.StatusCode)), Inline: false},
			discordEmbedField{Name: "DNS · " + host.DNS.Name, Value: t, Inline: true},
		)

		if dnsAddr != "" {
			fields = append(fields, discordEmbedField{Name: "IP", Value: dnsAddr, Inline: true})
		}

		if isUp {
			fields = append(fields, discordEmbedField{
				Name: "Latency", Value: strconv.Itoa(int(host.Latency)) + " ms", Inline: true,
			})
		} else {
			errMsg := host.ErrorMessage
			if len(errMsg) > 1000 {
				errMsg = errMsg[:1000] + "... (truncated)"
			}
			fields = append(fields, discordEmbedField{
				Name:   "Error",
				Value:  "```" + errMsg + "```",
				Inline: false,
			})
		}
	}

	embed := discordEmbed{
		Title:     title,
		Color:     color,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Fields:    fields,
	}

	payload := discordPayload{
		Username:  getUsername(d.Username),
		Content:   strings.TrimSpace(d.PrefixMessage),
		AvatarURL: avatarURL,
		Embeds:    []discordEmbed{embed},
	}

	if d.ChannelType == "createNewForumPost" && d.PostName != "" {
		payload.ThreadName = d.PostName
	}

	url := d.WebhookURL
	if d.ChannelType == "postToThread" && d.ThreadID != "" {
		url = addQueryParam(url, "thread_id", d.ThreadID)
	}

	return util.SendJSONWebhook(url, payload)
}

func getUsername(custom string) string {
	if s := strings.TrimSpace(custom); s != "" {
		return s
	}
	return "Pingopher"
}

func addQueryParam(url, key, value string) string {
	if strings.Contains(url, "?") {
		return url + "&" + key + "=" + value
	}
	return url + "?" + key + "=" + value
}

func SendSystemErrorWebhook(webhookURL, title, errorMessage string) error {
	embed := discordEmbed{
		Title:     title,
		Color:     colorDown,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Fields: []discordEmbedField{
			{Name: "Error", Value: "```" + errorMessage + "```", Inline: false},
		},
	}

	payload := discordPayload{
		Username: "Pingopher System",
		Embeds:   []discordEmbed{embed},
	}

	return util.SendJSONWebhook(webhookURL, payload)
}
