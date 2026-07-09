package app

import (
	"context"
	"fmt"
	"html"
	"log"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"telegram-business-bot/internal/domain"
	"telegram-business-bot/internal/service"
)

type Monitor struct {
	service    service.MessageService
	adminToken string
	server     *http.Server
	startTime  time.Time
}

func NewMonitor(service service.MessageService, adminToken string) *Monitor {
	return &Monitor{service: service, adminToken: adminToken, startTime: time.Now()}
}

func (m *Monitor) Run(addr string) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/favicon.ico", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("/", m.auth(m.handleIndex))
	mux.HandleFunc("/subscribers", m.auth(m.handleSubscribers))
	mux.HandleFunc("/users", m.auth(m.handleUsers))
	mux.HandleFunc("/chats", m.auth(m.handleChats))
	mux.HandleFunc("/user", m.auth(m.handleUser))
	mux.HandleFunc("/chat", m.auth(m.handleChat))
	mux.HandleFunc("/health", m.handleHealth)

	m.server = &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	log.Printf("Starting admin dashboard on %s", addr)
	return m.server.ListenAndServe()
}

func (m *Monitor) Shutdown(ctx context.Context) error {
	if m.server == nil {
		return nil
	}
	log.Println("Shutting down admin dashboard")
	return m.server.Shutdown(ctx)
}

func (m *Monitor) auth(next http.HandlerFunc) http.HandlerFunc {
	if m.adminToken == "" {
		return next
	}

	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("token") == m.adminToken || r.Header.Get("X-Admin-Token") == m.adminToken {
			next(w, r)
			return
		}
		w.Header().Set("WWW-Authenticate", "Bearer")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	}
}

func (m *Monitor) handleHealth(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("OK"))
}

