/**
 * E2E: Reschedule Booking
 *
 * Flow:
 * 1. Customer login
 * 2. Buat booking untuk slot hari+3 jam 18:00 → pending
 * 3. Reschedule ke hari+4 jam 18:00
 * 4. Verifikasi tanggal & waktu berubah di response
 * 5. Coba reschedule booking yang sudah confirmed → harus error
 * 6. Cleanup (cancel)
 *
 * Constraint dari service: reschedule hanya untuk status "pending"
 */
import { check, sleep, fail } from 'k6';
import { BASE, CREDS } from './config.js';
import { post, get, put, del, loginCustomer, loginOwner } from './helpers.js';

export const options = { vus: 1, iterations: 1 };

const BRANCH_ID     = 4;
const TABLE_TYPE_ID = 6;

function dateOffset(days) {
  const d = new Date();
  d.setDate(d.getDate() + days);
  return d.toISOString().split('T')[0];
}

export default function () {
  const originalDate  = dateOffset(3);
  const reschedDate   = dateOffset(4);
  const START_TIME    = '18:00';

  // ── Login ────────────────────────────────────────────────────────────────────
  const token = loginCustomer(BASE.customer, CREDS.customer.email, CREDS.customer.password);
  console.log('✅ Customer login OK');
  sleep(0.5);

  const ownerToken = loginOwner(BASE.owner, CREDS.owner.email, CREDS.owner.password);
  console.log('✅ Owner login OK');
  sleep(0.5);

  // ── Buat booking original ────────────────────────────────────────────────────
  const createRes = post(`${BASE.customer}/api/bookings`, {
    branch_id:     BRANCH_ID,
    table_type_id: TABLE_TYPE_ID,
    booking_date:  originalDate,
    start_time:    START_TIME,
    guest_count:   2,
    notes:         'k6 e2e reschedule — booking awal',
  }, token);

  const created = check(createRes, {
    'create: 201':          (r) => r.status === 201,
    'create: status pending': (r) => r.json('data.status') === 'pending',
  });
  if (!created) {
    console.error(`❌ Create gagal: ${createRes.body}`);
    fail('Booking awal gagal dibuat');
  }
  const bookingID = createRes.json('data.id');
  console.log(`✅ Booking dibuat — id=${bookingID} date=${originalDate} time=${START_TIME}`);
  sleep(0.5);

  // ── Reschedule ke tanggal baru ───────────────────────────────────────────────
  const reschedRes = put(`${BASE.customer}/api/bookings/${bookingID}/reschedule`, {
    booking_date: reschedDate,
    start_time:   START_TIME,
  }, token);

  check(reschedRes, {
    'reschedule: 200':                    (r) => r.status === 200,
    'reschedule: date berubah':           (r) => (r.json('data.booking_date') || '').startsWith(reschedDate),
    'reschedule: start_time tetap sama':  (r) => (r.json('data.start_time') || '').startsWith(START_TIME),
    'reschedule: status masih pending':   (r) => r.json('data.status') === 'pending',
  });
  if (reschedRes.status !== 200) {
    console.error(`❌ Reschedule gagal: ${reschedRes.body}`);
  } else {
    console.log(`✅ Reschedule OK — date berubah: ${originalDate} → ${reschedDate}`);
  }
  sleep(0.5);

  // ── Verifikasi via GET booking ───────────────────────────────────────────────
  const getRes = get(`${BASE.customer}/api/bookings/${bookingID}`, token);
  check(getRes, {
    'get: 200':              (r) => r.status === 200,
    'get: date sudah baru':  (r) => (r.json('data.booking_date') || '').startsWith(reschedDate),
  });
  console.log(`✅ GET booking confirmed date=${getRes.json('data.booking_date')}`);
  sleep(0.5);

  // ── Negative test: reschedule ke tanggal lampau → harus error ───────────────
  const pastRes = put(`${BASE.customer}/api/bookings/${bookingID}/reschedule`, {
    booking_date: '2020-01-01',
    start_time:   START_TIME,
  }, token);
  check(pastRes, {
    'reschedule past date: 400': (r) => r.status === 400,
  });
  console.log(`✅ Reschedule ke masa lalu → 400 (expected)`);
  sleep(0.5);

  // ── Negative test: reschedule booking confirmed → harus error ───────────────
  // Confirm dulu via owner
  const confirmRes = put(`${BASE.owner}/api/owner/bookings/${bookingID}/confirm`, {}, ownerToken);
  check(confirmRes, { 'owner confirm: 200': (r) => r.status === 200 });
  console.log(`✅ Owner confirmed booking #${bookingID}`);
  sleep(0.5);

  const reschedConfirmedRes = put(`${BASE.customer}/api/bookings/${bookingID}/reschedule`, {
    booking_date: dateOffset(5),
    start_time:   START_TIME,
  }, token);
  check(reschedConfirmedRes, {
    'reschedule confirmed: 400': (r) => r.status === 400,
  });
  console.log(`✅ Reschedule booking confirmed → 400 (expected)`);
  sleep(0.5);

  // ── Cleanup ───────────────────────────────────────────────────────────────────
  const cancelRes = put(`${BASE.owner}/api/owner/bookings/${bookingID}/cancel`, {}, ownerToken);
  check(cancelRes, { 'cleanup: owner cancel 200': (r) => r.status === 200 });
  console.log(`🧹 Cleanup: booking #${bookingID} cancelled`);
}
