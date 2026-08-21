package utils

import (
	"strings"
	"testing"
)

func TestEmailDomain(t *testing.T) {
	if got := emailDomain("info@stenakladno.cz"); got != "stenakladno.cz" {
		t.Errorf("emailDomain() = %q, want %q", got, "stenakladno.cz")
	}
	if got := emailDomain("apikey"); got != "localhost" {
		t.Errorf("emailDomain() = %q, want %q", got, "localhost")
	}
}

func TestNewMessageID(t *testing.T) {
	a := newMessageID("stenakladno.cz")
	b := newMessageID("stenakladno.cz")

	if !strings.HasPrefix(a, "<") || !strings.HasSuffix(a, "@stenakladno.cz>") {
		t.Errorf("newMessageID() = %q, want format <hex@domain>", a)
	}
	if a == b {
		t.Errorf("newMessageID() returned the same value twice: %q", a)
	}
}

func TestWrapHTMLDocument(t *testing.T) {
	fragment := "<p>hello</p>"
	wrapped := wrapHTMLDocument(fragment)
	if !strings.Contains(wrapped, "<html") || !strings.Contains(wrapped, "<body>"+fragment+"</body>") {
		t.Errorf("wrapHTMLDocument() = %q, want a full html document wrapping the fragment", wrapped)
	}

	full := "<html><head></head><body><p>hello</p></body></html>"
	if got := wrapHTMLDocument(full); got != full {
		t.Errorf("wrapHTMLDocument() should be a no-op when <html> is already present, got %q", got)
	}
}

func TestPlainTextFromHTML(t *testing.T) {
	// mirrors modules/security/security-service.go's login code email body
	loginBody := "<p><strong>Lezecká stěna Kladno – přihlášení / registrace</strong></p>" +
		"<p style='letter-spacing: 0.75px;'>Váš jednorázový přihlašovací kód je: <a target='_blank' href='https://example.com/sign-me-in?c=abc' style='color: rgb(219 39 119);' ><span style='font-size:20px;letter-spacing: 2px;'>12345</span></a></p>" +
		"<p style='letter-spacing: 0.75px;'>Kliknutím na kód je možné se rovnou přihlásit.</p>" +
		"<p style='font-size:13px;color: #f40d0d;letter-spacing: 0.5px;'>Tento kód je platný pouze 10 minut.</p>" +
		"<p style='font-size:13px;color: #4d4d4d;letter-spacing: 0.5px;'>Pokud jste o tento kód nepožádali, ignorujte tento email.</p>"

	plain := plainTextFromHTML(loginBody)
	if strings.Contains(plain, "<") || strings.Contains(plain, ">") {
		t.Errorf("plainTextFromHTML() left HTML tags behind: %q", plain)
	}
	if !strings.Contains(plain, "Lezecká stěna Kladno") || !strings.Contains(plain, "12345") {
		t.Errorf("plainTextFromHTML() lost expected content: %q", plain)
	}

	// mirrors modules/courses/courses-service.go's course application email,
	// which embeds a QR code image via a cid: reference
	courseBody := "<div style=\"width: 100%; max-width: 600px;\">\n" +
		"<p style=\"font-size: 20px;\">Dobrý den,</p>\n\n" +
		"<p>Pro okamžitou platbu kurzu můžete použít QR kód:</p>\n\n" +
		"<img src=\"cid:1qr.png\" style=\"margin-bottom: 20px;\"/>\n\n" +
		"<p>Variabilní symbol: 1</p>\n\n" +
		"</div>\n"

	plain = plainTextFromHTML(courseBody)
	if strings.Contains(plain, "<") || strings.Contains(plain, ">") {
		t.Errorf("plainTextFromHTML() left HTML tags behind: %q", plain)
	}
	if strings.Contains(plain, "cid:") {
		t.Errorf("plainTextFromHTML() leaked a cid: reference into plaintext: %q", plain)
	}
	if !strings.Contains(plain, "Dobrý den") || !strings.Contains(plain, "Variabilní symbol: 1") {
		t.Errorf("plainTextFromHTML() lost expected content: %q", plain)
	}
}