func (m *Monitor) handleIndex(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	subs, err := m.service.GetAllBusinessSubscribers(ctx)
	if err != nil {
		log.Printf("Monitor: failed to load subscribers: %v", err)
		http.Error(w, "Failed to load dashboard data", http.StatusInternalServerError)
		return
	}

	chatGroups := groupSubscribersByChat(subs)
	userGroups := groupSubscribersByConnectionID(subs)
	chatIDs := make([]int64, 0, len(chatGroups))
	for chatID := range chatGroups {
		chatIDs = append(chatIDs, chatID)
	}
	sort.Slice(chatIDs, func(i, j int) bool { return chatIDs[i] < chatIDs[j] })

	userIDs := make([]string, 0, len(userGroups))
	for connectionID := range userGroups {
		userIDs = append(userIDs, connectionID)
	}
	sort.Strings(userIDs)

	recentCount24h := m.countRecentSubscriptions(subs, 24*time.Hour)
	recentCount7d := m.countRecentSubscriptions(subs, 7*24*time.Hour)
	activityBars := m.buildActivityBars(subs)
	recentActivities := m.buildRecentActivities(subs)
	uptime := formatDuration(time.Since(m.startTime))

	userCount := len(userIDs)
	chatCount := len(chatIDs)
	totalSubs := len(subs)

	var htmlPage strings.Builder
	htmlPage.WriteString(`<!doctype html>
<html lang="ru">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Admin Dashboard</title>
<style>
:root{--bg:#0B0F19;--surface:rgba(18,24,38,0.92);--surface-soft:rgba(18,24,38,0.78);--surface-strong:#121826;--border:rgba(255,255,255,0.08);--text:#F8FAFC;--muted:#94A3B8;--accent:#5B8CFF;--success:#22C55E;--warning:#F59E0B;--danger:#EF4444;--shadow:0 25px 90px rgba(0,0,0,0.30);--radius:18px;}
body{margin:0;min-height:100vh;font-family:Inter,system-ui,-apple-system,'Segoe UI',Roboto,Arial,sans-serif;background:radial-gradient(circle at top left,rgba(91,140,255,0.14),transparent 28%%),radial-gradient(circle at bottom right,rgba(91,140,255,0.08),transparent 22%%),#0B0F19;color:var(--text);}
body.light{--bg:#F8FAFC;--surface:rgba(255,255,255,0.92);--surface-soft:rgba(248,250,252,0.85);--surface-strong:#ffffff;--border:rgba(15,23,42,0.08);--text:#0F172A;--muted:#64748B;--accent:#5B8CFF;--success:#16A34A;--warning:#D97706;--danger:#DC2626;--shadow:0 20px 45px rgba(15,23,42,0.12);}
body{background-color:var(--bg);}
.page-shell{display:grid;grid-template-columns:280px minmax(0,1fr);gap:24px;padding:24px;}
.sidebar{position:sticky;top:24px;align-self:flex-start;padding:28px 20px 24px 24px;background:var(--surface);border:1px solid var(--border);border-radius:22px;backdrop-filter:blur(28px);box-shadow:var(--shadow);}
.brand{display:flex;align-items:center;gap:14px;margin-bottom:28px;}
.brand-dot{width:12px;height:12px;border-radius:50%%;background:linear-gradient(135deg,var(--accent),#7C3AED);box-shadow:0 0 18px rgba(91,140,255,0.35);}
.brand-title{font-size:0.95rem;font-weight:700;letter-spacing:0.13em;text-transform:uppercase;color:var(--accent);}
.brand-sub{font-size:0.82rem;color:var(--muted);line-height:1.5;}
.nav-list{display:grid;gap:10px;}
.nav-link{display:flex;align-items:center;gap:12px;padding:14px 16px;border-radius:16px;color:var(--text);text-decoration:none;font-size:0.95rem;transition:all 180ms ease;}
.nav-link:hover,.nav-link.active{background:rgba(91,140,255,0.14);border:1px solid rgba(91,140,255,0.18);}
.main-content{display:flex;flex-direction:column;gap:22px;}
.topbar{display:flex;justify-content:space-between;align-items:flex-start;gap:18px;padding-bottom:8px;border-bottom:1px solid rgba(255,255,255,0.08);}
.page-title{display:flex;flex-direction:column;gap:10px;}
.page-title h1{margin:0;font-size:2rem;font-weight:800;letter-spacing:-0.03em;}
.page-title p{margin:0;color:var(--muted);max-width:680px;line-height:1.75;font-size:0.96rem;}
.topbar-actions{display:flex;align-items:center;gap:12px;flex-wrap:wrap;}
.search-field{position:relative;min-width:260px;flex:1;max-width:420px;}
.search-field input{width:100%%;padding:14px 14px 14px 44px;border:1px solid rgba(255,255,255,0.10);border-radius:16px;background:rgba(255,255,255,0.06);color:var(--text);outline:none;transition:background 180ms ease,border-color 180ms ease,transform 180ms ease;}
.search-field input:focus{background:rgba(255,255,255,0.12);border-color:rgba(91,140,255,0.28);transform:translateY(-1px);}
.search-field svg{position:absolute;top:50%%;left:14px;transform:translateY(-50%%);stroke:var(--muted);}
.button-pill{display:inline-flex;align-items:center;gap:10px;padding:13px 18px;border-radius:16px;border:1px solid rgba(255,255,255,0.12);background:rgba(255,255,255,0.08);color:var(--text);cursor:pointer;transition:all 180ms ease;}
.button-pill:hover{transform:translateY(-1px);box-shadow:0 18px 40px rgba(91,140,255,0.18);}
.profile-badge{display:inline-flex;align-items:center;gap:12px;padding:12px 16px;border-radius:18px;background:rgba(255,255,255,0.05);border:1px solid rgba(255,255,255,0.10);}
.profile-avatar{width:38px;height:38px;border-radius:16px;background:linear-gradient(135deg,var(--accent),#7C3AED);display:grid;place-items:center;color:#ffffff;font-weight:700;font-size:0.95rem;}
.profile-badge div{display:flex;flex-direction:column;gap:2px;}
.profile-badge strong{font-size:0.95rem;}
.profile-badge span{font-size:0.82rem;color:var(--muted);}
.card-grid{display:grid;grid-template-columns:repeat(3,minmax(0,1fr));gap:18px;}
.card{background:var(--surface);border:1px solid var(--border);border-radius:22px;box-shadow:var(--shadow);backdrop-filter:blur(24px);padding:24px;transition:transform 200ms ease, border-color 200ms ease;}
.card:hover{transform:translateY(-2px);}
.card-compact{padding:20px;}
.card h2{margin:0 0 16px;font-size:1rem;font-weight:700;color:var(--text);}
.stat-card{display:grid;gap:8px;}
.stat-card .label{color:var(--muted);font-size:0.9rem;}
.stat-card .value{font-size:2.4rem;font-weight:800;line-height:1;}
.stat-pill{display:inline-flex;align-items:center;gap:8px;padding:10px 14px;border-radius:999px;background:rgba(91,140,255,0.13);color:var(--accent);font-size:0.85rem;font-weight:600;}
.chart-grid{display:grid;grid-template-columns:1.1fr 0.9fr;gap:18px;}
.chart-panel{display:flex;flex-direction:column;gap:18px;}
.chart-bars{display:flex;align-items:flex-end;gap:12px;height:190px;padding:12px;}
.chart-bar{position:relative;flex:1;display:flex;align-items:flex-end;justify-content:center;}
.chart-bar strong{display:block;margin-top:12px;font-size:0.82rem;color:var(--muted);}
.chart-bar span{position:absolute;top:0;left:50%%;transform:translate(-50%%,-50%%);font-size:0.78rem;color:var(--text);background:rgba(11,15,25,0.8);padding:4px 8px;border-radius:999px;}
.chart-bar::before{content:'';width:100%%;border-radius:16px;background:linear-gradient(180deg,rgba(91,140,255,0.9),rgba(91,140,255,0.22));box-shadow:0 18px 40px rgba(91,140,255,0.12);position:absolute;bottom:0;left:0;right:0;}
.chart-bar .bar-fill{width:100%%;border-radius:16px;background:linear-gradient(180deg,rgba(91,140,255,0.95),rgba(91,140,255,0.35));position:relative;z-index:1;}
.activity-list{display:grid;gap:14px;}
.activity-item{display:flex;justify-content:space-between;gap:18px;padding:18px;border-radius:20px;background:rgba(255,255,255,0.05);border:1px solid rgba(255,255,255,0.08);}
.activity-item strong{font-size:0.95rem;}
.activity-item span{color:var(--muted);font-size:0.85rem;}
.table-card{overflow:hidden;}
.table-card table{width:100%%;border-collapse:collapse;}
.table-card thead tr{background:rgba(255,255,255,0.04);}
.table-card th,.table-card td{padding:14px 16px;text-align:left;border-bottom:1px solid rgba(255,255,255,0.08);color:var(--text);}
.table-card th{font-size:0.82rem;color:var(--muted);letter-spacing:0.02em;text-transform:uppercase;}
.table-card tr:hover{background:rgba(255,255,255,0.04);}
.table-card a{color:var(--accent);text-decoration:none;}
.table-card a:hover{text-decoration:underline;}
.footer-note{text-align:right;color:var(--muted);font-size:0.85rem;margin-top:14px;}
@media(max-width:1200px){.card-grid{grid-template-columns:repeat(2,minmax(0,1fr));}.chart-grid{grid-template-columns:1fr;}}
@media(max-width:900px){.page-shell{grid-template-columns:1fr;}.sidebar{position:relative;top:auto;width:auto;}.topbar{flex-direction:column;align-items:stretch;}.card-grid{grid-template-columns:1fr;}.search-field{min-width:100%%;}.profile-badge{width:100%%;justify-content:space-between;}.chart-bars{height:160px;}}
@media(max-width:640px){.page-shell{padding:16px;}.sidebar{padding:20px;}.topbar h1{font-size:1.6rem;}.nav-link{padding:12px;}.card{padding:20px;}.table-card th,.table-card td{padding:12px;}}
</style>
</head>
<body>
<div class="page-shell">
  <aside class="sidebar">
    <div class="brand">
      <span class="brand-dot"></span>
      <div>
        <div class="brand-title">Admin Studio</div>
        <div class="brand-sub">Telegram Bot @ssancodels_bot</div>
      </div>
    </div>
    <nav class="nav-list">
      <a class="nav-link active" href="` + html.EscapeString(m.authLink("/")) + `">Dashboard</a>
      <a class="nav-link" href="` + html.EscapeString(m.authLink("/users")) + `">Пользователи</a>
      <a class="nav-link" href="` + html.EscapeString(m.authLink("/subscribers")) + `">Подписчики</a>
      <a class="nav-link" href="` + html.EscapeString(m.authLink("/chats")) + `">Чаты</a>
    </nav>
  </aside>
  <main class="main-content">
    <header class="topbar">
      <div class="page-title">
        <h1>Admin Dashboard</h1>
        <p>Премиум-панель для контроля пользователей, чатов и активности Telegram-бота.</p>
      </div>
      <div class="topbar-actions">
        <label class="search-field">
          <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg"><circle cx="11" cy="11" r="7" stroke="currentColor" stroke-width="1.8"/><path d="M16.7071 16.7071L21 21" stroke="currentColor" stroke-width="1.8" stroke-linecap="round"/></svg>
          <input id="globalSearch" type="text" placeholder="Поиск пользователей, чатов, логов..." oninput="filterGlobal(this.value)">
        </label>
        <button class="button-pill" onclick="toggleTheme()" id="themeToggle">Тема</button>
        <div class="profile-badge"><span class="profile-avatar">AD</span><div><strong>Admin</strong><span>@ssancodels_bot</span></div></div>
      </div>
    </header>
    <section class="card-grid">
      <article class="card card-compact">
        <div class="stat-card"><span class="label">Пользователей</span><span class="value">` + fmt.Sprint(userCount) + `</span></div>
        <span class="stat-pill">Уникальные</span>
      </article>
      <article class="card card-compact">
        <div class="stat-card"><span class="label">Подключённых чатов</span><span class="value">` + fmt.Sprint(chatCount) + `</span></div>
        <span class="stat-pill">Активных</span>
      </article>
      <article class="card card-compact">
        <div class="stat-card"><span class="label">Подписок</span><span class="value">` + fmt.Sprint(totalSubs) + `</span></div>
        <span class="stat-pill">Всего</span>
      </article>
      <article class="card card-compact">
        <div class="stat-card"><span class="label">Новых подписок</span><span class="value">` + fmt.Sprint(recentCount24h) + `</span></div>
        <span class="stat-pill">Последние 24 часа</span>
      </article>
      <article class="card card-compact">
        <div class="stat-card"><span class="label">Аптайм</span><span class="value">` + html.EscapeString(uptime) + `</span></div>
        <span class="stat-pill">С момента запуска</span>
      </article>
    </section>
    <section class="card chart-grid">
      <div class="chart-panel">
        <div class="card-compact"><div class="stat-card"><span class="label">Активность подписок</span><span class="value">` + fmt.Sprint(recentCount7d) + ` за 7 дней</span></div></div>
        <div class="chart-bars">` + activityBars + `</div>
      </div>
      <div class="card card-compact">
        <div class="stat-card"><span class="label">Последние действия</span></div>
        <div class="activity-list">` + recentActivities + `</div>
      </div>
    </section>
    <section class="card table-card">
      <div class="card-compact"><h2>Пользователи и чаты</h2></div>
      <div class="table-summary" style="display:flex;gap:18px;flex-wrap:wrap;margin-bottom:16px;color:var(--muted);font-size:0.95rem;"><span>Пользователей: <strong id="usersCount">` + fmt.Sprint(userCount) + `</strong></span><span>Чатов: <strong id="chatsCount">` + fmt.Sprint(chatCount) + `</strong></span></div>
      <table id="usersTable">
        <thead><tr><th>Username</th><th>Connection ID</th><th>Чатов</th><th>Последний чат</th><th>Детали</th></tr></thead>
        <tbody>` + m.buildUserRows(userIDs, userGroups) + `</tbody>
      </table>
    </section>
    <section class="card table-card">
      <div class="card-compact"><h2>Чаты подписчиков</h2></div>
      <table id="chatsTable">
        <thead><tr><th>Chat ID</th><th>Подписок</th><th>Детали</th></tr></thead>
        <tbody>` + m.buildChatRows(chatIDs, chatGroups) + `</tbody>
      </table>
    </section>
  </main>
</div>
<script>
function filterGlobal(query) {
	query = query.trim().toLowerCase();
	filterTable('usersTable', query);
	filterTable('chatsTable', query);
}
function filterTable(tableId, query) {
	const table = document.getElementById(tableId);
	if (!table) return;
	const rows = table.tBodies[0].rows;
	let visible = 0;
	for (let i = 0; i < rows.length; i++) {
		const text = rows[i].innerText.toLowerCase();
		const show = text.indexOf(query) !== -1;
		rows[i].style.display = show ? '' : 'none';
		if (show) visible++;
	}
	if (tableId === 'usersTable') document.getElementById('usersCount').innerText = visible;
	if (tableId === 'chatsTable') document.getElementById('chatsCount').innerText = visible;
}
function setTheme(theme) {
	if (theme === 'light') {
		document.body.classList.add('light');
		document.getElementById('themeToggle').innerText = '🌙';
	} else {
		document.body.classList.remove('light');
		document.getElementById('themeToggle').innerText = '☀️';
	}
	localStorage.setItem('dashboardTheme', theme);
}
function toggleTheme() {
	setTheme(document.body.classList.contains('light') ? 'dark' : 'light');
}
document.addEventListener('DOMContentLoaded', function() {
	const savedTheme = localStorage.getItem('dashboardTheme') || 'dark';
	setTheme(savedTheme);
});
</script>
</body>
</html>`)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(htmlPage.String()))
}

