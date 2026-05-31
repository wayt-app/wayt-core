import http from 'k6/http';
import { check, fail } from 'k6';

export function post(url, body, token) {
  const headers = { 'Content-Type': 'application/json' };
  if (token) headers['Authorization'] = `Bearer ${token}`;
  return http.post(url, JSON.stringify(body), { headers });
}

export function get(url, token) {
  const headers = {};
  if (token) headers['Authorization'] = `Bearer ${token}`;
  return http.get(url, { headers });
}

export function put(url, body, token) {
  const headers = { 'Content-Type': 'application/json' };
  if (token) headers['Authorization'] = `Bearer ${token}`;
  return http.put(url, JSON.stringify(body), { headers });
}

export function del(url, token) {
  const headers = {};
  if (token) headers['Authorization'] = `Bearer ${token}`;
  return http.del(url, null, { headers });
}

export function expectOK(res, tag) {
  const ok = check(res, {
    [`${tag}: status 200`]: (r) => r.status === 200,
    [`${tag}: has data`]:   (r) => r.json('data') !== undefined,
  });
  if (!ok) {
    console.error(`[FAIL] ${tag} — status=${res.status} body=${res.body}`);
  }
  return ok;
}

export function expectCreated(res, tag) {
  const ok = check(res, {
    [`${tag}: status 201`]: (r) => r.status === 201,
  });
  if (!ok) {
    console.error(`[FAIL] ${tag} — status=${res.status} body=${res.body}`);
  }
  return ok;
}

export function expectError(res, tag, expectedStatus) {
  return check(res, {
    [`${tag}: status ${expectedStatus}`]: (r) => r.status === expectedStatus,
  });
}

// Login to customer API, return token
export function loginCustomer(base, email, password) {
  const res = post(`${base}/api/customers/login`, { email, password });
  check(res, { 'customer login: 200': (r) => r.status === 200 });
  const token = res.json('data.token');
  if (!token) fail(`Customer login failed: ${res.body}`);
  return token;
}

// Login to owner API, return token
export function loginOwner(base, email, password) {
  const res = post(`${base}/api/owner/login`, { email, password });
  check(res, { 'owner login: 200': (r) => r.status === 200 });
  const token = res.json('data.token');
  if (!token) fail(`Owner login failed: ${res.body}`);
  return token;
}

// Login to admin API, return token
export function loginAdmin(base, username, password) {
  const res = post(`${base}/auth/login`, { username, password });
  check(res, { 'admin login: 200': (r) => r.status === 200 });
  const token = res.json('data.token');
  if (!token) fail(`Admin login failed: ${res.body}`);
  return token;
}

// Return tomorrow's date as YYYY-MM-DD
export function tomorrow() {
  const d = new Date();
  d.setDate(d.getDate() + 1);
  return d.toISOString().split('T')[0];
}
