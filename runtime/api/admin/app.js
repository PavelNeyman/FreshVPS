(() => {
  const $ = (id) => document.getElementById(id);
  const state = { token: sessionStorage.getItem("fv_token") || "" };

  function headers() {
    return {
      Authorization: "Bearer " + state.token,
      "Content-Type": "application/json",
    };
  }

  async function api(path, opts = {}) {
    const r = await fetch(path, {
      ...opts,
      headers: { ...headers(), ...(opts.headers || {}) },
    });
    if (r.status === 401) throw new Error("unauthorized");
    const ct = r.headers.get("content-type") || "";
    if (ct.includes("application/json")) {
      const j = await r.json();
      if (!r.ok) throw new Error(j.error || r.statusText);
      return j;
    }
    if (!r.ok) throw new Error(r.statusText);
    return r;
  }

  function fmtBytes(n) {
    if (!n) return "0";
    const u = ["B", "KB", "MB", "GB", "TB"];
    let i = 0;
    let x = n;
    while (x >= 1024 && i < u.length - 1) {
      x /= 1024;
      i++;
    }
    return x.toFixed(i ? 1 : 0) + " " + u[i];
  }

  function showDash(on) {
    $("login").classList.toggle("hide", on);
    $("dash").classList.toggle("hide", !on);
    $("btn-logout").classList.toggle("hide", !on);
  }

  async function refreshStatus() {
    const s = await api("/api/status");
    $("host-line").textContent = s.hostname + " · " + new Date(s.ts * 1000).toLocaleString();
    $("m-cpu").textContent = s.cpu_pct + "%";
    const mu = s.mem.used || 0;
    const mt = s.mem.total || 1;
    $("m-ram").textContent = Math.round((100 * mu) / mt) + "%";
    $("m-ram").title = fmtBytes(mu) + " / " + fmtBytes(mt);
    const du = s.disk.used || 0;
    const dt = s.disk.total || 1;
    $("m-disk").textContent = Math.round((100 * du) / dt) + "%";
    $("m-load").textContent = (s.loadavg || []).map((x) => x.toFixed(2)).join(" · ");
    const ul = $("svc-list");
    ul.innerHTML = "";
    Object.entries(s.services || {}).forEach(([name, st]) => {
      const li = document.createElement("li");
      li.innerHTML =
        "<span>" +
        name +
        '</span><span class="' +
        (st === "active" ? "ok" : "bad") +
        '">' +
        st +
        "</span>";
      ul.appendChild(li);
    });
    const cl = $("ct-list");
    cl.innerHTML = "";
    (s.containers || []).forEach((c) => {
      const li = document.createElement("li");
      li.innerHTML = "<span>" + c.name + "</span><span class='muted'>" + c.status + "</span>";
      cl.appendChild(li);
    });
    if (!(s.containers || []).length) {
      cl.innerHTML = "<li class='muted'>none</li>";
    }
  }

  async function refreshUsers() {
    const data = await api("/vpn/users");
    const tb = $("users-table").querySelector("tbody");
    tb.innerHTML = "";
    (data.users || []).forEach((u) => {
      const tr = document.createElement("tr");
      tr.innerHTML =
        "<td>" +
        u.name +
        "</td><td class='" +
        (u.enabled ? "ok" : "bad") +
        "'>" +
        (u.enabled ? "on" : "off") +
        "</td><td class='row-actions'></td>";
      const actions = tr.querySelector("td:last-child");
      const mk = (label, cls, fn) => {
        const b = document.createElement("button");
        b.textContent = label;
        b.className = cls;
        b.type = "button";
        b.onclick = fn;
        actions.appendChild(b);
      };
      mk("Link", "primary", () => showUser(u.name));
      mk("On", "success", async () => {
        await api("/vpn/users/" + encodeURIComponent(u.name) + "/enable", { method: "POST", body: "{}" });
        refreshUsers();
      });
      mk("Off", "ghost", async () => {
        await api("/vpn/users/" + encodeURIComponent(u.name) + "/disable", { method: "POST", body: "{}" });
        refreshUsers();
      });
      mk("Revoke", "danger", async () => {
        if (!confirm("Revoke " + u.name + "?")) return;
        await api("/vpn/users/" + encodeURIComponent(u.name) + "/revoke", { method: "POST", body: "{}" });
        refreshUsers();
      });
      tb.appendChild(tr);
    });
  }

  async function showUser(name) {
    const link = await api("/vpn/users/" + encodeURIComponent(name) + "/link");
    $("user-detail").classList.remove("hide");
    $("detail-title").textContent = name;
    $("detail-sub").textContent = link.subscription || "";
    $("btn-copy-sub").onclick = async () => {
      try {
        await navigator.clipboard.writeText(link.subscription || "");
      } catch (_) {}
    };
    const img = $("detail-qr");
    img.classList.add("hide");
    try {
      const r = await fetch("/vpn/users/" + encodeURIComponent(name) + "/qr", {
        headers: headers(),
      });
      if (r.ok) {
        const blob = await r.blob();
        img.src = URL.createObjectURL(blob);
        img.classList.remove("hide");
      }
    } catch (_) {}
  }

  async function refreshHistory() {
    try {
      const h = await api("/api/metrics/history");
      const pts = h.points || [];
      $("hist-raw").textContent = pts.length
        ? pts
            .slice(-15)
            .map((p) => new Date(p.ts * 1000).toISOString() + " cpu=" + p.cpu_pct + " mem%=" + (p.mem_pct ?? ""))
            .join("\n")
        : "No history yet.";
      drawChart(pts);
    } catch (e) {
      $("hist-raw").textContent = String(e.message || e);
    }
  }

  function drawChart(pts) {
    const c = $("chart");
    const ctx = c.getContext("2d");
    const w = c.width;
    const h = c.height;
    ctx.clearRect(0, 0, w, h);
    if (!pts.length) return;
    const vals = pts.map((p) => p.cpu_pct || 0);
    const max = Math.max(100, ...vals);
    ctx.strokeStyle = "#3b82f6";
    ctx.lineWidth = 2;
    ctx.beginPath();
    vals.forEach((v, i) => {
      const x = (i / Math.max(1, vals.length - 1)) * (w - 20) + 10;
      const y = h - 10 - (v / max) * (h - 20);
      if (i === 0) ctx.moveTo(x, y);
      else ctx.lineTo(x, y);
    });
    ctx.stroke();
  }

  function setTab(name) {
    document.querySelectorAll(".tab").forEach((b) => {
      b.classList.toggle("active", b.dataset.tab === name);
    });
    ["overview", "users", "metrics"].forEach((t) => {
      $("tab-" + t).classList.toggle("hide", t !== name);
    });
    if (name === "users") refreshUsers().catch(console.error);
    if (name === "metrics") refreshHistory().catch(console.error);
    if (name === "overview") refreshStatus().catch(console.error);
  }

  $("btn-login").onclick = async () => {
    state.token = $("token").value.trim();
    $("login-err").textContent = "";
    if (!state.token) {
      $("login-err").textContent = "Token required";
      return;
    }
    try {
      await api("/api/status");
      sessionStorage.setItem("fv_token", state.token);
      showDash(true);
      setTab("overview");
    } catch (e) {
      $("login-err").textContent = e.message || "Login failed";
    }
  };

  $("btn-logout").onclick = () => {
    state.token = "";
    sessionStorage.removeItem("fv_token");
    showDash(false);
  };

  document.querySelectorAll(".tab").forEach((b) => {
    b.onclick = () => setTab(b.dataset.tab);
  });

  $("btn-refresh-users").onclick = () => refreshUsers().catch(alert);
  $("btn-add").onclick = async () => {
    const name = $("new-name").value.trim();
    const note = $("new-note").value.trim();
    if (!name) return;
    await api("/vpn/users", {
      method: "POST",
      body: JSON.stringify({ name, note }),
    });
    $("new-name").value = "";
    $("new-note").value = "";
    await refreshUsers();
    await showUser(name);
  };

  if (state.token) {
    api("/api/status")
      .then(() => {
        showDash(true);
        setTab("overview");
      })
      .catch(() => {
        sessionStorage.removeItem("fv_token");
        state.token = "";
      });
  }
})();