func (m *Monitor) handleSubscribers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	subs, err := m.service.GetAllBusinessSubscribers(ctx)
	if err != nil {
		log.Printf("Monitor: failed to load subscribers: %v", err)
		http.Error(w, "Failed to load subscriber data", http.StatusInternalServerError)
		return
	}

	sort.Slice(subs, func(i, j int) bool {
		return subs[i].CreatedAt.After(subs[j].CreatedAt)
	})

	rows := make([]string, 0, len(subs))
	for _, sub := range subs {
		rows = append(rows, fmt.Sprintf("<tr><td>%d</td><td>%s</td><td>%s</td><td>%s</td></tr>",
			sub.ChatID,
			html.EscapeString(sub.BusinessConnectionID),
			html.EscapeString(sub.Username),
			sub.CreatedAt.Format("2006-01-02 15:04:05"),
		))
	}

	htmlPage := fmt.Sprintf(`<!doctype html>
	<html lang="ru">
	<head>
	<meta charset="utf-8">
	<title>Subscribers</title>
	<style>
	:root{--bg:#070812;--panel:#0b0f17;--card:#0e1420;--muted:#9aa4b2;--accent:#06b6d4;--accent-2:#7c3aed;--border:rgba(255,255,255,0.04)}
	body{font-family:Inter,system-ui,-apple-system,"Segoe UI",Roboto,"Helvetica Neue",Arial;background:linear-gradient(180deg,#05060a 0%%,#071126 60%%);color:#e6eef8;margin:0;padding:28px}
	.container{max-width:1100px;margin:0 auto}
	.card{background:linear-gradient(180deg,rgba(255,255,255,0.01),rgba(255,255,255,0.005));border:1px solid var(--border);border-radius:12px;padding:16px;box-shadow:0 8px 30px rgba(2,6,23,0.6)}
	.back{color:var(--accent);display:inline-block;margin-bottom:12px}
	table{width:100%%;border-collapse:collapse;margin-top:8px}
	th,td{padding:12px;border-bottom:1px solid rgba(255,255,255,0.03);text-align:left}
	th{background:transparent;color:var(--muted);font-weight:600}
	code{background:rgba(255,255,255,0.03);padding:6px 8px;border-radius:8px;color:#e6eef8}
	.avatar{width:34px;height:34px;border-radius:50%%;display:inline-flex;align-items:center;justify-content:center;color:#fff;font-weight:700;margin-right:12px;background:linear-gradient(135deg,var(--accent-2),#4f46e5)}
	.username{font-weight:600;color:#e6eef8}
	.subtle{color:var(--muted);font-size:13px}
	.status-dot{width:10px;height:10px;border-radius:50%%;display:inline-block;margin-left:8px}
	.dot-online{background:#10b981}
	.dot-off{background:#ef4444}
	@media(max-width:600px){th,td{padding:8px}}
	</style>
	</head>
	<body>
	<div class="container card">
		<a class="back" href="%s">← назад</a>
		<h1 style="margin-top:6px">Список подписчиков</h1>
		<table>
			<thead><tr><th>Chat ID</th><th>Connection ID</th><th>Username</th><th>Created</th></tr></thead>
			<tbody>
			%s
			</tbody>
		</table>
	</div>
	</body>
	</html>`, m.authLink("/"), strings.Join(rows, "\n"))

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(htmlPage))
}

