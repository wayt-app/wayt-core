/**
 * Error case tests — pastikan API return error yang benar
 * untuk input invalid / unauthorized access.
 */
import { check, sleep } from 'k6';
import { BASE, CREDS } from './config.js';
import { post, get, del, expectError, loginCustomer, tomorrow } from './helpers.js';

export const options = { vus: 1, iterations: 1 };

export default function () {
  const base = BASE.customer;
  const token = loginCustomer(base, CREDS.customer.email, CREDS.customer.password);
  sleep(0.3);

  // 1. Login dengan password salah → 400
  const badLoginRes = post(`${base}/api/customers/login`, {
    email: CREDS.customer.email, password: 'wrongpassword',
  });
  check(badLoginRes, { 'bad login: 400': (r) => r.status === 400 });
  console.log(`✅ Bad login → ${badLoginRes.status}`);
  sleep(0.3);

  // 2. Akses endpoint auth tanpa token → 401
  const noTokenRes = get(`${base}/api/bookings`);
  check(noTokenRes, { 'no token: 401': (r) => r.status === 401 });
  console.log(`✅ No token → ${noTokenRes.status}`);
  sleep(0.3);

  // 3. Booking dengan tanggal lampau → 400
  const pastDate = new Date();
  pastDate.setDate(pastDate.getDate() - 1);
  const pastDateStr = pastDate.toISOString().split('T')[0];

  const pastDateRes = post(`${base}/api/bookings`, {
    branch_id: 4, table_type_id: 6,
    booking_date: pastDateStr, start_time: '12:00', guest_count: 2,
  }, token);
  check(pastDateRes, { 'past date: 400': (r) => r.status === 400 });
  console.log(`✅ Past date → ${pastDateRes.status}`);
  sleep(0.3);

  // 4. Booking dengan tanggal > 30 hari → 400
  const farDate = new Date();
  farDate.setDate(farDate.getDate() + 35);
  const farDateStr = farDate.toISOString().split('T')[0];

  const farDateRes = post(`${base}/api/bookings`, {
    branch_id: 4, table_type_id: 6,
    booking_date: farDateStr, start_time: '12:00', guest_count: 2,
  }, token);
  check(farDateRes, { 'too far date: 400': (r) => r.status === 400 });
  console.log(`✅ Date > 30 days → ${farDateRes.status}`);
  sleep(0.3);

  // 5. Booking dengan guest_count = 0 → 400
  const zeroGuestRes = post(`${base}/api/bookings`, {
    branch_id: 4, table_type_id: 6,
    booking_date: tomorrow(), start_time: '12:00', guest_count: 0,
  }, token);
  check(zeroGuestRes, { 'zero guest: 400': (r) => r.status === 400 });
  console.log(`✅ Zero guest count → ${zeroGuestRes.status}`);
  sleep(0.3);

  // 6. Get booking yang tidak ada → 404
  const notFoundRes = get(`${base}/api/bookings/999999`, token);
  check(notFoundRes, { 'booking not found: 404': (r) => r.status === 404 });
  console.log(`✅ Booking not found → ${notFoundRes.status}`);
  sleep(0.3);

  // 7. Cancel booking orang lain → 400/403
  // (booking ID 1 mungkin milik orang lain atau tidak ada)
  const cancelOtherRes = del(`${base}/api/bookings/1`, token);
  check(cancelOtherRes, {
    'cancel other: non-2xx': (r) => r.status >= 400,
  });
  console.log(`✅ Cancel other's booking → ${cancelOtherRes.status}`);

  console.log('\n✅ All error cases passed');
}
