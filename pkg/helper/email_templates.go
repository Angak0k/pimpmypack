package helper

import (
	"fmt"
	"html"
	"strings"
)

// BuildEmailHTML wraps body content in a branded HTML email layout.
// title and preheaderText are HTML-escaped; bodyContent is trusted pre-built HTML.
func BuildEmailHTML(title, preheaderText, bodyContent string) string {
	safeTitle := html.EscapeString(title)
	safePreheader := html.EscapeString(preheaderText)
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en" xmlns="http://www.w3.org/1999/xhtml">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<meta http-equiv="X-UA-Compatible" content="IE=edge">
<title>%s</title>
<!--[if mso]>
<style type="text/css">
table {border-collapse: collapse;}
</style>
<![endif]-->
</head>
<body style="margin:0;padding:0;background-color:#f5f5f4;
font-family:'Segoe UI',Roboto,'Helvetica Neue',Arial,sans-serif;">
<!-- Preheader (hidden) -->
<div style="display:none;max-height:0;overflow:hidden;">%s</div>
<!-- Wrapper -->
<table role="presentation" width="100%%" cellpadding="0" cellspacing="0" border="0" style="background-color:#f5f5f4;">
<tr><td align="center" style="padding:24px 16px;">
<!-- Container 600px -->
<table role="presentation" width="600" cellpadding="0" cellspacing="0" border="0" style="max-width:600px;width:100%%;">
<!-- Header -->
<tr><td style="background-color:#14532d;padding:24px 32px;border-radius:12px 12px 0 0;text-align:center;">
<span style="font-size:24px;font-weight:700;color:#ffffff;letter-spacing:0.5px;">PimpMyPack</span>
</td></tr>
<!-- Body -->
<tr><td style="background-color:#ffffff;padding:32px;border-radius:0 0 12px 12px;box-shadow:0 1px 3px rgba(0,0,0,0.1);">
%s
</td></tr>
<!-- Footer -->
<tr><td style="padding:24px 32px;text-align:center;">
<p style="margin:0;font-size:13px;color:#78716c;line-height:1.5;">
&copy; PimpMyPack &mdash; Optimize your pack, enjoy the trail.
</p>
</td></tr>
</table>
</td></tr>
</table>
</body>
</html>`, safeTitle, safePreheader, bodyContent)
}

// BuildConfirmationEmailHTML returns the branded HTML body for a confirmation email.
func BuildConfirmationEmailHTML(username, confirmURL string) string {
	safeName := html.EscapeString(username)
	safeURL := html.EscapeString(confirmURL)
	body := fmt.Sprintf(`<h1 style="margin:0 0 16px;font-size:22px;color:#292524;">Welcome, %s!</h1>
<p style="margin:0 0 24px;font-size:16px;color:#292524;line-height:1.6;">
Thanks for signing up for PimpMyPack. Please confirm your email address to activate your account.
</p>
<table role="presentation" cellpadding="0" cellspacing="0" border="0" style="margin:0 auto 24px;">
<tr><td style="border-radius:8px;background-color:#16a34a;text-align:center;">
<a href="%s" target="_blank"
style="display:inline-block;padding:14px 32px;font-size:16px;
font-weight:600;color:#ffffff;text-decoration:none;
border-radius:8px;">Confirm my email</a>
</td></tr>
</table>
<p style="margin:0;font-size:13px;color:#78716c;line-height:1.5;word-break:break-all;">
If the button doesn't work, copy and paste this link into your browser:<br>
<a href="%s" style="color:#16a34a;">%s</a>
</p>`, safeName, safeURL, safeURL, safeURL)

	return BuildEmailHTML(
		"Confirm your email — PimpMyPack",
		"Please confirm your email address to activate your PimpMyPack account.",
		body,
	)
}

// sanitizeTextContent strips \r and \n to prevent content injection in plain-text emails.
func sanitizeTextContent(v string) string {
	v = strings.ReplaceAll(v, "\r", "")
	v = strings.ReplaceAll(v, "\n", "")
	return v
}

// BuildConfirmationEmailText returns the plain-text body for a confirmation email.
func BuildConfirmationEmailText(username, confirmURL string) string {
	return fmt.Sprintf(`Welcome, %s!

Thanks for signing up for PimpMyPack. Please confirm your email address to activate your account.

Confirm your email:
%s

--
PimpMyPack - Optimize your pack, enjoy the trail.
`, sanitizeTextContent(username), confirmURL)
}

// BuildPasswordResetEmailHTML returns the branded HTML body for a password reset email.
func BuildPasswordResetEmailHTML(newPassword string) string {
	safePassword := html.EscapeString(newPassword)
	body := fmt.Sprintf(`<h1 style="margin:0 0 16px;font-size:22px;color:#292524;">Password Reset</h1>
<p style="margin:0 0 16px;font-size:16px;color:#292524;line-height:1.6;">
Your password has been reset. Here is your new temporary password:
</p>
<table role="presentation" cellpadding="0" cellspacing="0" border="0" width="100%%" style="margin:0 0 24px;">
<tr><td style="background-color:#fafaf9;border:1px solid #e7e5e4;
border-radius:8px;padding:16px 24px;text-align:center;">
<code style="font-size:20px;font-family:'Courier New',Courier,monospace;color:#292524;letter-spacing:1px;">%s</code>
</td></tr>
</table>
<p style="margin:0 0 8px;font-size:16px;color:#292524;line-height:1.6;">
We recommend changing your password after logging in.
</p>
<p style="margin:0;font-size:13px;color:#78716c;line-height:1.5;">
If you did not request this password reset, please contact us immediately.
</p>`, safePassword)

	return BuildEmailHTML(
		"Password Reset — PimpMyPack",
		"Your PimpMyPack password has been reset.",
		body,
	)
}

// BuildPasswordResetEmailText returns the plain-text body for a password reset email.
func BuildPasswordResetEmailText(newPassword string) string {
	return fmt.Sprintf(`Password Reset

Your password has been reset. Here is your new temporary password:

    %s

We recommend changing your password after logging in.

If you did not request this password reset, please contact us immediately.

--
PimpMyPack - Optimize your pack, enjoy the trail.
`, newPassword)
}
