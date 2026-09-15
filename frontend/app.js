const appEl = document.getElementById('app');

const STATUS_LABELS = {
  open: 'open',
  in_progress: 'in progress',
  closed: 'closed',
};

const NEXT_STATUS = {
  open: 'in_progress',
  in_progress: 'closed',
  closed: null,
};

function escapeHtml(str) {
  const div = document.createElement('div');
  div.textContent = str == null ? '' : String(str);
  return div.innerHTML;
}

// The backend contract for field casing (e.g. created_at vs createdAt) isn't
// verified from here, so ticket fields are read defensively.
function normalizeTicket(raw) {
  return {
    id: raw.id ?? raw.ID ?? raw.Id ?? raw.ticket_id,
    title: raw.title ?? raw.Title ?? '',
    description: raw.description ?? raw.Description ?? '',
    status: raw.status ?? raw.Status ?? 'open',
    createdAt: raw.created_at ?? raw.createdAt ?? raw.CreatedAt ?? null,
  };
}

function normalizeTicketList(data) {
  const list = Array.isArray(data) ? data : data?.tickets ?? data?.data ?? [];
  return list.map(normalizeTicket);
}

function formatDate(value) {
  if (!value) return '—';
  const d = new Date(value);
  if (isNaN(d.getTime())) return String(value);
  return d.toLocaleString(undefined, {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  });
}

function navigate(hash) {
  window.location.hash = hash;
}

function currentRoute() {
  const hash = window.location.hash || '#/login';
  const [path, param] = hash.replace('#/', '').split('/');
  return { path: path || 'login', param };
}

// ---------- Shell for logged-in pages ----------

function renderShell(contentHtml, activePage) {
  const links = [
    { key: 'dashboard', label: 'Tickets', href: '#/dashboard' },
    { key: 'new', label: 'New ticket', href: '#/tickets/new' },
  ];

  const navHtml = links
    .map(
      (l) =>
        `<a class="nav-link${l.key === activePage ? ' active' : ''}" href="${l.href}">${l.label}</a>`
    )
    .join('');

  appEl.innerHTML = `
    <div class="shell">
      <aside class="sidebar">
        <div class="sidebar-brand">EVA Bharat<strong>Ticket System</strong></div>
        <nav>${navHtml}</nav>
        <div class="sidebar-footer">
          <div class="sidebar-user" id="sidebar-user"></div>
          <button class="btn btn-danger" id="logout-btn">Log out</button>
        </div>
      </aside>
      <main class="main">${contentHtml}</main>
    </div>
  `;

  document.getElementById('logout-btn').addEventListener('click', () => {
    auth.clearToken();
    navigate('#/login');
  });

  const email = localStorage.getItem('eva_user_email');
  const userEl = document.getElementById('sidebar-user');
  if (userEl && email) userEl.textContent = email;
}

// ---------- Login ----------

function renderLogin() {
  appEl.innerHTML = `
    <div class="auth-shell">
      <div class="auth-card">
        <div class="auth-brand">EVA Bharat</div>
        <h1 class="auth-title">Log in</h1>
        <div id="login-error"></div>
        <form id="login-form">
          <div class="field">
            <label for="login-email">Email</label>
            <input id="login-email" type="email" autocomplete="email" required />
          </div>
          <div class="field">
            <label for="login-password">Password</label>
            <input id="login-password" type="password" autocomplete="current-password" required />
          </div>
          <button class="btn" type="submit" id="login-submit">Log in</button>
        </form>
        <div class="auth-switch">
          No account? <a href="#/register">Register</a>
        </div>
      </div>
    </div>
  `;

  document.getElementById('login-form').addEventListener('submit', async (e) => {
    e.preventDefault();
    const email = document.getElementById('login-email').value.trim();
    const password = document.getElementById('login-password').value;
    const errorEl = document.getElementById('login-error');
    const submitBtn = document.getElementById('login-submit');

    errorEl.innerHTML = '';
    submitBtn.disabled = true;
    submitBtn.innerHTML = '<span class="spinner"></span>';

    try {
      const result = await api.login(email, password);
      const token = result?.token ?? result?.access_token ?? result?.jwt;
      if (!token) {
        throw new ApiError('Login succeeded but no token was returned by the server.', 500);
      }
      auth.setToken(token);
      localStorage.setItem('eva_user_email', email);
      navigate('#/dashboard');
    } catch (err) {
      errorEl.innerHTML = `<div class="banner banner-error">${escapeHtml(err.message)}</div>`;
      submitBtn.disabled = false;
      submitBtn.textContent = 'Log in';
    }
  });
}

// ---------- Register ----------

