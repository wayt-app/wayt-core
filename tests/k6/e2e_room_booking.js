/**
 * E2E: Room-Based Booking (CreateFromRoom)
 *
 * Flow:
 * 1. GET /api/branches/4/rooms → verifikasi ada ruangan
 * 2. GET availability dengan room_id → lihat slot tersedia
 * 3. Buat booking dengan room_id (tanpa table_type_id) → sistem auto-assign meja optimal
 * 4. Verifikasi: room_id tersimpan, tables_count > 0 (ada assignment)
 * 5. Verifikasi optimal assignment: 4 tamu di room Indoor → 1 meja sedang (cap 4) bukan 2 meja kecil
 * 6. Negative: booking dengan room_id yang tidak ada → 400
 * 7. Cleanup
 *
 * Staging data:
 *   branch_id=4, room_id=4 (Indoor)
 *   Meja Kecil 1 (cap=2, 3 tables), Meja sedang 1 (cap=4, 2 tables)
 *   Optimal 4 guests → 1 meja sedang (tables_count=1, least waste)
 */
import { check, sleep, fail } from 'k6';
import { BASE, CREDS } from './config.js';
import { post, get, put, loginCustomer, loginOwner } from './helpers.js';

export const options = { vus: 1, iterations: 1 };

const BRANCH_ID  = 4;
const ROOM_ID    = 4;   // Indoor
const GUEST_COUNT = 4;  // Optimal: 1 meja sedang (cap 4), bukan 2 meja kecil
const SLOT_TIME  = '15:00';

function dateOffset(days) {
  const d = new Date();
  d.setDate(d.getDate() + days);
  return d.toISOString().split('T')[0];
}

export default function () {
  const date = dateOffset(3);

  // ── Login ────────────────────────────────────────────────────────────────────
  const token = loginCustomer(BASE.customer, CREDS.customer.email, CREDS.customer.password);
  console.log('✅ Customer login OK');
  sleep(0.5);

  const ownerToken = loginOwner(BASE.owner, CREDS.owner.email, CREDS.owner.password);
  console.log('✅ Owner login OK');
  sleep(0.5);

  // ── List rooms untuk branch 4 ─────────────────────────────────────────────────
  const roomsRes = get(`${BASE.customer}/api/branches/${BRANCH_ID}/rooms`);
  const roomsOK = check(roomsRes, {
    'rooms: 200':            (r) => r.status === 200,
    'rooms: ada ruangan':    (r) => Array.isArray(r.json('data')) && r.json('data').length > 0,
    'rooms: ada room Indoor': (r) => {
      const rooms = r.json('data');
      return Array.isArray(rooms) && rooms.some(rm => rm.id === ROOM_ID);
    },
  });
  if (!roomsOK) fail(`Rooms list gagal: ${roomsRes.body}`);
  console.log(`✅ Rooms OK — ${roomsRes.json('data').length} ruangan ditemukan`);
  sleep(0.5);

  // ── Cek availability dengan room_id ──────────────────────────────────────────
  const availRes = get(
    `${BASE.customer}/api/branches/${BRANCH_ID}/availability?date=${date}&start_time=${SLOT_TIME}&guests=${GUEST_COUNT}&room_id=${ROOM_ID}`
  );
  check(availRes, {
    'availability room: 200':           (r) => r.status === 200,
    'availability room: ada slot':      (r) => Array.isArray(r.json('data')) && r.json('data').length > 0,
    'availability room: ada meja sedang': (r) => {
      const slots = r.json('data');
      return Array.isArray(slots) && slots.some(s => s.capacity >= GUEST_COUNT && s.available > 0);
    },
  });
  console.log(`✅ Availability room OK — ${availRes.json('data') ? availRes.json('data').length : 0} tipe tersedia`);
  sleep(0.5);

  // ── Buat booking dengan room_id (tanpa table_type_id) ────────────────────────
  const createRes = post(`${BASE.customer}/api/bookings`, {
    branch_id:   BRANCH_ID,
    room_id:     ROOM_ID,
    booking_date: date,
    start_time:  SLOT_TIME,
    guest_count: GUEST_COUNT,
    notes:       'k6 e2e room_booking — auto assign',
  }, token);

  const created = check(createRes, {
    'create room booking: 201':          (r) => r.status === 201,
    'create room booking: room_id cocok': (r) => r.json('data.room_id') === ROOM_ID,
    'create room booking: tables_count > 0': (r) => r.json('data.tables_count') > 0,
  });
  if (!created) {
    fail(`Room booking gagal: ${createRes.body}`);
  }
  const bookingID   = createRes.json('data.id');
  const tablesCount = createRes.json('data.tables_count');
  console.log(`✅ Room booking dibuat — id=${bookingID} room_id=${createRes.json('data.room_id')} tables_count=${tablesCount}`);
  sleep(0.5);

  // ── Verifikasi optimal assignment: 4 tamu → 1 meja (sedang cap 4) ────────────
  // Sistem DFS memilih assignment dengan least wasted seats:
  // Option A: 1 meja sedang (cap 4) → waste = 0
  // Option B: 2 meja kecil (cap 2*2=4) → waste = 0 (sama, tapi lebih sedikit meja)
  // Optimal = 1 meja (fewer tables used)
  check(createRes, {
    'optimal: 1 meja untuk 4 tamu': (r) => r.json('data.tables_count') === 1,
  });
  if (tablesCount === 1) {
    console.log(`✅ Optimal assignment: ${tablesCount} meja untuk ${GUEST_COUNT} tamu`);
  } else {
    console.log(`ℹ️  tables_count=${tablesCount} (sistem memilih assignment berbeda, mungkin slot sedang tersedia habis)`);
  }
  sleep(0.5);

  // ── Verifikasi via GET booking ────────────────────────────────────────────────
  const getRes = get(`${BASE.customer}/api/bookings/${bookingID}`, token);
  check(getRes, {
    'get: 200':                         (r) => r.status === 200,
    'get: room_id tersimpan':           (r) => r.json('data.room_id') === ROOM_ID,
    'get: tables_count tersimpan':      (r) => r.json('data.tables_count') > 0,
    'get: guest_count benar':           (r) => r.json('data.guest_count') === GUEST_COUNT,
  });
  console.log(`✅ GET booking OK — room_id=${getRes.json('data.room_id')} tables_count=${getRes.json('data.tables_count')}`);
  sleep(0.5);

  // ── Negative: room_id tidak ada (9999) → 400 ─────────────────────────────────
  const badRoomRes = post(`${BASE.customer}/api/bookings`, {
    branch_id:    BRANCH_ID,
    room_id:      9999,
    booking_date: date,
    start_time:   SLOT_TIME,
    guest_count:  2,
    notes:        'k6 e2e room_booking — bad room_id',
  }, token);
  check(badRoomRes, {
    'bad room_id: 400': (r) => r.status === 400,
  });
  console.log(`✅ Booking room_id=9999 → ${badRoomRes.status} (expected 400)`);
  sleep(0.5);

  // ── Cleanup ───────────────────────────────────────────────────────────────────
  const cancelRes = put(`${BASE.owner}/api/owner/bookings/${bookingID}/cancel`, {}, ownerToken);
  check(cancelRes, { 'cleanup: owner cancel 200': (r) => r.status === 200 });
  console.log(`🧹 Cleanup: booking #${bookingID} cancelled`);
}
