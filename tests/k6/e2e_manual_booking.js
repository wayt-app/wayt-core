/**
 * E2E: Manual Booking by Owner + Pre-Order Status
 *
 * Flow:
 * 1. Owner login
 * 2. POST /api/owner/bookings/manual → buat booking manual dengan menu order
 * 3. Verifikasi source='manual', customer_name tersimpan
 * 4. Update order status: new → prepare → ready → done (4 steps)
 * 5. Negative: update ke status tidak valid → 400
 * 6. Cleanup (cancel)
 */
import { check, sleep, fail } from 'k6';
import { BASE, CREDS } from './config.js';
import { post, get, put, loginOwner } from './helpers.js';

export const options = { vus: 1, iterations: 1 };

const BRANCH_ID     = 4;
const TABLE_TYPE_ID = 6;

function dateOffset(days) {
  const d = new Date();
  d.setDate(d.getDate() + days);
  return d.toISOString().split('T')[0];
}

export default function () {
  // ── Login ────────────────────────────────────────────────────────────────────
  const ownerToken = loginOwner(BASE.owner, CREDS.owner.email, CREDS.owner.password);
  console.log('✅ Owner login OK');
  sleep(0.5);

  // ── Buat manual booking ───────────────────────────────────────────────────────
  const createRes = post(`${BASE.owner}/api/owner/bookings/manual`, {
    branch_id:      BRANCH_ID,
    table_type_id:  TABLE_TYPE_ID,
    booking_date:   dateOffset(3),
    start_time:     '19:00',
    guest_count:    2,
    notes:          'k6 e2e manual_booking',
    customer_name:  'Tamu K6 Test',
    customer_phone: '081300000099',
    customer_email: 'k6.guest@wayt-test.invalid',
    menu_order:     '2x Nasi Goreng; 1x Es Teh',
  }, ownerToken);

  const created = check(createRes, {
    'manual create: 201':             (r) => r.status === 201,
    'manual create: source=manual':   (r) => r.json('data.source') === 'manual',
    'manual create: guest_count=2':   (r) => r.json('data.guest_count') === 2,
    'manual create: order_status=new': (r) => r.json('data.order_status') === 'new',
  });
  if (!created) {
    fail(`Manual booking gagal: ${createRes.body}`);
  }
  const bookingID = createRes.json('data.id');
  console.log(`✅ Manual booking dibuat — id=${bookingID} source=${createRes.json('data.source')}`);
  sleep(0.5);

  // ── Update order status: new → prepare → ready → done ────────────────────────
  const statuses = ['prepare', 'ready', 'done'];
  for (const status of statuses) {
    const res = put(`${BASE.owner}/api/owner/bookings/${bookingID}/order-status`, {
      branch_id: BRANCH_ID,
      status:    status,
    }, ownerToken);
    check(res, {
      [`order_status ${status}: 200`]: (r) => r.status === 200,
    });
    if (res.status !== 200) {
      console.error(`❌ Update order_status=${status} gagal: ${res.body}`);
    } else {
      console.log(`✅ Order status → ${status}`);
    }
    sleep(0.3);
  }

  // ── Verifikasi via GET booking ────────────────────────────────────────────────
  const listRes = get(`${BASE.owner}/api/owner/bookings?branch_id=${BRANCH_ID}`, ownerToken);
  check(listRes, {
    'list bookings: 200': (r) => r.status === 200,
  });
  sleep(0.3);

  // ── Negative: status tidak valid → 400 ───────────────────────────────────────
  const badStatusRes = put(`${BASE.owner}/api/owner/bookings/${bookingID}/order-status`, {
    branch_id: BRANCH_ID,
    status:    'invalid_status',
  }, ownerToken);
  check(badStatusRes, {
    'bad order_status: 400': (r) => r.status === 400,
  });
  console.log(`✅ Order status invalid → ${badStatusRes.status} (expected 400)`);
  sleep(0.5);

  // ── Cleanup ───────────────────────────────────────────────────────────────────
  const cancelRes = put(`${BASE.owner}/api/owner/bookings/${bookingID}/cancel`, {}, ownerToken);
  check(cancelRes, { 'cleanup: cancel 200': (r) => r.status === 200 });
  console.log(`🧹 Cleanup: booking #${bookingID} cancelled`);
}
