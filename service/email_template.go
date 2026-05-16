package service

import (
	"fmt"

	"github.com/wayt-app/wayt-core/model"
)

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

func buildEmailFooter(logoURL, instagramURL, facebookURL, tiktokURL, websiteURL, supportEmail, footerNote, copyright string) string {
	logoHTML := ""
	if logoURL != "" {
		logoHTML = fmt.Sprintf(`<img src="%s" alt="Wayt" style="height:36px;display:block;" />`, logoURL)
	} else {
		logoHTML = `<span style="font-size:22px;font-weight:900;color:#7c3aed;letter-spacing:-1px;">wayt</span>`
	}

	socialHTML := ""
	if instagramURL != "" {
		socialHTML += fmt.Sprintf(`<a href="%s" style="display:inline-block;width:36px;height:36px;border-radius:50%%;background:#7c3aed;margin:0 4px;text-align:center;line-height:36px;color:#fff;text-decoration:none;font-size:14px;">ig</a>`, instagramURL)
	}
	if facebookURL != "" {
		socialHTML += fmt.Sprintf(`<a href="%s" style="display:inline-block;width:36px;height:36px;border-radius:50%%;background:#7c3aed;margin:0 4px;text-align:center;line-height:36px;color:#fff;text-decoration:none;font-size:14px;">fb</a>`, facebookURL)
	}
	if tiktokURL != "" {
		socialHTML += fmt.Sprintf(`<a href="%s" style="display:inline-block;width:36px;height:36px;border-radius:50%%;background:#7c3aed;margin:0 4px;text-align:center;line-height:36px;color:#fff;text-decoration:none;font-size:14px;">tt</a>`, tiktokURL)
	}

	socialSection := ""
	if socialHTML != "" {
		socialSection = fmt.Sprintf(`<p style="margin:0 0 12px;">%s</p>`, socialHTML)
	}

	return fmt.Sprintf(`<tr><td style="background:#f9fafb;padding:28px 40px;border-top:1px solid #e5e7eb;">
<table width="100%%" cellpadding="0" cellspacing="0">
<tr>
<td width="40%%" valign="top">
%s
<p style="margin:8px 0 0;font-size:12px;color:#6b7280;">Reservasi lebih mudah,<br/>antrean jadi lebih baik.</p>
</td>
<td width="30%%" valign="top" align="center">
<p style="margin:0 0 8px;font-size:12px;font-weight:700;color:#374151;">Temukan kami</p>
%s
</td>
<td width="30%%" valign="top" align="right">
<a href="%s" style="display:inline-block;background:#7c3aed;color:#fff;text-decoration:none;font-size:12px;font-weight:600;padding:8px 16px;border-radius:8px;">Kunjungi %s →</a>
</td>
</tr>
</table>
<table width="100%%" cellpadding="0" cellspacing="0" style="margin-top:20px;padding-top:20px;border-top:1px solid #e5e7eb;">
<tr>
<td valign="middle">
<span style="font-size:12px;color:#6b7280;">🎧 Butuh bantuan? Tim Wayt siap membantu.</span>
</td>
<td align="right" valign="middle">
<a href="mailto:%s" style="font-size:12px;color:#7c3aed;text-decoration:none;">✉ %s</a>
</td>
</tr>
</table>
<p style="margin:16px 0 0;font-size:11px;color:#9ca3af;text-align:center;">%s<br/>%s</p>
</td></tr>`, logoHTML, socialSection, websiteURL, websiteURL, supportEmail, supportEmail, footerNote, copyright)
}