function renderRegister() {
  appEl.innerHTML = `
    <div class="auth-shell">
      <div class="auth-card">
        <div class="auth-brand">EVA Bharat</div>
        <h1 class="auth-title">Create an account</h1>
        <div id="register-error"></div>
        <form id="register-form">
          <div class="field">
            <label for="reg-email">Email</label>
            <input id="reg-email" type="email" autocomplete="email" required />
          </div>
          <div class="field">
            <label for="reg-password">Password</label>
            <input id="reg-password" type="password" autocomplete="new-password" required minlength="8" />
          </div>
          <div class="field">
            <label for="reg-confirm">Confirm password</label>
            <input id="reg-confirm" type="password" autocomplete="new-password" required />
            <div class="field-error" id="confirm-error"></div>
          </div>
          <button class="btn" type="submit" id="register-submit">Register</button>
        </form>
        <div class="auth-switch">
          Already have an account? <a href="#/login">Log in</a>
        </div>
      </div>
    </div>
  `;

  document.getElementById('register-form').addEventListener('submit', async (e) => {
    e.preventDefault();
    const email = document.getElementById('reg-email').value.trim();
    const password = document.getElementById('reg-password').value;
    const confirm = document.getElementById('reg-confirm').value;
    const errorEl = document.getElementById('register-error');
    const confirmErrorEl = document.getElementById('confirm-error');
    const submitBtn = document.getElementById('register-submit');

    errorEl.innerHTML = '';
    confirmErrorEl.textContent = '';

    if (password !== confirm) {
      confirmErrorEl.textContent = 'Passwords do not match.';
      return;
    }

    submitBtn.disabled = true;
    submitBtn.innerHTML = '<span class="spinner"></span>';

    try {
      await api.register(email, password);
      // Registration succeeded; send the user to log in with their new credentials.
      navigate('#/login');
    } catch (err) {
      errorEl.innerHTML = `<div class="banner banner-error">${escapeHtml(err.message)}</div>`;
      submitBtn.disabled = false;
      submitBtn.textContent = 'Register';
    }
  });
}

// ---------- Dashboard ----------

async function renderDashboard() {
  renderShell(
    `
      <div class="page-header">
        <div>
          <h1 class="page-title">Tickets</h1>
          <div class="page-sub">Tickets you've created</div>
        </div>
        <a class="btn btn-secondary btn-inline" href="#/tickets/new">Create ticket</a>
      </div>
      <div id="dashboard-body">
        <div class="skeleton-row"></div>
        <div class="skeleton-row"></div>
        <div class="skeleton-row"></div>
      </div>
    `,
    'dashboard'
  );

  const bodyEl = document.getElementById('dashboard-body');

  try {
    const raw = await api.listTickets();
    const tickets = normalizeTicketList(raw);

    if (tickets.length === 0) {
      bodyEl.innerHTML = `
        <div class="empty-state">
          <p>No tickets yet.</p>
          <a class="btn btn-secondary btn-inline" href="#/tickets/new">Create your first ticket</a>
        </div>
      `;
      return;
    }

    bodyEl.innerHTML = `
      <table class="ticket-table">
        <thead>
          <tr>
            <th>Title</th>
            <th>Description</th>
            <th>Status</th>
            <th>Created</th>
          </tr>
        </thead>
        <tbody>
          ${tickets
            .map(
              (t) => `
            <tr class="ticket-row" data-id="${escapeHtml(t.id)}">
              <td class="ticket-title-cell">${escapeHtml(t.title)}</td>
              <td class="ticket-desc-cell">${escapeHtml(t.description)}</td>
              <td><span class="badge badge-${t.status}">${STATUS_LABELS[t.status] ?? t.status}</span></td>
              <td class="ticket-date-cell">${formatDate(t.createdAt)}</td>
            </tr>
          `
            )
            .join('')}
        </tbody>
      </table>
      <div class="ticket-cards">
        ${tickets
          .map(
            (t) => `
          <div class="ticket-card" data-id="${escapeHtml(t.id)}">
            <div class="ticket-card-top">
              <div class="ticket-card-title">${escapeHtml(t.title)}</div>
              <span class="badge badge-${t.status}">${STATUS_LABELS[t.status] ?? t.status}</span>
            </div>
            <div class="ticket-card-desc">${escapeHtml(t.description)}</div>
            <div class="ticket-card-date">${formatDate(t.createdAt)}</div>
          </div>
        `
          )
          .join('')}
      </div>
    `;

    bodyEl.querySelectorAll('[data-id]').forEach((el) => {
      el.addEventListener('click', () => navigate(`#/tickets/${el.dataset.id}`));
    });
  } catch (err) {
    if (err.status === 401) {
      navigate('#/login');
      return;
    }
    bodyEl.innerHTML = `<div class="banner banner-error">${escapeHtml(err.message)}</div>`;
  }
}

// ---------- Create ticket ----------

