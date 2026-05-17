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
	footerBgURL := ""

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
		if cfg.FooterBgURL != "" {
			footerBgURL = cfg.FooterBgURL
		}
	}

	header := buildEmailHeader(headerImageURL)
	profileLogoURL := restaurantLogoURL
	if profileLogoURL == "" {
		profileLogoURL = logoURL
	}
	profileSection := buildProfileSection(profileLogoURL, restaurantName)
	footer := buildEmailFooter(logoURL, instagramURL, facebookURL, tiktokURL, websiteURL, supportEmail, footerNote, copyright, footerBgURL)

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

func socialIconCell(url, label string) string {
	return fmt.Sprintf(`<td style="padding:0 5px;"><a href="%s" target="_blank" style="display:inline-block;width:40px;height:40px;line-height:40px;border-radius:20px;background:#7c3aed;text-align:center;text-decoration:none;font-size:11px;font-weight:700;color:#ffffff;font-family:Arial,Helvetica,sans-serif;">%s</a></td>`, url, label)
}

func buildEmailFooter(logoURL, instagramURL, facebookURL, tiktokURL, websiteURL, supportEmail, footerNote, copyright, footerBgURL string) string {
	logoHTML := ""
	if logoURL != "" {
		logoHTML = fmt.Sprintf(`<img src="%s" alt="Wayt" style="height:48px;display:block;max-width:160px;" />`, logoURL)
	} else {
		logoHTML = `<span style="font-size:36px;font-weight:900;color:#7c3aed;letter-spacing:-2px;font-family:Arial,Helvetica,sans-serif;">wayt</span>`
	}

	// Social icons — pure CSS circles (works in all email clients without images)
	var socialCells string
	if instagramURL != "" {
		socialCells += socialIconCell(instagramURL, "IG")
	}
	if facebookURL != "" {
		socialCells += socialIconCell(facebookURL, "FB")
	}
	if tiktokURL != "" {
		socialCells += socialIconCell(tiktokURL, "TK")
	}
	socialRow := ""
	if socialCells != "" {
		socialRow = fmt.Sprintf(`<table cellpadding="0" cellspacing="0" align="center" style="margin:0 auto;"><tr>%s</tr></table>`, socialCells)
	}

	// Bottom section: use footer background image if provided, else plain white
	bgAttr := ""
	bgStyle := "background:#ffffff;"
	if footerBgURL != "" {
		bgAttr = fmt.Sprintf(` background="%s"`, footerBgURL)
		bgStyle = fmt.Sprintf(`background:#ffffff url('%s') no-repeat center bottom;background-size:cover;`, footerBgURL)
	}

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
<td%s style="%s padding:24px 32px 20px;">
<table width="100%%" cellpadding="0" cellspacing="0" style="padding-bottom:16px;margin-bottom:12px;border-bottom:1px solid rgba(0,0,0,0.08);">
<tr>
<td valign="middle" style="padding-right:12px;">
<span style="font-size:14px;font-weight:700;color:#7c3aed;font-family:Arial,Helvetica,sans-serif;">Butuh bantuan?</span><br/>
<span style="font-size:13px;color:#374151;font-family:Arial,Helvetica,sans-serif;">Tim Wayt siap membantu Anda.</span>
</td>
<td valign="middle" align="right" style="border-left:1px solid rgba(0,0,0,0.08);padding-left:12px;white-space:nowrap;">
<a href="mailto:%s" style="font-size:13px;color:#374151;text-decoration:none;font-family:Arial,Helvetica,sans-serif;">%s</a>
</td>
</tr>
</table>
<p style="margin:0;font-size:11px;color:#9ca3af;text-align:center;font-family:Arial,Helvetica,sans-serif;line-height:1.6;">%s<br/>%s</p>
</td>
</tr>
</table>
</td></tr>`,
		logoHTML, socialRow, websiteURL,
		bgAttr, bgStyle,
		supportEmail, supportEmail,
		footerNote, copyright)
}