func (m *Monitor) handleChats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	subs, err := m.service.GetAllBusinessSubscribers(ctx)
	if err != nil {
		log.Printf("Monitor: failed to load subscribers: %v", err)
		http.Error(w, "Failed to load chat data", http.StatusInternalServerError)
		return
	}

	chatGroups := groupSubscribersByChat(subs)
	chatIDs := make([]int64, 0, len(chatGroups))
	for chatID := range chatGroups {
		chatIDs = append(chatIDs, chatID)
	}
	sort.Slice(chatIDs, func(i, j int) bool { return chatIDs[i] < chatIDs[j] })

	rows := make([]string, 0, len(chatIDs))
	for _, chatID := range chatIDs {
		rows = append(rows, fmt.Sprintf("<tr><td><code>%d</code></td><td>%d</td><td><a href=\"%s\">Просмотр</a></td></tr>",
			chatID,
			len(chatGroups[chatID]),
			m.authLink(fmt.Sprintf("/chat?chat_id=%d", chatID)),
		))
	}

	htmlPage := fmt.Sprintf(`<!doctype html>
	<html lang="ru">
	<head>
	<meta charset="utf-8">
	<title>Чаты</title>
	<style>
	:root{--bg:#070812;--panel:#0b0f17;--card:#0e1420;--muted:#9aa4b2;--accent:#06b6d4;--accent-2:#7c3aed;--border:rgba(255,255,255,0.04)}
	body{font-family:Inter,system-ui,-apple-system,"Segoe UI",Roboto,"Helvetica Neue",Arial;background:linear-gradient(180deg,#05060a 0%%,#071126 60%%);color:#e6eef8;margin:0;padding:28px}
	.container{max-width:1100px;margin:0 auto}
	.card{background:linear-gradient(180deg,rgba(255,255,255,0.01),rgba(255,255,255,0.005));border:1px solid var(--border);border-radius:12px;padding:16px;box-shadow:0 8px 30px rgba(2,6,23,0.6)}
	.back{color:var(--accent);display:inline-block;margin-bottom:12px}
	table{width:100%%;border-collapse:collapse;margin-top:8px}
	th,td{padding:12px;border-bottom:1px solid rgba(255,255,255,0.03);text-align:left}
	th{background:transparent;color:var(--muted);font-weight:600}
	code{background:rgba(255,255,255,0.03);padding:6px 8px;border-radius:8px;color:#e6eef8}
	@media(max-width:600px){th,td{padding:8px}}
	</style>
	</head>
	<body>
	<div class="container card">
		<a class="back" href="%s">← назад</a>
		<h1 style="margin-top:6px">Чаты</h1>
		<table>
			<thead><tr><th>Chat ID</th><th>Подписок</th><th>Детали</th></tr></thead>
			<tbody>
			%s
			</tbody>
		</table>
	</div>
	</body>
	</html>`,
		m.authLink("/"),
		strings.Join(rows, "\n"))

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(htmlPage))
}

