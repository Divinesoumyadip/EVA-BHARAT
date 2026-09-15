// Single source of truth for the backend location.
// The Go backend is already deployed and is never modified from here.
const API_BASE_URL = 'https://eva-bharat-p6w2.onrender.com';

const TOKEN_KEY = 'eva_token';

const auth = {
  getToken() {
    return localStorage.getItem(TOKEN_KEY);
  },
  setToken(token) {
    localStorage.setItem(TOKEN_KEY, token);
  },
  clearToken() {
    localStorage.removeItem(TOKEN_KEY);
  },
  isLoggedIn() {
    return !!localStorage.getItem(TOKEN_KEY);
  },
};

// ApiError carries a status code so the UI can react appropriately
// (401 -> send to login, 400 -> show validation message, etc).
class ApiError extends Error {
  constructor(message, status) {
    super(message);
    this.status = status;
  }
}

async function apiRequest(path, { method = 'GET', body, auth: needsAuth = false } = {}) {
  const headers = { 'Content-Type': 'application/json' };

  if (needsAuth) {
    const token = auth.getToken();
    if (!token) {
      throw new ApiError('Not authenticated.', 401);
    }
    headers['Authorization'] = `Bearer ${token}`;
  }

  let response;
  try {
    response = await fetch(`${API_BASE_URL}${path}`, {
      method,
      headers,
      body: body !== undefined ? JSON.stringify(body) : undefined,
    });
  } catch (err) {
    throw new ApiError('Could not reach the server. Check your connection and try again.', 0);
  }

  let data = null;
  const text = await response.text();
  if (text) {
    try {
      data = JSON.parse(text);
    } catch (err) {
      data = null;
    }
  }

  if (!response.ok) {
    const message =
      (data && (data.error || data.message)) ||
      defaultMessageForStatus(response.status);

    if (response.status === 401) {
      auth.clearToken();
    }

    throw new ApiError(message, response.status);
  }

  return data;
}

function defaultMessageForStatus(status) {
  switch (status) {
    case 400:
      return 'That request was not valid. Check the fields and try again.';
    case 401:
      return 'Your session has expired. Please log in again.';
    case 404:
      return 'That ticket could not be found.';
    case 500:
      return 'Something went wrong on the server. Try again shortly.';
    default:
      return `Request failed (${status}).`;
  }
}

const api = {
  health() {
    return apiRequest('/health');
  },
  register(email, password) {
    return apiRequest('/auth/register', { method: 'POST', body: { email, password } });
  },
  login(email, password) {
    return apiRequest('/auth/login', { method: 'POST', body: { email, password } });
  },
  createTicket(title, description) {
    return apiRequest('/tickets', { method: 'POST', body: { title, description }, auth: true });
  },
  listTickets() {
    return apiRequest('/tickets', { auth: true });
  },
  getTicket(id) {
    return apiRequest(`/tickets/${encodeURIComponent(id)}`, { auth: true });
  },
  updateTicketStatus(id, status) {
    return apiRequest(`/tickets/${encodeURIComponent(id)}/status`, {
      method: 'PATCH',
      body: { status },
      auth: true,
    });
  },
};
