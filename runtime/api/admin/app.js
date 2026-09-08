(() => {
  const $ = (id) => document.getElementById(id);
  const state = {
    token: sessionStorage.getItem("fv_token") || "",
    users: [],
    linkCache: null,
    timers: [],
  };

  function headers() {
    return { Authorization: "Bearer " + state.token, "Content-Type": "application/json" };
  }

  function toast(msg, kind) {
    const el = $("toast");
    el.textContent = msg;
    el.classList.remove("hide", "err", "ok");
    if (kind) el.classList.add(kind);
    clearTimeout(el._t);
    el._t = setTimeout(() => el.classList.add("hide"), 4000);
  }

  async function api(path, opts = {}) {
    let r;
    try {
      r = await fetch(path, { ...opts, headers: { ...headers(), ...(opts.headers || {}) } });
    } catch (e) {
      toast(String(e.message || e), "err");
      throw e;
    }
    if (r.status === 401) {
      toast("Session expired — login again", "err");
      logout(false);
      throw new Error("unauthorized");
    }
    const ct = r.headers.get("content-type") || "";
    if (ct.includes("application/json")) {
      const j = await r.json();
      if (!r.ok) {
        toast(j.error || r.statusText, "err");
        throw new Error(j.error || r.statusText);
      }
      return j;
    }
    if (!r.ok) {
      toast(r.statusText, "err");
      throw new Error(r.statusText);
    }
    return r;
  }

  function fmtBytes(n) {
    if (!n) return "0 B";
    const u = ["B", "KB", "MB", "GB", "TB"];
    let i = 0, x = n;
    while (x >= 1024 && i < u.length - 1) { x /= 1024; i++; }
    return x.toFixed(i ? 1 : 0) + " " + u[i];
  }

  function fmtDur(sec) {
    if (sec < 60) return sec + "s";
    if (sec < 3600) return Math.floor(sec / 60) + "m";
    if (sec < 86400) return Math.floor(sec / 3600) + "h";
    return Math.floor(sec / 86400) + "d";
  }

  function showDash(on) {
    $("login").classList.toggle("hide", on);
    $("dash").classList.toggle("hide", !on);
    $("btn-logout").classList.toggle("hide", !on);
    $("session-line").classList.toggle("hide", !on);
  }

  function logout(clearMsg) {
    state.token = "";
    sessionStorage.removeItem("fv_token");
    state.timers.forEach(clearInterval);
    state.timers = [];
    showDash(false);
    if (clearMsg !== false) toast("Logged out", "ok");
  }

  function renderProbeStrip(probes) {
    const el = $("probe-strip");
    if (!probes || !probes.length) {
      el.innerHTML = '<span class="muted">No probe data</span>';
      return;
    }
    el.innerHTML = probes.map((p) => {
      const cls = p.ok ? "ok" : "bad";
      const ms = p.ms != null ? " " + p.ms + "ms" : "";
      return '<span class="pill ' + cls + '">' + (p.name || "?") + (p.ok ? " ✓" : " ✗") + ms + "</span>";
    }).join("");
  }

  async function refreshSession() {
    try {
      const s = await api("/api/session");
      $("session-line").textContent = "Session " + fmtDur(s.expires_in_sec) + " left";
      if (s.expires_in_sec < 600) toast("Session expires soon — renew via /session", "err");
    } catch (_) {}
  }

  async function refreshStatus() {
    const s = await api("/api/status");
    $("host-line").textContent = s.hostname + " · " + new Date(s.ts * 1000).toLocaleString();
    $("m-cpu").textContent = s.cpu_pct + "%";
    const mu = s.mem.used || 0, mt = s.mem.total || 1;
    $("m-ram").textContent = Math.round((100 * mu) / mt) + "%";
    $("m-ram").title = fmtBytes(mu) + " / " + fmtBytes(mt);
    const du = s.disk.used || 0, dt = s.disk.total || 1;
    $("m-disk").textContent = Math.round((100 * du) / dt) + "%";
    $("m-load").textContent = (s.loadavg || []).map((x) => x.toFixed(2)).join(" · ");
    if (s.net) {
      $("m-rx").textContent = fmtBytes(s.net.rx_bytes || 0);
      $("m-tx").textContent = fmtBytes(s.net.tx_bytes || 0);
    }
    $("collector-warn").classList.toggle("hide", !s.collector_stale);
    renderProbeStrip(s.probes || []);
    const ul = $("svc-list"); ul.innerHTML = "";
    Object.entries(s.services || {}).forEach(([name, st]) => {
      const li = document.createElement("li");
      li.innerHTML = "<span>" + name + '</span><span class="' + (st === "active" ? "ok" : "bad") + '">' + st + "</span>";
      ul.appendChild(li);
    });
    const cl = $("ct-list"); cl.innerHTML = "";
    (s.containers || []).forEach((c) => {
      const li = document.createElement("li");
      li.innerHTML = "<span>" + c.name + "</span><span class='muted'>" + c.status + "</span>";
      cl.appendChild(li);
    });
    if (!(s.containers || []).length) cl.innerHTML = "<li class='muted'>none</li>";
  }

  function renderUsers() {
    const q = ($("user-filter").value || "").toLowerCase();
    const tb = $("users-table").querySelector("tbody");
    tb.innerHTML = "";
    state.users.filter((u) => {
      if (!q) return true;
      return (u.name + " " + (u.note || "")).toLowerCase().includes(q);
    }).forEach((u) => {
      const tr = document.createElement("tr");
      tr.innerHTML = "<td>" + u.name + "</td><td class='" + (u.enabled ? "ok" : "bad") + "'>" +
        (u.enabled ? "on" : "off") + "</td><td class='muted'>" + (u.note || "") + "</td><td class='row-actions'></td>";
      const actions = tr.querySelector("td:last-child");
      const mk = (label, cls, fn) => {
        const b = document.createElement("button");
        b.textContent = label; b.className = cls; b.type = "button"; b.onclick = fn;
        actions.appendChild(b);
      };
      mk("Open", "primary", () => showUser(u.name));
      mk("On", "success", async () => {
        await api("/vpn/users/" + encodeURIComponent(u.name) + "/enable", { method: "POST", body: "{}" });
        toast("Enabled " + u.name, "ok"); refreshUsers();
      });
      mk("Off", "ghost", async () => {
        await api("/vpn/users/" + encodeURIComponent(u.name) + "/disable", { method: "POST", body: "{}" });
        toast("Disabled " + u.name, "ok"); refreshUsers();
      });
      mk("Revoke", "danger", async () => {
        if (!confirm("Revoke " + u.name + "?")) return;
        await api("/vpn/users/" + encodeURIComponent(u.name) + "/revoke", { method: "POST", body: "{}" });
        toast("Revoked " + u.name, "ok"); refreshUsers();
      });
      tb.appendChild(tr);
    });
  }

  async function refreshUsers() {
    const data = await api("/vpn/users");
    state.users = data.users || [];
    renderUsers();
  }

  async function showUser(name) {
    const link = await api("/vpn/users/" + encodeURIComponent(name) + "/link");
    state.linkCache = link;
    const u = state.users.find((x) => x.name === name);
    $("user-detail").classList.remove("hide");
    $("detail-title").textContent = name;
    $("detail-note").value = (u && u.note) || "";
    $("detail-sub").textContent = link.subscription || "";
    $("btn-copy-sub").onclick = async () => {
      try { await navigator.clipboard.writeText(link.subscription || ""); toast("Copied subscription", "ok"); } catch (_) { toast("Copy failed", "err"); }
    };
    $("btn-copy-vless").onclick = async () => {
      try { await navigator.clipboard.writeText(link.vless || ""); toast("Copied VLESS", "ok"); } catch (_) {}
    };
    $("btn-copy-hy2").onclick = async () => {
      try { await navigator.clipboard.writeText(link.hy2 || ""); toast("Copied HY2", "ok"); } catch (_) {}
    };
    $("btn-save-note").onclick = async () => {
      await api("/vpn/users/" + encodeURIComponent(name) + "/note", {
        method: "POST", body: JSON.stringify({ note: $("detail-note").value }),
      });
      toast("Note saved", "ok");
      refreshUsers();
    };
    const img = $("detail-qr");
    img.classList.add("hide");
    try {
      const r = await fetch("/vpn/users/" + encodeURIComponent(name) + "/qr", { headers: headers() });
      if (r.ok) { img.src = URL.createObjectURL(await r.blob()); img.classList.remove("hide"); }
    } catch (_) {}
  }

  function drawChart(pts) {
    const c = $("chart");
    const ctx = c.getContext("2d");
    const w = c.width, h = c.height, pad = 16;
    const bg = getComputedStyle(document.body).getPropertyValue("--bg").trim() || "#0d1218";
    ctx.fillStyle = bg;
    ctx.fillRect(0, 0, w, h);
    if (!pts.length) return;
    ctx.strokeStyle = getComputedStyle(document.body).getPropertyValue("--border").trim() || "#2a3548";
    ctx.lineWidth = 1;
    for (let i = 0; i <= 4; i++) {
      const y = pad + ((h - pad * 2) * i) / 4;
      ctx.beginPath(); ctx.moveTo(pad, y); ctx.lineTo(w - pad, y); ctx.stroke();
    }
    const draw = (key, color) => {
      const vals = pts.map((p) => Number(p[key] || 0));
      ctx.strokeStyle = color; ctx.lineWidth = 2; ctx.beginPath();
      vals.forEach((v, i) => {
        const x = pad + (i / Math.max(1, vals.length - 1)) * (w - pad * 2);
        const y = h - pad - (Math.min(100, v) / 100) * (h - pad * 2);
        if (i === 0) ctx.moveTo(x, y); else ctx.lineTo(x, y);
      });
      ctx.stroke();
    };
    draw("cpu_pct", "#3b82f6");
    draw("mem_pct", "#22c55e");
  }

  async function refreshHistory() {
    const h = await api("/api/metrics/history?limit=180");
    drawChart(h.points || []);
  }

  async function refreshProbes() {
    const data = await api("/api/probes");
    const up = data.uptime || {};
    const tb = $("probes-table").querySelector("tbody");
    tb.innerHTML = "";
    (data.probes || []).forEach((p) => {
      const u = up[p.name] || {};
      const tr = document.createElement("tr");
      const detail = p.error || (p.code != null ? "HTTP " + p.code : (p.via || ""));
      tr.innerHTML = "<td>" + (p.name || "") + "</td><td class='" + (p.ok ? "ok" : "bad") + "'>" +
        (p.ok ? "ok" : "fail") + "</td><td>" + (p.ms != null ? p.ms : "—") + "</td><td>" +
        (u.uptime_pct != null ? u.uptime_pct + "%" : "—") + "</td><td class='muted'>" + detail + "</td>";
      tb.appendChild(tr);
    });
    if (!(data.probes || []).length) tb.innerHTML = "<tr><td colspan='5' class='muted'>No data</td></tr>";
    renderProbeStrip(data.probes || []);
  }

  async function loadSettings() {
    const cfg = await api("/api/probes/config");
    const al = cfg.alerts || {};
    $("al-cpu").value = al.cpu_pct ?? 90;
    $("al-mem").value = al.mem_pct ?? 92;
    $("al-disk").value = al.disk_pct ?? 90;
    $("al-cd").value = al.cooldown_sec ?? 1800;
    $("al-svc").checked = al.service_not_active !== false;
    $("al-probe").checked = al.probe_fail !== false;
    $("probes-json").value = JSON.stringify(cfg.probes || [], null, 2);
  }

  async function saveAlerts() {
    const cfg = await api("/api/probes/config");
    cfg.alerts = {
      cpu_pct: Number($("al-cpu").value),
      mem_pct: Number($("al-mem").value),
      disk_pct: Number($("al-disk").value),
      cooldown_sec: Number($("al-cd").value),
      service_not_active: $("al-svc").checked,
      probe_fail: $("al-probe").checked,
    };
    await api("/api/probes/config", { method: "POST", body: JSON.stringify(cfg) });
    toast("Alerts saved", "ok");
  }

  async function saveProbes() {
    const cfg = await api("/api/probes/config");
    try {
      cfg.probes = JSON.parse($("probes-json").value);
    } catch (e) {
      toast("Invalid JSON", "err");
      return;
    }
    await api("/api/probes/config", { method: "POST", body: JSON.stringify(cfg) });
    toast("Probes saved", "ok");
  }

  function setTab(name) {
    document.querySelectorAll(".tab").forEach((b) => b.classList.toggle("active", b.dataset.tab === name));
    ["overview", "users", "metrics", "probes", "settings"].forEach((t) => {
      $("tab-" + t).classList.toggle("hide", t !== name);
    });
    if (name === "users") refreshUsers().catch(() => {});
    if (name === "metrics") refreshHistory().catch(() => {});
    if (name === "probes") refreshProbes().catch(() => {});
    if (name === "overview") refreshStatus().catch(() => {});
    if (name === "settings") loadSettings().catch(() => {});
  }

  function startTimers() {
    state.timers.forEach(clearInterval);
    state.timers = [
      setInterval(() => {
        if ($("dash").classList.contains("hide")) return;
        refreshStatus().catch(() => {});
        refreshSession().catch(() => {});
      }, 20000),
      setInterval(() => {
        if ($("dash").classList.contains("hide")) return;
        if (!$("tab-probes").classList.contains("hide")) refreshProbes().catch(() => {});
      }, 30000),
    ];
  }

  $("btn-login").onclick = async () => {
    state.token = $("token").value.trim();
    $("login-err").textContent = "";
    if (!state.token) { $("login-err").textContent = "Token required"; return; }
    try {
      await api("/api/status");
      sessionStorage.setItem("fv_token", state.token);
      showDash(true);
      setTab("overview");
      startTimers();
      refreshSession();
      toast("Signed in", "ok");
    } catch (e) {
      $("login-err").textContent = e.message || "Login failed";
    }
  };

  $("btn-logout").onclick = () => logout();
  document.querySelectorAll(".tab").forEach((b) => { b.onclick = () => setTab(b.dataset.tab); });
  $("btn-refresh-users").onclick = () => refreshUsers().catch(() => {});
  $("btn-refresh-probes").onclick = () => refreshProbes().catch(() => {});
  $("user-filter").oninput = () => renderUsers();
  $("btn-save-alerts").onclick = () => saveAlerts().catch(() => {});
  $("btn-save-probes").onclick = () => saveProbes().catch(() => {});
  $("btn-add").onclick = async () => {
    const name = $("new-name").value.trim(), note = $("new-note").value.trim();
    if (!name) return;
    await api("/vpn/users", { method: "POST", body: JSON.stringify({ name, note }) });
    $("new-name").value = ""; $("new-note").value = "";
    toast("User created", "ok");
    await refreshUsers();
    await showUser(name);
  };

  if (state.token) {
    api("/api/status")
      .then(() => { showDash(true); setTab("overview"); startTimers(); refreshSession(); })
      .catch(() => { sessionStorage.removeItem("fv_token"); state.token = ""; });
  }
})();