func (m *Monitor) handleUser(w http.ResponseWriter, r *http.Request) {
	connectionID := r.URL.Query().Get("connection_id")
	if connectionID == "" {
		http.Error(w, "Missing connection_id", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	allSubs, err := m.service.GetAllBusinessSubscribers(ctx)
	if err != nil {
		log.Printf("Monitor: failed to load subscribers: %v", err)
		http.Error(w, "Failed to load user data", http.StatusInternalServerError)
		return
	}

	userSubs := make([]*domain.BusinessSubscriber, 0)
	for _, sub := range allSubs {
		if sub.BusinessConnectionID == connectionID {
			userSubs = append(userSubs, sub)
		}
	}

	if len(userSubs) == 0 {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	rows := make([]string, 0, len(userSubs))
	for _, sub := range userSubs {
		rows = append(rows, fmt.Sprintf("<tr><td>%d</td><td>%s</td><td>%s</td></tr>",
			sub.ChatID,
			html.EscapeString(sub.Username),
			sub.CreatedAt.Format("2006-01-02 15:04:05"),
		))
	}

	htmlPage := fmt.Sprintf(`<!doctype html>
	<html lang="ru">
	<head>
	<meta charset="utf-8">
	<title>User %s</title>
	<style>
	:root{--bg:#070812;--panel:#0b0f17;--card:#0e1420;--muted:#9aa4b2;--accent:#06b6d4;--accent-2:#7c3aed;--border:rgba(255,255,255,0.04)}
	body{font-family:Inter,system-ui,-apple-system,"Segoe UI",Roboto,"Helvetica Neue",Arial;background:linear-gradient(180deg,#05060a 0%%,#071126 60%%);color:#e6eef8;margin:0;padding:28px}
	.container{max-width:1100px;margin:0 auto}
	.card{background:linear-gradient(180deg,rgba(255,255,255,0.01),rgba(255,255,255,0.005));border:1px solid var(--border);border-radius:12px;padding:16px;box-shadow:0 8px 30px rgba(2,6,23,0.6)}
	.back{color:var(--accent);display:inline-block;margin-bottom:12px}
	table{width:100%%;border-collapse:collapse;margin-top:8px}
	th,td{padding:12px;border-bottom:1px solid rgba(255,255,255,0.03);text-align:left}
	th{background:transparent;color:var(--muted);font-weight:600}
	code{background:rgba(255,255,255,0.03);padding:6px 8px;border-radius:8px;color:#e6eef8}
	.avatar{width:34px;height:34px;border-radius:50%%;display:inline-flex;align-items:center;justify-content:center;color:#fff;font-weight:700;margin-right:12px;background:linear-gradient(135deg,var(--accent-2),#4f46e5)}
	.username{font-weight:600;color:#e6eef8}
	.subtle{color:var(--muted);font-size:13px}
	.status-dot{width:10px;height:10px;border-radius:50%%;display:inline-block;margin-left:8px}
	.dot-online{background:#10b981}
	.dot-off{background:#ef4444}
	@media(max-width:600px){th,td{padding:8px}}
	</style>
	</head>
	<body>
	<div class="container card">
	  <a class="back" href="%s">← назад</a>
	  <h1 style="margin-top:6px">User %s</h1>
	  <table>
	    <thead><tr><th>Chat ID</th><th>Username</th><th>Created</th></tr></thead>
	    <tbody>
	    %s
	    </tbody>
	  </table>
	</div>
	</body>
	</html>`,
		html.EscapeString(connectionID),
		m.authLink("/"),
		html.EscapeString(connectionID),
		strings.Join(rows, "\n"))

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(htmlPage))
}

func (m *Monitor) handleUsers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	subs, err := m.service.GetAllBusinessSubscribers(ctx)
	if err != nil {
		log.Printf("Monitor: failed to load subscribers: %v", err)
		http.Error(w, "Failed to load user data", http.StatusInternalServerError)
		return
	}

	userGroups := groupSubscribersByConnectionID(subs)
	userIDs := make([]string, 0, len(userGroups))
	for connectionID := range userGroups {
		userIDs = append(userIDs, connectionID)
	}
	sort.Strings(userIDs)

	rows := make([]string, 0, len(userIDs))
	for _, connectionID := range userIDs {
		rowData := userGroups[connectionID]
		rows = append(rows, m.buildUserRow(connectionID, rowData))
	}

	htmlPage := fmt.Sprintf(`<!doctype html>
	<html lang="ru">
	<head>
	<meta charset="utf-8">
	<title>Пользователи бота</title>
	<style>
	:root{--bg:#070812;--panel:#0b0f17;--card:#0e1420;--muted:#9aa4b2;--accent:#06b6d4;--accent-2:#7c3aed;--border:rgba(255,255,255,0.04)}
	body{font-family:Inter,system-ui,-apple-system,"Segoe UI",Roboto,"Helvetica Neue",Arial;background:linear-gradient(180deg,#05060a 0%%,#071126 60%%);color:#e6eef8;margin:0;padding:28px}
	.container{max-width:1100px;margin:0 auto}
	.card{background:linear-gradient(180deg,rgba(255,255,255,0.01),rgba(255,255,255,0.005));border:1px solid var(--border);border-radius:12px;padding:16px;box-shadow:0 8px 30px rgba(2,6,23,0.6)}
	.back{color:var(--accent);display:inline-block;margin-bottom:12px}
	table{width:100%%;border-collapse:collapse;margin-top:8px}
	th,td{padding:12px;border-bottom:1px solid rgba(255,255,255,0.03);text-align:left}
	th{background:transparent;color:var(--muted);font-weight:600}
	code{background:rgba(255,255,255,0.03);padding:6px 8px;border-radius:8px;color:#e6eef8}
	.avatar{width:34px;height:34px;border-radius:50%%;display:inline-flex;align-items:center;justify-content:center;color:#fff;font-weight:700;margin-right:12px;background:linear-gradient(135deg,var(--accent-2),#4f46e5)}
	.username{font-weight:600;color:#e6eef8}
	.subtle{color:var(--muted);font-size:13px}
	.status-dot{width:10px;height:10px;border-radius:50%%;display:inline-block;margin-left:8px}
	.dot-online{background:#10b981}
	.dot-off{background:#ef4444}
	@media(max-width:600px){th,td{padding:8px}}
	</style>
	</head>
	<body>
	<div class="container card">
		<a class="back" href="%s">← назад</a>
		<h1 style="margin-top:6px">Пользователи бота</h1>
		<table>
			<thead><tr><th>Username</th><th>Connection ID</th><th>Чатов</th><th>Последний чат</th><th>Детали</th></tr></thead>
			<tbody>
			%s
			</tbody>
		</table>
	</div>
	</body>
	</html>`,
		m.authLink("/"),
		strings.Join(rows, "\n"))

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(htmlPage))
}

