package service

import (
	"encoding/base64"
	"fmt"

	"github.com/wayt-app/wayt-core/model"
)

func svgDataURI(svg string) string {
	return "data:image/svg+xml;base64," + base64.StdEncoding.EncodeToString([]byte(svg))
}

// wrapEmailHTML wraps an email body fragment with branded header and footer.
func wrapEmailHTML(body string, cfg *model.EmailConfig, recipientName, restaurantName, restaurantLogoURL string) string {
	// Defaults when config is nil or fields are empty
	headerImageURL := ""
	logoURL := ""
	instagramURL := ""
	facebookURL := ""
	tiktokURL := ""
	websiteURL := "https://wayt.fun"
	supportEmail := "support@wayt.fun"
	footerNote := "Email ini dikirim secara otomatis, mohon tidak membalas email ini."
	copyright := "© 2026 Wayt. All rights reserved."

	if cfg != nil {
		if cfg.HeaderImageURL != "" {
			headerImageURL = cfg.HeaderImageURL
		}
		if cfg.LogoURL != "" {
			logoURL = cfg.LogoURL
		}
		if cfg.InstagramURL != "" {
			instagramURL = cfg.InstagramURL
		}
		if cfg.FacebookURL != "" {
			facebookURL = cfg.FacebookURL
		}
		if cfg.TiktokURL != "" {
			tiktokURL = cfg.TiktokURL
		}
		if cfg.WebsiteURL != "" {
			websiteURL = cfg.WebsiteURL
		}
		if cfg.SupportEmail != "" {
			supportEmail = cfg.SupportEmail
		}
		if cfg.FooterNote != "" {
			footerNote = cfg.FooterNote
		}
		if cfg.Copyright != "" {
			copyright = cfg.Copyright
		}
	}

	header := buildEmailHeader(headerImageURL)
	profileLogoURL := restaurantLogoURL
	if profileLogoURL == "" {
		profileLogoURL = logoURL
	}
	profileSection := buildProfileSection(profileLogoURL, restaurantName)
	footer := buildEmailFooter(logoURL, instagramURL, facebookURL, tiktokURL, websiteURL, supportEmail, footerNote, copyright)

	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="id">
<head>
<meta charset="UTF-8"/>
<meta name="viewport" content="width=device-width,initial-scale=1.0"/>
<title>Wayt</title>
</head>
<body style="margin:0;padding:0;background:#f5f5f5;font-family:Arial,Helvetica,sans-serif;">
<table width="100%%" cellpadding="0" cellspacing="0" style="background:#f5f5f5;">
<tr><td align="center" style="padding:20px 16px;">
<table width="600" cellpadding="0" cellspacing="0" style="max-width:600px;width:100%%;background:#ffffff;border-radius:16px;overflow:hidden;box-shadow:0 2px 8px rgba(0,0,0,0.08);">
%s
%s
<tr><td style="padding:32px 40px;">%s</td></tr>
%s
</table>
</td></tr>
</table>
</body>
</html>`, header, profileSection, body, footer)
}

func buildEmailHeader(headerImageURL string) string {
	if headerImageURL != "" {
		return fmt.Sprintf(`<tr><td style="padding:0;">
<img src="%s" alt="Wayt" style="width:100%%;display:block;border-radius:16px 16px 0 0;" />
</td></tr>`, headerImageURL)
	}
	// CSS fallback header
	return `<tr><td style="background:linear-gradient(135deg,#7c3aed 0%%,#6d28d9 100%%);padding:32px 40px;border-radius:16px 16px 0 0;">
<table width="100%%" cellpadding="0" cellspacing="0">
<tr>
<td><span style="font-size:28px;font-weight:900;color:#ffffff;letter-spacing:-1px;">wayt</span></td>
<td align="right"><span style="font-size:13px;color:rgba(255,255,255,0.85);">Reservasi lebih mudah,<br/>antrean jadi lebih baik.</span></td>
</tr>
</table>
</td></tr>`
}

func buildProfileSection(restaurantLogoURL, restaurantName string) string {
	if restaurantLogoURL == "" && restaurantName == "" {
		return ""
	}
	logoHTML := ""
	if restaurantLogoURL != "" {
		logoHTML = fmt.Sprintf(`<img src="%s" alt="%s" style="width:56px;height:56px;border-radius:12px;object-fit:cover;border:2px solid #ede9fe;display:block;" />`, restaurantLogoURL, restaurantName)
	} else {
		initial := "R"
		if len(restaurantName) > 0 {
			initial = string([]rune(restaurantName)[0:1])
		}
		logoHTML = fmt.Sprintf(`<div style="width:56px;height:56px;border-radius:12px;background:#ede9fe;display:flex;align-items:center;justify-content:center;font-size:24px;font-weight:bold;color:#7c3aed;">%s</div>`, initial)
	}
	return fmt.Sprintf(`<tr><td style="padding:24px 40px 0;border-bottom:1px solid #f3f4f6;">
<table cellpadding="0" cellspacing="0">
<tr>
<td style="padding-right:12px;">%s</td>
<td><span style="font-size:15px;font-weight:700;color:#1f2937;">%s</span></td>
</tr>
</table>
</td></tr>`, logoHTML, restaurantName)
}

func socialIconCell(url, iconDataURI string) string {
	return fmt.Sprintf(`<td style="padding:0 5px;"><a href="%s" target="_blank" style="display:block;text-decoration:none;"><img src="%s" width="40" height="40" alt="" style="display:block;border:0;" /></a></td>`, url, iconDataURI)
}

func buildEmailFooter(logoURL, instagramURL, facebookURL, tiktokURL, websiteURL, supportEmail, footerNote, copyright string) string {
	logoHTML := ""
	if logoURL != "" {
		logoHTML = fmt.Sprintf(`<img src="%s" alt="Wayt" style="height:48px;display:block;max-width:160px;" />`, logoURL)
	} else {
		logoHTML = `<span style="font-size:36px;font-weight:900;color:#7c3aed;letter-spacing:-2px;font-family:Arial,Helvetica,sans-serif;">wayt</span>`
	}

	// Social icons — encode as base64 data URI so Gmail renders them (inline SVG is stripped by Gmail)
	igURI := svgDataURI(`<svg xmlns="http://www.w3.org/2000/svg" width="40" height="40" viewBox="0 0 40 40"><circle cx="20" cy="20" r="20" fill="#7c3aed"/><rect x="11" y="11" width="18" height="18" rx="5" ry="5" fill="none" stroke="white" stroke-width="1.5"/><circle cx="20" cy="20" r="4.5" fill="none" stroke="white" stroke-width="1.5"/><circle cx="26" cy="14" r="1.5" fill="white"/></svg>`)
	fbURI := svgDataURI(`<svg xmlns="http://www.w3.org/2000/svg" width="40" height="40" viewBox="0 0 40 40"><circle cx="20" cy="20" r="20" fill="#7c3aed"/><path d="M24 11h-2.5C18.9 11 17 12.9 17 15.5V18h-2.5v3.5H17V29h4v-7.5h2.5l.5-3.5H21v-2.5c0-.8.7-1.5 1.5-1.5H24V11z" fill="white"/></svg>`)
	ttURI := svgDataURI(`<svg xmlns="http://www.w3.org/2000/svg" width="40" height="40" viewBox="0 0 40 40"><circle cx="20" cy="20" r="20" fill="#7c3aed"/><path d="M28 15.5c-1.5-.8-2.5-2.2-2.9-3.8H22.7V24a2.5 2.5 0 0 1-2.4 2.3 2.5 2.5 0 0 1-2.5-2.5 2.5 2.5 0 0 1 2.5-2.4c.2 0 .5 0 .7.1v-2.6c-.2 0-.5-.1-.7-.1A5 5 0 0 0 15.3 24a5 5 0 0 0 5 5 5 5 0 0 0 5-5V19.6a9.2 9.2 0 0 0 3.3 1.1V18c-.4 0-1-.2-.6-.5z" fill="white"/></svg>`)

	var socialCells string
	if instagramURL != "" {
		socialCells += socialIconCell(instagramURL, igURI)
	}
	if facebookURL != "" {
		socialCells += socialIconCell(facebookURL, fbURI)
	}
	if tiktokURL != "" {
		socialCells += socialIconCell(tiktokURL, ttURI)
	}
	socialRow := ""
	if socialCells != "" {
		socialRow = fmt.Sprintf(`<table cellpadding="0" cellspacing="0" align="center" style="margin:0 auto;"><tr>%s</tr></table>`, socialCells)
	}

	// Wave & heart decorations — base64 data URI so Gmail renders them
	waveURI := svgDataURI(`<svg xmlns="http://www.w3.org/2000/svg" width="150" height="130" viewBox="0 0 150 130"><path d="M0 130 L0 75 Q35 20 70 50 Q105 80 135 35 Q143 18 150 22 L150 130 Z" fill="#7c3aed" opacity="0.13"/><path d="M0 130 L0 95 Q40 60 75 75 Q110 90 140 65 L150 60 L150 130 Z" fill="#7c3aed" opacity="0.09"/></svg>`)
	heartURI := svgDataURI(`<svg xmlns="http://www.w3.org/2000/svg" width="95" height="105" viewBox="0 0 95 105"><rect x="12" y="18" width="66" height="66" rx="15" fill="#5b21b6" opacity="0.3"/><rect x="6" y="10" width="66" height="66" rx="15" fill="#7c3aed"/><path d="M39 56C39 56 20 44 20 31.5 20 24.5 25.5 19 32.5 19 36.5 19 39 21.5 39 21.5 41.5 19 45.5 19 52.5 19 58 24.5 58 31.5 58 44 39 56 39 56Z" fill="white"/><circle cx="80" cy="16" r="3.5" fill="#a78bfa"/><circle cx="87" cy="28" r="2" fill="#c4b5fd"/><circle cx="74" cy="8" r="2.5" fill="#ddd6fe"/><circle cx="90" cy="14" r="1.5" fill="#ede9fe"/></svg>`)

	return fmt.Sprintf(`<tr><td style="background:#ffffff;padding:0;border-top:1px solid #e5e7eb;">
<table width="100%%" cellpadding="0" cellspacing="0">
<tr>
<td valign="top" style="padding:28px 20px 28px 32px;border-right:1px solid #e5e7eb;">
%s
<p style="margin:12px 0 0;font-size:14px;color:#374151;line-height:1.7;font-family:Arial,Helvetica,sans-serif;">Reservasi lebih mudah,<br/>antrean jadi lebih baik.</p>
</td>
<td align="center" valign="top" style="padding:28px 16px;border-right:1px solid #e5e7eb;">
<p style="margin:0 0 14px;font-size:13px;font-weight:700;color:#111827;font-family:Arial,Helvetica,sans-serif;">Temukan kami</p>
%s
</td>
<td valign="top" style="padding:28px 32px 28px 20px;">
<p style="margin:0 0 12px;font-size:13px;font-weight:700;color:#111827;font-family:Arial,Helvetica,sans-serif;">Info lebih lanjut tentang Wayt</p>
<a href="%s" style="display:block;background:#7c3aed;color:#fff;text-decoration:none;font-size:13px;font-weight:600;padding:12px 16px;border-radius:10px;text-align:center;font-family:Arial,Helvetica,sans-serif;">Kunjungi web ke wayt.fun &#8594;</a>
</td>
</tr>
</table>
<table width="100%%" cellpadding="0" cellspacing="0" style="border-top:1px solid #e5e7eb;">
<tr>
<td width="150" valign="bottom" style="padding:0;line-height:0;font-size:0;vertical-align:bottom;"><img src="%s" width="150" height="130" alt="" style="display:block;border:0;" /></td>
<td valign="top" style="padding:20px 8px 0;">
<table width="100%%" cellpadding="0" cellspacing="0" style="padding-bottom:16px;margin-bottom:12px;border-bottom:1px solid #e5e7eb;">
<tr>
<td valign="middle" style="padding-right:12px;">
<span style="font-size:14px;font-weight:700;color:#7c3aed;font-family:Arial,Helvetica,sans-serif;">Butuh bantuan?</span><br/>
<span style="font-size:13px;color:#374151;font-family:Arial,Helvetica,sans-serif;">Tim Wayt siap membantu Anda.</span>
</td>
<td valign="middle" align="right" style="border-left:1px solid #e5e7eb;padding-left:12px;white-space:nowrap;">
<a href="mailto:%s" style="font-size:13px;color:#374151;text-decoration:none;font-family:Arial,Helvetica,sans-serif;">%s</a>
</td>
</tr>
</table>
<p style="margin:0 0 8px;font-size:11px;color:#9ca3af;text-align:center;font-family:Arial,Helvetica,sans-serif;line-height:1.6;">%s<br/>%s</p>
</td>
<td width="95" valign="bottom" style="padding:0;line-height:0;font-size:0;vertical-align:bottom;"><img src="%s" width="95" height="105" alt="" style="display:block;border:0;margin-left:auto;" /></td>
</tr>
</table>
</td></tr>`,
		logoHTML, socialRow, websiteURL,
		waveURI,
		supportEmail, supportEmail,
		footerNote, copyright,
		heartURI)
}
