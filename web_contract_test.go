package main

import (
	"os"
	"strings"
	"testing"
)

func TestFrontendHidesManualClipboardIDs(t *testing.T) {
	index, err := os.ReadFile("web/templates/index.html")
	if err != nil {
		t.Fatal(err)
	}
	app, err := os.ReadFile("web/static/js/app.js")
	if err != nil {
		t.Fatal(err)
	}
	i18n, err := os.ReadFile("web/static/js/i18n.js")
	if err != nil {
		t.Fatal(err)
	}

	html := string(index)
	js := string(app)
	translations := string(i18n)

	for _, forbidden := range []string{`id="textId"`, `id="fileId"`} {
		if strings.Contains(html, forbidden) {
			t.Fatalf("manual clipboard ID input still present: %s", forbidden)
		}
	}

	for _, forbidden := range []string{"copyId(", "Copy ID", "ID:"} {
		if strings.Contains(js, forbidden) {
			t.Fatalf("bare clipboard ID UI still present in app.js: %s", forbidden)
		}
	}

	for _, forbidden := range []string{"text-id-placeholder", "file-id-placeholder", "please-enter-text-id", "please-enter-file-id", "id-copied"} {
		if strings.Contains(translations, forbidden) {
			t.Fatalf("bare clipboard ID translation still present: %s", forbidden)
		}
	}

	for _, forbidden := range []string{"copyLink(", "createShareUrl(", "loadSharedLinkFromHash(", "text-link-copied", "file-link-copied", "copy-link"} {
		if strings.Contains(js+translations, forbidden) {
			t.Fatalf("sharing workflow marker still present: %s", forbidden)
		}
	}
}

func TestLoginDoesNotRestoreShareLinks(t *testing.T) {
	auth, err := os.ReadFile("web/static/js/auth.js")
	if err != nil {
		t.Fatal(err)
	}
	login, err := os.ReadFile("web/templates/login.html")
	if err != nil {
		t.Fatal(err)
	}

	for _, forbidden := range []string{"redirect=", "window.location.hash", "encodeURIComponent", "getRedirectTarget", "URLSearchParams"} {
		if strings.Contains(string(auth)+string(login), forbidden) {
			t.Fatalf("share-link redirect workflow still present: %s", forbidden)
		}
	}
}

func TestFrontendDecodesRFC5987DownloadFilenames(t *testing.T) {
	app, err := os.ReadFile("web/static/js/app.js")
	if err != nil {
		t.Fatal(err)
	}

	js := string(app)
	for _, required := range []string{"getDownloadFilename(", "filename\\*", "decodeURIComponent"} {
		if !strings.Contains(js, required) {
			t.Fatalf("frontend filename parser does not decode RFC 5987 filename*: missing %s", required)
		}
	}
}

func TestRecentItemPrimaryActionsCopyTextAndDownloadFiles(t *testing.T) {
	app, err := os.ReadFile("web/static/js/app.js")
	if err != nil {
		t.Fatal(err)
	}
	i18n, err := os.ReadFile("web/static/js/i18n.js")
	if err != nil {
		t.Fatal(err)
	}

	js := string(app)
	translations := string(i18n)
	for _, required := range []string{"copyTextItem(", "copyTextItem(id)", "downloadFile(id)", "item-action-copy-text", "item-action-download-file"} {
		if !strings.Contains(js+translations, required) {
			t.Fatalf("recent item action contract missing: %s", required)
		}
	}

	if strings.Contains(js, `title="Load"`) {
		t.Fatal("recent item primary action still exposes a generic Load title")
	}
}