func (m *Monitor) handleChat(w http.ResponseWriter, r *http.Request) {
	chatIDStr := r.URL.Query().Get("chat_id")
	if chatIDStr == "" {
		http.Error(w, "Missing chat_id", http.StatusBadRequest)
		return
	}

	chatID, err := strconv.ParseInt(chatIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid chat_id", http.StatusBadRequest)
		return
	}

	subs, err := m.service.GetBusinessSubscribersByChatID(r.Context(), chatID)
	if err != nil {
		log.Printf("Monitor: failed to load chat subscribers: %v", err)
		http.Error(w, "Failed to load chat data", http.StatusInternalServerError)
		return
	}

	messages, err := m.service.GetMessageHistory(r.Context(), chatID)
	if err != nil {
		log.Printf("Monitor: failed to load message history: %v", err)
		http.Error(w, "Failed to load message history", http.StatusInternalServerError)
		return
	}

	subRows := make([]string, 0, len(subs))
	for _, sub := range subs {
		subRows = append(subRows, fmt.Sprintf("<tr><td>%d</td><td>%s</td><td>%s</td><td>%s</td></tr>",
			sub.ChatID,
			html.EscapeString(sub.BusinessConnectionID),
			html.EscapeString(sub.Username),
			sub.CreatedAt.Format("2006-01-02 15:04:05"),
		))
	}

	msgRows := make([]string, 0, len(messages))
	for _, msg := range messages {
		text := html.EscapeString(msg.Text)
		if text == "" {
			text = fmt.Sprintf("[%s] %s", html.EscapeString(msg.MediaType), html.EscapeString(msg.MediaPath))
		}
		status := ""
		if msg.DeletedAt != nil {
			status = " (deleted)"
		}
		msgRows = append(msgRows, fmt.Sprintf("<tr><td>%d</td><td>%s</td><td>%s</td><td>%s</td><td>%s</td></tr>",
			msg.MessageID,
			html.EscapeString(msg.Username),
			msg.CreatedAt.Format("2006-01-02 15:04:05"),
			status,
			text,
		))
	}

	htmlPage := fmt.Sprintf(`<!doctype html>
	<html lang="ru">
	<head>
	<meta charset="utf-8">
	<title>Chat %d</title>
	<style>
	:root{--bg:#070812;--panel:#0b0f17;--card:#0e1420;--muted:#9aa4b2;--accent:#06b6d4;--accent-2:#7c3aed;--border:rgba(255,255,255,0.04)}
	body{font-family:Inter,system-ui,-apple-system,"Segoe UI",Roboto,"Helvetica Neue",Arial;background:linear-gradient(180deg,#05060a 0%%,#071126 60%%);color:#e6eef8;margin:0;padding:28px}
	.container{max-width:1100px;margin:0 auto}
	.card{background:linear-gradient(180deg,rgba(255,255,255,0.01),rgba(255,255,255,0.005));border:1px solid var(--border);border-radius:12px;padding:16px;box-shadow:0 8px 30px rgba(2,6,23,0.6)}
	.back{color:var(--accent);display:inline-block;margin-bottom:12px}
	table{width:100%%;border-collapse:collapse;margin-top:8px}
	th,td{padding:12px;border-bottom:1px solid rgba(255,255,255,0.03);text-align:left}
	th{background:transparent;color:var(--muted);font-weight:600}
	code{background:rgba(255,255,255,0.03);padding:6px 8px;border-radius:8px;color:#e6eef8}
	.avatar{width:34px;height:34px;border-radius:50%%;display:inline-flex;align-items:center;justify-content:center;color:#fff;font-weight:700;margin-right:12px;background:linear-gradient(135deg,var(--accent-2),#4f46e5)}
	.username{font-weight:600;color:#e6eef8}
	.subtle{color:var(--muted);font-size:13px}
	.status-dot{width:10px;height:10px;border-radius:50%%;display:inline-block;margin-left:8px}
	.dot-online{background:#10b981}
	.dot-off{background:#ef4444}
	@media(max-width:600px){th,td{padding:8px}}
	</style>
	</head>
	<body>
	<div class="container card">
		<a class="back" href="%s">← назад</a>
		<h1 style="margin-top:6px">Chat %d</h1>
		<h2 style="margin-top:12px">Message history (%d)</h2>
		<table>
			<thead><tr><th>Message ID</th><th>Username</th><th>Created</th><th>Status</th><th>Text / Media</th></tr></thead>
			<tbody>
			%s
			</tbody>
		</table>
		<h2 style="margin-top: 24px;">Subscribers in chat</h2>
		<table>
			<thead><tr><th>Chat ID</th><th>Connection ID</th><th>Username</th><th>Created</th></tr></thead>
			<tbody>
			%s
			</tbody>
		</table>
	</div>
	</body>
	</html>`,
		m.authLink("/"),
		chatID,
		chatID,
		len(messages),
		strings.Join(msgRows, "\n"),
		strings.Join(subRows, "\n"),
	)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(htmlPage))
}

