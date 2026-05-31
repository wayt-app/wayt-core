/**
 * Flow: Admin operations
 * 1. Login as admin
 * 2. Get dashboard stats
 * 3. List bookings
 * 4. List restaurants
 */
import { check, sleep } from 'k6';
import { BASE, CREDS } from './config.js';
import { get, loginAdmin } from './helpers.js';

export const options = { vus: 1, iterations: 1 };

export default function () {
  const base = BASE.admin;
  const { username, password } = CREDS.admin;

  // 1. Login
  const token = loginAdmin(base, username, password);
  console.log('✅ Admin login OK');
  sleep(0.5);

  // 2. Dashboard stats for branch 4
  const today = new Date().toISOString().split('T')[0];
  const statsRes = get(`${base}/internal/branches/4/dashboard?date=${today}`, token);
  check(statsRes, {
    'admin: dashboard stats 200':  (r) => r.status === 200,
    'admin: has counts':           (r) => r.json('data.counts') !== null,
  });
  console.log(`✅ Dashboard stats OK — total_today=${statsRes.json('data.total_today')}`);
  sleep(0.5);

  // 3. List bookings
  const bookingsRes = get(`${base}/internal/bookings?branch_id=4&date=${today}`, token);
  check(bookingsRes, {
    'admin: list bookings 200': (r) => r.status === 200,
  });
  const bookingCount = Array.isArray(bookingsRes.json('data')) ? bookingsRes.json('data').length : 0;
  console.log(`✅ List bookings OK — count=${bookingCount}`);
  sleep(0.5);

  // 4. List restaurants
  const restaurantsRes = get(`${base}/internal/restaurants`, token);
  check(restaurantsRes, {
    'admin: list restaurants 200': (r) => r.status === 200,
    'admin: has restaurants':      (r) => Array.isArray(r.json('data')) && r.json('data').length > 0,
  });
  console.log(`✅ List restaurants OK — count=${restaurantsRes.json('data').length}`);
}
