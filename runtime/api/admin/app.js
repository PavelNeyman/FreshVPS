(() => {
  const $ = (s) => document.querySelector(s);
  const state = { token: localStorage.getItem('nd_token') || '', lang: localStorage.getItem('nd_lang') || 'en' };
  const i18n = {
    en: { login_title: 'Operator sign-in', login_hint: 'Token: netductor vpn session 72', sign_in: 'Sign in' },
    ru: { login_title: 'Вход оператора', login_hint: 'Токен: netductor vpn session 72', sign_in: 'Войти' },
  };
  function t(k) { return (i18n[state.lang] || i18n.en)[k] || k; }
  function applyI18n() {
    document.querySelectorAll('[data-i18n]').forEach((el) => { el.textContent = t(el.dataset.i18n); });
  }
  function toast(msg) {
    const el = $('#toast'); el.textContent = msg; el.classList.remove('hide');
    setTimeout(() => el.classList.add('hide'), 2500);
  }
  async function api(path, opts = {}) {
    const headers = Object.assign({ 'Content-Type': 'application/json' }, opts.headers || {});
    if (state.token) headers['Authorization'] = 'Bearer ' + state.token;
    const r = await fetch(path, { ...opts, headers });
    if (r.status === 401) { logout(); throw new Error('unauthorized'); }
    return r;
  }
  function logout() {
    state.token = ''; localStorage.removeItem('nd_token');
    $('#dash').classList.add('hide'); $('#login').classList.remove('hide');
    $('#btn-logout').classList.add('hide');
  }
  function showDash() {
    $('#login').classList.add('hide'); $('#dash').classList.remove('hide');
    $('#btn-logout').classList.remove('hide');
    refreshAll();
  }
  function tab(name) {
    document.querySelectorAll('.tab').forEach((b) => b.classList.toggle('active', b.dataset.tab === name));
    document.querySelectorAll('.tab-panel').forEach((p) => p.classList.toggle('hide', p.id !== 'tab-' + name));
  }
  async function refreshOverview() {
    try {
      const st = await (await api('/api/status')).json();
      $('#host-line').textContent = st.hostname || st.host || '';
      const m = await (await api('/api/metrics')).json().catch(() => ({}));
      const mem = m.mem_pct != null ? m.mem_pct + '%' : '—';
      const disk = m.disk_pct != null ? m.disk_pct + '%' : '—';
      $('#m-cpu').textContent = m.cpu_pct != null ? m.cpu_pct + '%' : (m.cpu || '—');
      $('#m-ram').textContent = mem;
      $('#m-disk').textContent = disk;
      $('#m-load').textContent = (m.load && (m.load['1'] || m.load[0])) || '—';
      $('#m-rx').textContent = m.net_rx || m.rx || '—';
      $('#m-tx').textContent = m.net_tx || m.tx || '—';
      if (m.ts && Date.now()/1000 - m.ts > 180) $('#collector-warn').classList.remove('hide');
      else $('#collector-warn').classList.add('hide');
      const pr = await (await api('/api/probes')).json().catch(() => []);
      const strip = $('#probe-strip'); strip.innerHTML = '';
      (Array.isArray(pr) ? pr : (pr.probes || [])).forEach((p) => {
        const s = document.createElement('span');
        s.className = 'pill ' + (p.ok ? 'ok' : 'bad');
        s.textContent = (p.name || p.id || 'probe') + (p.ok ? ' ✓' : ' ✗');
        strip.appendChild(s);
      });
    } catch (e) { console.warn(e); }
  }
  async function refreshUsers() {
    const r = await api('/vpn/users');
    const data = await r.json();
    const users = data.users || data || [];
    const tb = $('#users-table tbody'); tb.innerHTML = '';
    users.forEach((u) => {
      const tr = document.createElement('tr');
      tr.innerHTML = `<td>${u.name}</td><td>${u.enabled ? 'on' : 'off'}</td><td>${u.note || ''}</td>
        <td><button class="ghost btn-open" data-name="${u.name}">Open</button>
        <button class="ghost btn-dis" data-name="${u.name}">${u.enabled ? 'Disable' : 'Enable'}</button></td>`;
      tb.appendChild(tr);
    });
    tb.querySelectorAll('.btn-open').forEach((b) => b.onclick = () => openUser(b.dataset.name));
    tb.querySelectorAll('.btn-dis').forEach((b) => b.onclick = async () => {
      await api('/vpn/users/' + b.dataset.name + '/' + (b.textContent === 'Disable' ? 'disable' : 'enable'), { method: 'POST' });
      refreshUsers();
    });
  }
  async function openUser(name) {
    $('#user-detail').classList.remove('hide');
    $('#detail-title').textContent = name;
    const r = await api('/vpn/users/' + name + '/subscription');
    const text = await r.text();
    $('#detail-sub').textContent = text;
    const img = $('#detail-qr');
    img.classList.add('hide');
    try {
      const qr = await api('/vpn/users/' + name + '/qr');
      if (qr.ok) {
        const blob = await qr.blob();
        img.src = URL.createObjectURL(blob);
        img.classList.remove('hide');
      }
    } catch (_) {}
    $('#btn-copy-sub').onclick = () => { navigator.clipboard.writeText(text); toast('copied'); };
  }
  async function refreshMetrics() {
    const m = await (await api('/api/metrics')).json();
    $('#metrics-raw').textContent = JSON.stringify(m, null, 2);
  }
  async function refreshPending() {
    try {
      const data = await (await api('/api/edge/pending')).json();
      const list = data.pending || [];
      let box = document.getElementById('edge-pending');
      if (!box) {
        box = document.createElement('div');
        box.id = 'edge-pending';
        box.className = 'card';
        const routers = document.getElementById('tab-routers');
        if (routers) routers.prepend(box);
      }
      if (!list.length) { box.innerHTML = '<div class="muted">No pending devices</div>'; return; }
      box.innerHTML = '<h3>Pending approval</h3>' + list.map((d) => {
        const id = d.device_id || '';
        return `<div class="row"><code>${id}</code> ${d.board||''} ${d.wan_ip||''} ${d.hostname||''}
          <button class="primary btn-appr" data-id="${id}">Approve</button>
          <button class="ghost btn-deny" data-id="${id}">Deny</button></div>`;
      }).join('');
      box.querySelectorAll('.btn-appr').forEach((b) => b.onclick = async () => {
        await api('/api/edge/approve', { method: 'POST', body: JSON.stringify({ device_id: b.dataset.id }) });
        toast('approved'); refreshPending(); refreshRouters();
      });
      box.querySelectorAll('.btn-deny').forEach((b) => b.onclick = async () => {
        await api('/api/edge/deny', { method: 'POST', body: JSON.stringify({ device_id: b.dataset.id }) });
        toast('denied'); refreshPending();
      });
    } catch (e) { console.warn(e); }
  }
  async function refreshRouters() {
    refreshPending();
    const data = await (await api('/api/edge/devices')).json();
    const list = data.devices || data || [];
    const tb = $('#routers-table tbody'); tb.innerHTML = '';
    list.forEach((d) => {
      const tr = document.createElement('tr');
      const id = d.device_id || d.id || '';
      tr.innerHTML = `<td>${id}</td><td>${d.hostname || ''}</td><td>${d.last_seen || ''}</td><td>${d.wan_ip || ''}</td>
        <td><button class="ghost btn-ping" data-id="${id}">ping</button></td>`;
      tb.appendChild(tr);
    });
    tb.querySelectorAll('.btn-ping').forEach((b) => b.onclick = async () => {
      await api('/api/edge/cmd', { method: 'POST', body: JSON.stringify({ device_id: b.dataset.id, action: 'ping' }) });
      toast('queued');
      setTimeout(refreshRouters, 1500);
    });
    try {
      const res = await (await api('/api/edge/results')).json();
      let box = document.getElementById('edge-results');
      if (!box) {
        box = document.createElement('pre');
        box.id = 'edge-results';
        box.className = 'code';
        $('#tab-routers').appendChild(box);
      }
      box.textContent = JSON.stringify(res.results || res, null, 2);
      // per-device backups for first device
      if (list.length) {
        const id0 = list[0].device_id || list[0].id;
        const bk = await (await api('/api/edge/backups?device_id=' + encodeURIComponent(id0))).json();
        let bb = document.getElementById('edge-backups');
        if (!bb) {
          bb = document.createElement('pre');
          bb.id = 'edge-backups';
          bb.className = 'code';
          $('#tab-routers').appendChild(bb);
        }
        bb.textContent = 'backups ' + id0 + ':\n' + JSON.stringify(bk.backups || bk, null, 2);
      }
    } catch (_) {}
  }
  async function refreshProbes() {
    const p = await (await api('/api/probes')).json();
    $('#probes-raw').textContent = JSON.stringify(p, null, 2);
  }
  function refreshAll() {
    refreshOverview(); refreshUsers(); refreshMetrics(); refreshRouters(); refreshProbes();
  }

  $('#btn-login').onclick = async () => {
    state.token = $('#token').value.trim();
    try {
      const r = await api('/api/session');
      if (!r.ok) throw new Error('bad token');
      localStorage.setItem('nd_token', state.token);
      showDash();
    } catch (e) { $('#login-err').textContent = 'Invalid session'; }
  };
  $('#btn-logout').onclick = logout;
  $('#btn-lang').onclick = () => {
    state.lang = state.lang === 'en' ? 'ru' : 'en';
    localStorage.setItem('nd_lang', state.lang); applyI18n();
  };
  document.querySelectorAll('.tab').forEach((b) => b.onclick = () => tab(b.dataset.tab));
  $('#btn-refresh-users').onclick = refreshUsers;
  $('#btn-add-user').onclick = async () => {
    const name = $('#new-user').value.trim();
    const note = $('#new-note').value.trim();
    if (!name) return;
    await api('/vpn/users', { method: 'POST', body: JSON.stringify({ name, note }) });
    $('#new-user').value = ''; refreshUsers(); toast('added');
  };
  applyI18n();
  if (state.token) showDash();
})();
