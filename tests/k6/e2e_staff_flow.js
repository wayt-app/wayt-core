/**
 * E2E: Staff Flow
 *
 * Flow:
 * 1. Owner login
 * 2. Owner buat staff baru (POST /api/owner/staff)
 * 3. Staff login (POST /api/staff/login)
 * 4. Staff GET /api/staff/me → verifikasi identitas
 * 5. Staff GET /api/staff/bookings → bisa lihat booking
 * 6. Customer buat booking → owner confirm → staff check-in → staff complete
 * 7. Negative: staff akses endpoint owner-only → 401/403
 * 8. Owner hapus staff (DELETE /api/owner/staff/:id) — cleanup
 */
import { check, sleep, fail } from 'k6';
import { BASE, CREDS } from './config.js';
import { post, get, put, del, loginCustomer, loginOwner } from './helpers.js';

export const options = { vus: 1, iterations: 1 };

const BRANCH_ID     = 4;
const TABLE_TYPE_ID = 6;
const STAFF_PASSWORD = 'StaffTest123!';
const STAFF_NAME     = 'K6 Test Staff';

function dateOffset(days) {
  const d = new Date();
  d.setDate(d.getDate() + days);
  return d.toISOString().split('T')[0];
}

export default function () {
  // Email unik — Date.now() di dalam default() berjalan di runtime k6
  const staffEmail = `k6.staff.${Date.now()}@wayt-test.invalid`;

  // ── Owner login ──────────────────────────────────────────────────────────────
  const ownerToken = loginOwner(BASE.owner, CREDS.owner.email, CREDS.owner.password);
  console.log('✅ Owner login OK');
  sleep(0.5);

  // ── Owner buat staff baru ─────────────────────────────────────────────────────
  const createStaffRes = post(`${BASE.owner}/api/owner/staff`, {
    branch_id: BRANCH_ID,
    name:      STAFF_NAME,
    email:     staffEmail,
    password:  STAFF_PASSWORD,
  }, ownerToken);

  const staffCreated = check(createStaffRes, {
    'create staff: 201':         (r) => r.status === 201,
    'create staff: ada id':      (r) => r.json('data.id') > 0,
    'create staff: email cocok': (r) => r.json('data.email') === staffEmail,
  });
  if (!staffCreated) fail(`Create staff gagal: ${createStaffRes.body}`);
  const staffID = createStaffRes.json('data.id');
  console.log(`✅ Staff dibuat — id=${staffID} email=${staffEmail}`);
  sleep(0.5);

  // ── Staff login ───────────────────────────────────────────────────────────────
  const staffLoginRes = post(`${BASE.owner}/api/staff/login`, {
    email:    staffEmail,
    password: STAFF_PASSWORD,
  });
  const staffLoggedIn = check(staffLoginRes, {
    'staff login: 200':       (r) => r.status === 200,
    'staff login: ada token': (r) => (r.json('data.token') || '').length > 0,
  });
  if (!staffLoggedIn) {
    del(`${BASE.owner}/api/owner/staff/${staffID}`, ownerToken);
    fail(`Staff login gagal: ${staffLoginRes.body}`);
  }
  const staffToken = staffLoginRes.json('data.token');
  console.log('✅ Staff login OK');
  sleep(0.5);

  // ── Staff GET /me ─────────────────────────────────────────────────────────────
  const meRes = get(`${BASE.owner}/api/staff/me`, staffToken);
  check(meRes, {
    'staff me: 200':          (r) => r.status === 200,
    'staff me: nama benar':   (r) => r.json('data.name') === STAFF_NAME,
    'staff me: branch cocok': (r) => r.json('data.branch_id') === BRANCH_ID,
  });
  console.log(`✅ Staff /me OK — name=${meRes.json('data.name')}`);
  sleep(0.5);

  // ── Staff GET bookings ────────────────────────────────────────────────────────
  const listRes = get(`${BASE.owner}/api/staff/bookings?branch_id=${BRANCH_ID}`, staffToken);
  check(listRes, {
    'staff list bookings: 200': (r) => r.status === 200,
  });
  console.log('✅ Staff list bookings OK');
  sleep(0.5);

  // ── Customer buat booking, owner confirm, staff check-in & complete ──────────
  const customerToken = loginCustomer(BASE.customer, CREDS.customer.email, CREDS.customer.password);
  sleep(0.5);

  const bookingRes = post(`${BASE.customer}/api/bookings`, {
    branch_id:     BRANCH_ID,
    table_type_id: TABLE_TYPE_ID,
    booking_date:  dateOffset(3),
    start_time:    '14:00',
    guest_count:   2,
    notes:         'k6 e2e staff_flow',
  }, customerToken);

  const bookingCreated = check(bookingRes, {
    'booking: 201': (r) => r.status === 201,
  });
  if (!bookingCreated) {
    del(`${BASE.owner}/api/owner/staff/${staffID}`, ownerToken);
    fail(`Booking gagal: ${bookingRes.body}`);
  }
  const bookingID = bookingRes.json('data.id');
  console.log(`✅ Booking dibuat — id=${bookingID}`);
  sleep(0.5);

  // Owner confirm
  const confirmRes = put(`${BASE.owner}/api/owner/bookings/${bookingID}/confirm`, {}, ownerToken);
  check(confirmRes, { 'owner confirm: 200': (r) => r.status === 200 });
  console.log(`✅ Owner confirmed booking #${bookingID}`);
  sleep(0.5);

  // Staff check-in
  const checkinRes = put(`${BASE.owner}/api/staff/bookings/${bookingID}/checkin`, {}, staffToken);
  check(checkinRes, { 'staff checkin: 200': (r) => r.status === 200 });
  if (checkinRes.status !== 200) {
    console.error(`❌ Staff check-in gagal: ${checkinRes.body}`);
  } else {
    console.log(`✅ Staff check-in OK — booking #${bookingID}`);
  }
  sleep(0.5);

  // Staff complete
  const completeRes = put(`${BASE.owner}/api/staff/bookings/${bookingID}/complete`, {
    notes:      'k6 staff complete',
    total_bill: 75000,
  }, staffToken);
  check(completeRes, { 'staff complete: 200': (r) => r.status === 200 });
  if (completeRes.status !== 200) {
    console.error(`❌ Staff complete gagal: ${completeRes.body}`);
  } else {
    console.log('✅ Staff complete OK');
  }
  sleep(0.5);

  // ── Negative: staff akses endpoint owner-only ──────────────────────────────────
  const badRes = post(`${BASE.owner}/api/owner/bookings/manual`, {
    branch_id:      BRANCH_ID,
    table_type_id:  TABLE_TYPE_ID,
    booking_date:   dateOffset(4),
    start_time:     '14:00',
    guest_count:    2,
    customer_name:  'Test',
    customer_phone: '08120000000',
  }, staffToken);
  check(badRes, {
    'staff akses owner-only: 401/403': (r) => r.status === 401 || r.status === 403,
  });
  console.log(`✅ Staff akses owner-only → ${badRes.status} (expected 401/403)`);
  sleep(0.5);

  // ── Owner hapus staff (cleanup) ───────────────────────────────────────────────
  const deleteRes = del(`${BASE.owner}/api/owner/staff/${staffID}`, ownerToken);
  check(deleteRes, { 'delete staff: 200': (r) => r.status === 200 });
  console.log(`🧹 Staff #${staffID} dihapus`);
}
