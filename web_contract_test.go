package main

import (
	"os"
	"strings"
	"testing"
)

func readFrontendFile(t *testing.T, path string) string {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(content)
}

func TestFrontendUsesReactComponentStructure(t *testing.T) {
	index := readFrontendFile(t, "frontend/templates/index.html")
	login := readFrontendFile(t, "frontend/templates/login.html")
	app := readFrontendFile(t, "frontend/static/js/app.jsx")

	for _, html := range []string{index, login} {
		for _, required := range []string{`id="root"`, "/vendor/react.production.min.js", "/vendor/react-dom.production.min.js"} {
			if !strings.Contains(html, required) {
				t.Fatalf("React page shell missing %s", required)
			}
		}
	}

	for _, required := range []string{"function AppShell(", "function ClipboardPanel(", "function RecentItems(", "function AccountMenu(", "function UserManagement("} {
		if !strings.Contains(app, required) {
			t.Fatalf("React component boundary missing: %s", required)
		}
	}
}

func TestFrontendHidesManualClipboardIDs(t *testing.T) {
	html := readFrontendFile(t, "frontend/templates/index.html")
	js := readFrontendFile(t, "frontend/static/js/app.jsx")
	translations := readFrontendFile(t, "frontend/static/js/i18n.js")

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
	auth := readFrontendFile(t, "frontend/static/js/auth.js")
	login := readFrontendFile(t, "frontend/templates/login.html")

	for _, forbidden := range []string{"redirect=", "window.location.hash", "encodeURIComponent", "getRedirectTarget", "URLSearchParams"} {
		if strings.Contains(auth+login, forbidden) {
			t.Fatalf("share-link redirect workflow still present: %s", forbidden)
		}
	}
}

func TestAuthCheckRejectsNonOKCurrentUserResponses(t *testing.T) {
	auth := readFrontendFile(t, "frontend/static/js/auth.js")

	checkStart := strings.Index(auth, "static async checkAuth()")
	checkEnd := strings.Index(auth, "static getAuthHeader()")
	if checkStart < 0 || checkEnd < 0 || checkEnd <= checkStart {
		t.Fatal("Auth.checkAuth function boundary missing")
	}
	checkAuth := auth[checkStart:checkEnd]
	if !strings.Contains(checkAuth, "if (!response.ok)") {
		t.Fatal("Auth.checkAuth must reject non-OK /api/auth/me responses before caching currentUser")
	}
}

func TestFrontendDecodesRFC5987DownloadFilenames(t *testing.T) {
	js := readFrontendFile(t, "frontend/static/js/app.jsx")
	for _, required := range []string{"getDownloadFilename(", "filename\\*", "decodeURIComponent"} {
		if !strings.Contains(js, required) {
			t.Fatalf("frontend filename parser does not decode RFC 5987 filename*: missing %s", required)
		}
	}
}

func TestRecentItemPrimaryActionsCopyTextAndDownloadFiles(t *testing.T) {
	js := readFrontendFile(t, "frontend/static/js/app.jsx")
	translations := readFrontendFile(t, "frontend/static/js/i18n.js")
	for _, required := range []string{"copyTextItem(", "copyTextItem(id)", "downloadFile(id)", "item-action-copy-text", "item-action-download-file"} {
		if !strings.Contains(js+translations, required) {
			t.Fatalf("recent item action contract missing: %s", required)
		}
	}

	if strings.Contains(js, `title="Load"`) {
		t.Fatal("recent item primary action still exposes a generic Load title")
	}
}

func TestReactFilePickerKeepsDragAndDrop(t *testing.T) {
	js := readFrontendFile(t, "frontend/static/js/app.jsx")

	for _, required := range []string{"onDragOver", "onDragLeave", "onDrop", "handleDroppedFile(", "setDragActive("} {
		if !strings.Contains(js, required) {
			t.Fatalf("React file picker lost drag-and-drop support: missing %s", required)
		}
	}
}

func TestFrontendProvidesPasswordAndUserManagement(t *testing.T) {
	app := readFrontendFile(t, "frontend/static/js/app.jsx")
	auth := readFrontendFile(t, "frontend/static/js/auth.js")
	translations := readFrontendFile(t, "frontend/static/js/i18n.js")

	for _, required := range []string{"changePassword(", "`/api/users/${userId}/password`", "newPassword", "function ChangePasswordModal("} {
		if !strings.Contains(app+auth, required) {
			t.Fatalf("change password workflow missing: %s", required)
		}
	}

	for _, required := range []string{"function UserManagement(", "loadUsers(", "createUser(", "updateUser(", "deleteUser(", "resetUserPassword(", "/api/users"} {
		if !strings.Contains(app, required) {
			t.Fatalf("user management workflow missing: %s", required)
		}
	}

	for _, required := range []string{
		"function ResetPasswordModal({ user, onClose, onSubmit, showMessage })",
		"showMessage(i18n.t('password-too-short'), 'error')",
		"showMessage(i18n.t('password-mismatch'), 'error')",
		"showMessage: showMessage",
		"includeStatus: false",
		"includeStatus &&",
		"{ username: form.username, password: form.password, email: form.email, role: form.role }",
	} {
		if !strings.Contains(app, required) {
			t.Fatalf("user management validation or API contract missing: %s", required)
		}
	}

	for _, required := range []string{"change-password", "user-management", "create-user", "reset-password"} {
		if !strings.Contains(translations, required) {
			t.Fatalf("user management translation missing: %s", required)
		}
	}
}

func TestGoRouterServesReactStaticAssets(t *testing.T) {
	main := readFrontendFile(t, "backend/cmd/web-clipboard/main.go")

	for _, required := range []string{`mime.AddExtensionType(".jsx", "application/javascript")`, `router.Static("/static", "./frontend/static")`, `router.StaticFile("/favicon.ico"`} {
		if !strings.Contains(main, required) {
			t.Fatalf("Go router static asset route missing: %s", required)
		}
	}

	for _, forbidden := range []string{`router.GET("/app.js"`, `router.GET("/auth.js"`, `router.GET("/i18n.js"`} {
		if strings.Contains(main, forbidden) {
			t.Fatalf("legacy fixed script route still present: %s", forbidden)
		}
	}
}

func TestPasswordRouteIsAuthenticatedButNotAdminOnly(t *testing.T) {
	main := readFrontendFile(t, "backend/cmd/web-clipboard/main.go")

	if !strings.Contains(main, `api.PUT("/users/:id/password", handler.ChangeUserPassword)`) {
		t.Fatal("password change route must be available to authenticated users")
	}
	if strings.Contains(main, `users.PUT("/:id/password", handler.ChangeUserPassword)`) {
		t.Fatal("password change route is still nested under admin-only users group")
	}
}