func groupSubscribersByChat(subs []*domain.BusinessSubscriber) map[int64][]*domain.BusinessSubscriber {
	group := make(map[int64][]*domain.BusinessSubscriber)
	for _, sub := range subs {
		group[sub.ChatID] = append(group[sub.ChatID], sub)
	}
	return group
}

func groupSubscribersByConnectionID(subs []*domain.BusinessSubscriber) map[string][]*domain.BusinessSubscriber {
	group := make(map[string][]*domain.BusinessSubscriber)
	for _, sub := range subs {
		group[sub.BusinessConnectionID] = append(group[sub.BusinessConnectionID], sub)
	}
	return group
}

func (m *Monitor) buildUserRows(connectionIDs []string, groups map[string][]*domain.BusinessSubscriber) string {
	rows := make([]string, 0, len(connectionIDs))
	for _, connectionID := range connectionIDs {
		rows = append(rows, m.buildUserRow(connectionID, groups[connectionID]))
	}
	return strings.Join(rows, "\n")
}

func (m *Monitor) buildUserRow(connectionID string, subs []*domain.BusinessSubscriber) string {
	sort.Slice(subs, func(i, j int) bool { return subs[i].CreatedAt.After(subs[j].CreatedAt) })
	username := "-"
	if subs[0].Username != "" {
		username = subs[0].Username
	}
	lastChat := subs[0].ChatID
	return fmt.Sprintf("<tr><td>%s</td><td><code>%s</code></td><td>%d</td><td>%d</td><td><a href=\"%s\">Просмотр</a></td></tr>",
		html.EscapeString(username),
		html.EscapeString(connectionID),
		len(subs),
		lastChat,
		m.authLink(fmt.Sprintf("/user?connection_id=%s", url.QueryEscape(connectionID))),
	)
}