function renderCreateTicket() {
  renderShell(
    `
      <a class="back-link" href="#/dashboard">&larr; Back to tickets</a>
      <div class="page-header">
        <div>
          <h1 class="page-title">Create ticket</h1>
        </div>
      </div>
      <div class="form-card">
        <div id="create-error"></div>
        <form id="create-form">
          <div class="field">
            <label for="ticket-title">Title</label>
            <input id="ticket-title" type="text" required maxlength="120" />
          </div>
          <div class="field">
            <label for="ticket-description">Description</label>
            <textarea id="ticket-description" required></textarea>
          </div>
          <div class="form-actions">
            <button class="btn btn-inline" type="submit" id="create-submit">Create ticket</button>
            <a class="btn btn-secondary" href="#/dashboard">Cancel</a>
          </div>
        </form>
      </div>
    `,
    'new'
  );

  document.getElementById('create-form').addEventListener('submit', async (e) => {
    e.preventDefault();
    const title = document.getElementById('ticket-title').value.trim();
    const description = document.getElementById('ticket-description').value.trim();
    const errorEl = document.getElementById('create-error');
    const submitBtn = document.getElementById('create-submit');

    errorEl.innerHTML = '';
    submitBtn.disabled = true;
    submitBtn.innerHTML = '<span class="spinner"></span>';

    try {
      await api.createTicket(title, description);
      navigate('#/dashboard');
    } catch (err) {
      if (err.status === 401) {
        navigate('#/login');
        return;
      }
      errorEl.innerHTML = `<div class="banner banner-error">${escapeHtml(err.message)}</div>`;
      submitBtn.disabled = false;
      submitBtn.textContent = 'Create ticket';
    }
  });
}

// ---------- Ticket detail ----------

async function renderTicketDetail(id) {
  renderShell(
    `
      <a class="back-link" href="#/dashboard">&larr; Back to tickets</a>
      <div id="detail-body">
        <div class="skeleton-row"></div>
        <div class="skeleton-row"></div>
      </div>
    `,
    'dashboard'
  );

  const bodyEl = document.getElementById('detail-body');

  try {
    const raw = await api.getTicket(id);
    const ticket = normalizeTicket(raw);
    renderDetailBody(bodyEl, ticket);
  } catch (err) {
    if (err.status === 401) {
      navigate('#/login');
      return;
    }
    bodyEl.innerHTML = `<div class="banner banner-error">${escapeHtml(err.message)}</div>`;
  }
}

function renderDetailBody(bodyEl, ticket) {
  const next = NEXT_STATUS[ticket.status];

  bodyEl.innerHTML = `
    <div class="detail-card">
      <div class="detail-top">
        <div>
          <h1 class="detail-title">${escapeHtml(ticket.title)}</h1>
          <div class="detail-meta">#${escapeHtml(ticket.id)} &middot; created ${formatDate(ticket.createdAt)}</div>
        </div>
        <span class="badge badge-${ticket.status}">${STATUS_LABELS[ticket.status] ?? ticket.status}</span>
      </div>

      <div class="detail-label">Description</div>
      <div class="detail-desc">${escapeHtml(ticket.description)}</div>

      <div class="detail-label">Status</div>
      <div id="status-area">
        ${
          next
            ? `
          <div class="status-row">
            <button class="btn btn-secondary btn-inline" id="advance-status">
              Mark as ${STATUS_LABELS[next]}
            </button>
          </div>
          <div class="status-note">Tickets move forward only: open &rarr; in progress &rarr; closed.</div>
        `
            : `<div class="closed-note">This ticket is closed and cannot be reopened.</div>`
        }
      </div>
      <div id="status-error"></div>
    </div>
  `;

  const advanceBtn = document.getElementById('advance-status');
  if (advanceBtn) {
    advanceBtn.addEventListener('click', async () => {
      const errorEl = document.getElementById('status-error');
      errorEl.innerHTML = '';
      advanceBtn.disabled = true;
      advanceBtn.innerHTML = '<span class="spinner"></span>';

      try {
        await api.updateTicketStatus(ticket.id, next);
        const raw = await api.getTicket(ticket.id);
        renderDetailBody(bodyEl, normalizeTicket(raw));
      } catch (err) {
        if (err.status === 401) {
          navigate('#/login');
          return;
        }
        errorEl.innerHTML = `<div class="banner banner-error">${escapeHtml(err.message)}</div>`;
        advanceBtn.disabled = false;
        advanceBtn.textContent = `Mark as ${STATUS_LABELS[next]}`;
      }
    });
  }
}

// ---------- Router ----------

function router() {
  const { path, param } = currentRoute();
  const loggedIn = auth.isLoggedIn();

  const publicPaths = new Set(['login', 'register']);

  if (!publicPaths.has(path) && !loggedIn) {
    navigate('#/login');
    return;
  }
  if (publicPaths.has(path) && loggedIn) {
    navigate('#/dashboard');
    return;
  }

  switch (path) {
    case 'login':
      renderLogin();
      break;
    case 'register':
      renderRegister();
      break;
    case 'dashboard':
      renderDashboard();
      break;
    case 'tickets':
      if (param === 'new') {
        renderCreateTicket();
      } else if (param) {
        renderTicketDetail(param);
      } else {
        navigate('#/dashboard');
      }
      break;
    default:
      navigate(loggedIn ? '#/dashboard' : '#/login');
  }
}

window.addEventListener('hashchange', router);
window.addEventListener('DOMContentLoaded', router);
