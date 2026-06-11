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
	index := readFrontendFile(t, "frontend/index.html")
	login := readFrontendFile(t, "frontend/login.html")
	settings := readFrontendFile(t, "frontend/settings.html")
	app := readFrontendFile(t, "frontend/src/App.jsx")
	shared := readFrontendFile(t, "frontend/src/shared.jsx")
	account := readFrontendFile(t, "frontend/src/account.jsx")
	users := readFrontendFile(t, "frontend/src/users.jsx")
	settingsJS := readFrontendFile(t, "frontend/src/settings.jsx")

	for _, html := range []string{index, login, settings} {
		for _, required := range []string{`id="root"`, `type="module"`} {
			if !strings.Contains(html, required) {
				t.Fatalf("React page shell missing %s", required)
			}
		}
		for _, forbidden := range []string{"cdn.tailwindcss.com", "/static/vendor/", "react.production.min.js", "react-dom.production.min.js"} {
			if strings.Contains(html, forbidden) {
				t.Fatalf("React page shell still directly loads external/vendor package: %s", forbidden)
			}
		}
	}

	for _, required := range []string{"function AppShell(", "function ClipboardPanel(", "function RecentItems("} {
		if !strings.Contains(app, required) {
			t.Fatalf("main page component boundary missing: %s", required)
		}
	}
	for _, required := range []string{"function IconLabel(", "function StatusMessage(", "function Modal(", "function ModalActions(", "function TextField(", "function PasswordField("} {
		if !strings.Contains(shared, required) {
			t.Fatalf("shared component boundary missing: %s", required)
		}
	}
	for _, required := range []string{"function AccountMenu(", "function ChangePasswordModal("} {
		if !strings.Contains(account, required) {
			t.Fatalf("account component boundary missing: %s", required)
		}
	}
	for _, required := range []string{"function UserManagement(", "function UserFormModal(", "function ResetPasswordModal("} {
		if !strings.Contains(users, required) {
			t.Fatalf("user management component boundary missing: %s", required)
		}
	}
	for _, required := range []string{"function SettingsShell(", "function AccountSettings(", "function AdminSettings("} {
		if !strings.Contains(settingsJS, required) {
			t.Fatalf("settings page component boundary missing: %s", required)
		}
	}
}

func TestFrontendHidesManualClipboardIDs(t *testing.T) {
	html := readFrontendFile(t, "frontend/index.html")
	js := readFrontendFile(t, "frontend/src/App.jsx")
	translations := readFrontendFile(t, "frontend/src/i18n.js")

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
	auth := readFrontendFile(t, "frontend/src/auth.js")
	login := readFrontendFile(t, "frontend/login.html")
	loginApp := readFrontendFile(t, "frontend/src/LoginApp.jsx")

	for _, forbidden := range []string{"redirect=", "window.location.hash", "encodeURIComponent", "getRedirectTarget", "URLSearchParams"} {
		if strings.Contains(auth+login, forbidden) {
			t.Fatalf("share-link redirect workflow still present: %s", forbidden)
		}
	}
	for _, forbidden := range []string{"redirect=", "window.location.hash", "getRedirectTarget"} {
		if strings.Contains(loginApp, forbidden) {
			t.Fatalf("share-link redirect workflow still present in LoginApp: %s", forbidden)
		}
	}
}

func TestAuthCheckRejectsNonOKCurrentUserResponses(t *testing.T) {
	auth := readFrontendFile(t, "frontend/src/auth.js")

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
	js := readFrontendFile(t, "frontend/src/App.jsx")
	for _, required := range []string{"getDownloadFilename(", "filename\\*", "decodeURIComponent"} {
		if !strings.Contains(js, required) {
			t.Fatalf("frontend filename parser does not decode RFC 5987 filename*: missing %s", required)
		}
	}
}

func TestRecentItemPrimaryActionsCopyTextAndDownloadFiles(t *testing.T) {
	js := readFrontendFile(t, "frontend/src/App.jsx")
	translations := readFrontendFile(t, "frontend/src/i18n.js")
	for _, required := range []string{"copyTextItem(", "copyTextItem(id)", "downloadFile(id)", "item-action-copy-text", "item-action-download-file"} {
		if !strings.Contains(js+translations, required) {
			t.Fatalf("recent item action contract missing: %s", required)
		}
	}

	if strings.Contains(js, `title="Load"`) {
		t.Fatal("recent item primary action still exposes a generic Load title")
	}
}