func (m *Monitor) authLink(path string) string {
	if m.adminToken == "" {
		return path
	}
	separator := "?"
	if strings.Contains(path, "?") {
		separator = "&"
	}
	return path + separator + "token=" + url.QueryEscape(m.adminToken)
}

func (m *Monitor) countRecentSubscriptions(subs []*domain.BusinessSubscriber, duration time.Duration) int {
	cutoff := time.Now().Add(-duration)
	count := 0
	for _, sub := range subs {
		if sub.CreatedAt.After(cutoff) {
			count++
		}
	}
	return count
}

func (m *Monitor) buildActivityBars(subs []*domain.BusinessSubscriber) string {
	now := time.Now()
	counts := make([]int, 7)
	labels := make([]string, 7)
	maxCount := 1
	for i := 0; i < 7; i++ {
		day := now.AddDate(0, 0, -6+i)
		labels[i] = day.Format("Mon")
		for _, sub := range subs {
			if sameDay(sub.CreatedAt, day) {
				counts[i]++
			}
		}
		if counts[i] > maxCount {
			maxCount = counts[i]
		}
	}

	bars := make([]string, 0, 7)
	for i, count := range counts {
		height := 42
		if maxCount > 0 {
			height = 42 + (count * 96 / maxCount)
		}
		bars = append(bars, fmt.Sprintf(`<div class="chart-bar" style="height:%dpx"><div class="bar-fill" style="height:%dpx"></div><span>%d</span><strong>%s</strong></div>`, height, height-12, count, labels[i]))
	}
	return strings.Join(bars, "\n")
}

func (m *Monitor) buildRecentActivities(subs []*domain.BusinessSubscriber) string {
	sort.Slice(subs, func(i, j int) bool { return subs[i].CreatedAt.After(subs[j].CreatedAt) })
	limit := 5
	if len(subs) < limit {
		limit = len(subs)
	}
	activities := make([]string, 0, limit)
	for i := 0; i < limit; i++ {
		sub := subs[i]
		title := "Новая подписка"
		if sub.Username != "" {
			title = fmt.Sprintf("Новый пользователь @%s", html.EscapeString(sub.Username))
		}
		activities = append(activities, fmt.Sprintf(`<div class="activity-item"><strong>%s</strong><span>Chat %d · %s</span></div>`, title, sub.ChatID, sub.CreatedAt.Format("02 Jan 15:04")))
	}
	if len(activities) == 0 {
		return `<div class="activity-item"><strong>Нет активности</strong><span>За последние дни</span></div>`
	}
	return strings.Join(activities, "\n")
}

func sameDay(a, b time.Time) bool {
	return a.Year() == b.Year() && a.Month() == b.Month() && a.Day() == b.Day()
}

func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm %ds", int(d.Minutes()), int(d.Seconds())%60)
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%dh %dm", int(d.Hours()), int(d.Minutes())%60)
	}
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	return fmt.Sprintf("%dd %dh", days, hours)
}

func (m *Monitor) buildChatRows(chatIDs []int64, groups map[int64][]*domain.BusinessSubscriber) string {
	rows := make([]string, 0, len(chatIDs))
	for _, chatID := range chatIDs {
		subs := groups[chatID]
		rows = append(rows, fmt.Sprintf("<tr><td><code>%d</code></td><td>%d</td><td><a href=\"%s\">Просмотр</a></td></tr>",
			chatID,
			len(subs),
			m.authLink(fmt.Sprintf("/chat?chat_id=%d", chatID)),
		))
	}
	return strings.Join(rows, "\n")
}