func TestFrontendLoadsRecentItemsFromAuthenticatedAPI(t *testing.T) {
	app := readFrontendFile(t, "frontend/src/App.jsx")
	main := readFrontendFile(t, "backend/cmd/web-clipboard/main.go")
	translations := readFrontendFile(t, "frontend/src/i18n.js")

	for _, required := range []string{"loadRecentItems(", "loadRecentItems(false)", "Auth.json('/api/items')", "setRecentItems(data.items || [])", "load-recent-failed"} {
		if !strings.Contains(app+translations, required) {
			t.Fatalf("server-backed recent items workflow missing: %s", required)
		}
	}
	if !strings.Contains(main, `api.GET("/items", handler.ListRecentItems)`) {
		t.Fatal("recent items route must be available to authenticated users")
	}
	for _, forbidden := range []string{"localStorage.getItem('recentItems')", "localStorage.setItem('recentItems'"} {
		if strings.Contains(app, forbidden) {
			t.Fatalf("recent items must not be isolated in browser localStorage: %s", forbidden)
		}
	}
}

func TestFrontendUsesIconsAndToastFeedback(t *testing.T) {
	app := readFrontendFile(t, "frontend/src/App.jsx")
	shared := readFrontendFile(t, "frontend/src/shared.jsx")
	account := readFrontendFile(t, "frontend/src/account.jsx")
	loginApp := readFrontendFile(t, "frontend/src/LoginApp.jsx")
	packageJSON := readFrontendFile(t, "frontend/package.json")
	source := app + shared + account

	if !strings.Contains(packageJSON, `"lucide-react"`) {
		t.Fatal("frontend package.json must include a pnpm-managed icon library")
	}
	for _, required := range []string{
		"import {",
		"Copy,",
		"Download,",
		"FileIcon,",
		"FileText,",
		"FolderOpen,",
		"LogOut,",
		"Save,",
		"Upload,",
		"} from 'lucide-react';",
		"function IconLabel(",
		"function RecentTypeIcon(",
		"aria-label': type === 'text' ? 'Text item' : 'File item'",
		"type === 'text' ? FileText : FileIcon",
		"e(RecentTypeIcon, { type: item.type })",
		"sr-only",
	} {
		if !strings.Contains(source, required) {
			t.Fatalf("icon UI contract missing: %s", required)
		}
	}
	if strings.Contains(app, "item.type === 'text' ? 'Text' : 'File'") {
		t.Fatal("recent item type is still rendered as visible Text/File copy instead of an icon")
	}
	for _, required := range []string{
		"function StatusMessage({ message })",
		"fixed top-4 right-4",
		"role: 'status'",
		"'aria-live': 'polite'",
		"showMessage(i18n.t('text-copied'))",
	} {
		if !strings.Contains(source, required) {
			t.Fatalf("copy success toast contract missing: %s", required)
		}
	}
	for _, required := range []string{
		"from 'lucide-react';",
		"LogIn",
		"function IconLabel(",
		"icon: LogIn",
		"inline-flex items-center justify-center gap-2",
	} {
		if !strings.Contains(loginApp, required) {
			t.Fatalf("login icon UI contract missing: %s", required)
		}
	}
}

func TestReactFilePickerKeepsDragAndDrop(t *testing.T) {
	js := readFrontendFile(t, "frontend/src/App.jsx")

	for _, required := range []string{"onDragOver", "onDragLeave", "onDrop", "handleDroppedFile(", "setDragActive("} {
		if !strings.Contains(js, required) {
			t.Fatalf("React file picker lost drag-and-drop support: missing %s", required)
		}
	}
}

func TestFrontendProvidesPasswordAndUserManagement(t *testing.T) {
	app := readFrontendFile(t, "frontend/src/App.jsx")
	account := readFrontendFile(t, "frontend/src/account.jsx")
	users := readFrontendFile(t, "frontend/src/users.jsx")
	settings := readFrontendFile(t, "frontend/src/settings.jsx")
	auth := readFrontendFile(t, "frontend/src/auth.js")
	translations := readFrontendFile(t, "frontend/src/i18n.js")

	for _, required := range []string{"changePassword(", "`/api/users/${userId}/password`", "newPassword", "function ChangePasswordModal("} {
		if !strings.Contains(account+auth, required) {
			t.Fatalf("change password workflow missing: %s", required)
		}
	}

	for _, required := range []string{"function UserManagement(", "loadUsers(", "createUser(", "updateUser(", "deleteUser(", "resetUserPassword(", "/api/users"} {
		if !strings.Contains(users, required) {
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
		if !strings.Contains(users, required) {
			t.Fatalf("user management validation or API contract missing: %s", required)
		}
	}

	for _, required := range []string{"user?.role === 'admin'", "e(AdminSettings", "e(UserManagement"} {
		if !strings.Contains(settings, required) {
			t.Fatalf("settings page must gate admin-only user management: %s", required)
		}
	}
	if strings.Contains(app, "e(UserManagement") {
		t.Fatal("main page must not render administrator user management")
	}

	for _, required := range []string{"change-password", "user-management", "create-user", "reset-password"} {
		if !strings.Contains(translations, required) {
			t.Fatalf("user management translation missing: %s", required)
		}
	}
}

func TestGoRouterServesReactStaticAssets(t *testing.T) {
	main := readFrontendFile(t, "backend/cmd/web-clipboard/main.go")
	middleware := readFrontendFile(t, "backend/internal/middleware/middleware.go")

	for _, required := range []string{`router.Static("/assets", "./frontend/dist/assets")`, `router.StaticFile("/favicon.ico", "./frontend/dist/favicon.ico")`, `c.File("./frontend/dist/login.html")`, `c.File("./frontend/dist/settings.html")`, `c.File("./frontend/dist/index.html")`} {
		if !strings.Contains(main, required) {
			t.Fatalf("Go router static asset route missing: %s", required)
		}
	}

	for _, forbidden := range []string{`router.GET("/app.js"`, `router.GET("/auth.js"`, `router.GET("/i18n.js"`} {
		if strings.Contains(main, forbidden) {
			t.Fatalf("legacy fixed script route still present: %s", forbidden)
		}
	}
	if strings.Contains(middleware, "cdn.tailwindcss.com") {
		t.Fatal("CSP still allows the old Tailwind CDN after moving styles into pnpm build")
	}
}

func TestFrontendUsesPnpmManagedBuild(t *testing.T) {
	packageJSON := readFrontendFile(t, "frontend/package.json")
	viteConfig := readFrontendFile(t, "frontend/vite.config.js")
	pnpmLock := readFrontendFile(t, "frontend/pnpm-lock.yaml")
	pnpmWorkspace := readFrontendFile(t, "frontend/pnpm-workspace.yaml")
	makefile := readFrontendFile(t, "Makefile")

	for _, required := range []string{`"react"`, `"react-dom"`, `"vite"`, `"tailwindcss"`, `"build"`} {
		if !strings.Contains(packageJSON, required) {
			t.Fatalf("frontend package.json missing pnpm-managed dependency or script: %s", required)
		}
	}
	for _, required := range []string{"index.html", "login.html", "settings.html", "manifest: true"} {
		if !strings.Contains(viteConfig, required) {
			t.Fatalf("Vite multi-page build config missing: %s", required)
		}
	}
	if !strings.Contains(pnpmLock, "react") || !strings.Contains(pnpmLock, "vite") {
		t.Fatal("pnpm-lock.yaml does not include frontend runtime/build dependencies")
	}
	if !strings.Contains(pnpmWorkspace, "esbuild: true") {
		t.Fatal("pnpm-workspace.yaml must allow esbuild build scripts for pnpm 11")
	}
	if !strings.Contains(makefile, "FRONTEND_DIR := frontend") ||
		!strings.Contains(makefile, "pnpm --dir $(FRONTEND_DIR) install") ||
		!strings.Contains(makefile, "pnpm --dir $(FRONTEND_DIR) build") {
		t.Fatal("Makefile must install and build the pnpm frontend before Go build")
	}
}

func TestSettingsPageEntryAndMainNavigation(t *testing.T) {
	settingsHTML := readFrontendFile(t, "frontend/settings.html")
	app := readFrontendFile(t, "frontend/src/App.jsx")
	account := readFrontendFile(t, "frontend/src/account.jsx")
	settings := readFrontendFile(t, "frontend/src/settings.jsx")
	translations := readFrontendFile(t, "frontend/src/i18n.js")

	for _, required := range []string{`src="/src/settings.jsx"`, `id="root"`, `type="module"`} {
		if !strings.Contains(settingsHTML, required) {
			t.Fatalf("settings page shell missing: %s", required)
		}
	}
	for _, required := range []string{`href: '/settings.html'`, "settings"} {
		if !strings.Contains(app+account+translations, required) {
			t.Fatalf("main page settings entry missing: %s", required)
		}
	}
	for _, required := range []string{"Auth.requireAuth()", "window.location.href = '/login.html'", "href: '/'", "account-settings", "system-settings"} {
		if !strings.Contains(settings+translations, required) {
			t.Fatalf("settings page authenticated shell missing: %s", required)
		}
	}
}

func TestPersistentDataDirectoryDefaultsToDataVolume(t *testing.T) {
	main := readFrontendFile(t, "backend/cmd/web-clipboard/main.go")
	dockerfile := readFrontendFile(t, "Dockerfile")
	compose := readFrontendFile(t, "docker-compose.yml")
	readme := readFrontendFile(t, "README.md")

	for _, required := range []string{"func getDataDir() string", `os.Getenv("WEB_CLIPBOARD_DATA_DIR")`, `return "/data"`, "services.NewUserManager(getDataDir())"} {
		if !strings.Contains(main, required) {
			t.Fatalf("backend data directory contract missing: %s", required)
		}
	}
	for _, required := range []string{`mkdir -p /data`, `chown -R appuser:appuser /app /data`, `VOLUME ["/data"]`} {
		if !strings.Contains(dockerfile, required) {
			t.Fatalf("Dockerfile persistent data contract missing: %s", required)
		}
	}
	if !strings.Contains(compose, "./data:/data") {
		t.Fatal("docker-compose.yml must mount ./data to /data")
	}
	for _, required := range []string{"-v ./data:/data", "WEB_CLIPBOARD_DATA_DIR", "/settings.html", "/data/users.json"} {
		if !strings.Contains(readme, required) {
			t.Fatalf("README persistent data or settings route documentation missing: %s", required)
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

func TestOAuthRoutesAndLoginEntryPointsExist(t *testing.T) {
	main := readFrontendFile(t, "backend/cmd/web-clipboard/main.go")
	auth := readFrontendFile(t, "frontend/src/auth.js")
	loginApp := readFrontendFile(t, "frontend/src/LoginApp.jsx")

	for _, required := range []string{
		`auth.GET("/providers", handler.ListAuthProviders)`,
		`auth.GET("/oauth/:provider/start", handler.StartOAuthLogin)`,
		`auth.GET("/oauth/:provider/callback", handler.HandleOAuthCallback)`,
		`auth.POST("/oauth/complete", handler.CompleteOAuthLogin)`,
	} {
		if !strings.Contains(main, required) {
			t.Fatalf("OAuth route missing: %s", required)
		}
	}

	for _, required := range []string{
		"getAuthProviders(",
		"completeOAuthLogin(",
		"'/api/auth/providers'",
		"'/api/auth/oauth/complete'",
		"`/api/auth/oauth/${provider}/start`",
	} {
		if !strings.Contains(auth, required) {
			t.Fatalf("frontend OAuth auth helper missing: %s", required)
		}
	}

	for _, required := range []string{
		"loadAuthProviders(",
		"params.get('oauth') === 'complete'",
		"Auth.completeOAuthLogin()",
		"Auth.startOAuthLogin(provider.name)",
	} {
		if !strings.Contains(loginApp, required) {
			t.Fatalf("login OAuth entry point missing: %s", required)
		}
	}
}
