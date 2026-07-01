/**
 * E2E: Owner Onboarding — Setup Ruangan dan Meja
 *
 * Catatan: Registrasi owner butuh verifikasi email (tidak bisa di-automate).
 * Create branch dibatasi oleh subscription plan.
 * Test ini fokus ke setup SETELAH owner punya cabang: ruangan → meja → verifikasi.
 *
 * Flow:
 * 1. Owner login
 * 2. GET /api/owner/restaurant → verifikasi restaurant
 * 3. GET /api/owner/branches → ambil existing branch
 * 4. POST rooms → buat 2 ruangan baru di existing branch
 * 5. POST table-types → buat 3 meja per ruangan
 * 6. Verifikasi via public API (rooms + availability)
 * 7. Negative: buat meja dengan capacity=0 → 400
 * 8. Cleanup (delete table-types → rooms)
 */
import { check, sleep, fail } from 'k6';
import { BASE, CREDS } from './config.js';
import { post, get, put, del, loginOwner } from './helpers.js';

export const options = { vus: 1, iterations: 1 };

const BRANCH_ID = 4; // Existing staging branch

function tomorrow() {
  const d = new Date();
  d.setDate(d.getDate() + 1);
  return d.toISOString().split('T')[0];
}

export default function () {
  // ── Owner login ──────────────────────────────────────────────────────────────
  const ownerToken = loginOwner(BASE.owner, CREDS.owner.email, CREDS.owner.password);
  console.log('✅ Owner login OK');
  sleep(0.5);

  // ── GET restaurant ────────────────────────────────────────────────────────────
  const restaurantRes = get(`${BASE.owner}/api/owner/restaurant`, ownerToken);
  check(restaurantRes, {
    'restaurant: 200':      (r) => r.status === 200,
    'restaurant: ada nama': (r) => (r.json('data.name') || '').length > 0,
  });
  console.log(`✅ Restaurant: "${restaurantRes.json('data.name')}"`);
  sleep(0.5);

  // ── Buat 2 ruangan baru di existing branch ────────────────────────────────────
  const room1Res = post(`${BASE.owner}/api/owner/branches/${BRANCH_ID}/rooms`, {
    name:       'K6 Room Indoor',
    is_smoking: false,
  }, ownerToken);
  const room1Created = check(room1Res, {
    'create room 1: 201':    (r) => r.status === 201,
    'create room 1: ada id': (r) => r.json('data.id') > 0,
  });
  if (!room1Created) fail(`Create room 1 gagal: ${room1Res.body}`);
  const room1ID = room1Res.json('data.id');
  console.log(`✅ Ruangan 1: id=${room1ID} "${room1Res.json('data.name')}"`);
  sleep(0.3);

  const room2Res = post(`${BASE.owner}/api/owner/branches/${BRANCH_ID}/rooms`, {
    name:       'K6 Room Outdoor',
    is_smoking: true,
  }, ownerToken);
  check(room2Res, {
    'create room 2: 201':    (r) => r.status === 201,
    'create room 2: ada id': (r) => r.json('data.id') > 0,
  });
  const room2ID = room2Res.json('data.id');
  console.log(`✅ Ruangan 2: id=${room2ID} "${room2Res.json('data.name')}"`);
  sleep(0.5);

  // ── Buat meja per ruangan ─────────────────────────────────────────────────────
  const tableIDs = [];

  // 2 meja kecil di Room 1
  for (let i = 1; i <= 2; i++) {
    const r = post(`${BASE.owner}/api/owner/branches/${BRANCH_ID}/table-types`, {
      name:     `K6 Meja Kecil ${i}`,
      capacity: 2,
      room_id:  room1ID,
    }, ownerToken);
    check(r, { [`meja kecil ${i}: 201`]: (rr) => rr.status === 201 });
    if (r.status === 201) tableIDs.push(r.json('data.id'));
    sleep(0.2);
  }

  // 1 meja besar di Room 2
  const bigTableRes = post(`${BASE.owner}/api/owner/branches/${BRANCH_ID}/table-types`, {
    name:     'K6 Meja Besar 1',
    capacity: 6,
    room_id:  room2ID,
  }, ownerToken);
  check(bigTableRes, { 'meja besar 1: 201': (r) => r.status === 201 });
  if (bigTableRes.status === 201) tableIDs.push(bigTableRes.json('data.id'));
  console.log(`✅ ${tableIDs.length} meja dibuat`);
  sleep(0.5);

  // ── Negative: meja dengan capacity=0 → 400 ───────────────────────────────────
  const badTableRes = post(`${BASE.owner}/api/owner/branches/${BRANCH_ID}/table-types`, {
    name:     'Bad Meja',
    capacity: 0,
  }, ownerToken);
  check(badTableRes, {
    'meja capacity=0: 400': (r) => r.status === 400,
  });
  console.log(`✅ Meja capacity=0 → ${badTableRes.status} (expected 400)`);
  sleep(0.5);

  // ── Verifikasi via public API ─────────────────────────────────────────────────
  // GET rooms untuk branch
  const publicRoomsRes = get(`${BASE.customer}/api/branches/${BRANCH_ID}/rooms`);
  check(publicRoomsRes, {
    'public rooms: 200':              (r) => r.status === 200,
    'public rooms: ada K6 room':      (r) => {
      const rooms = r.json('data');
      return Array.isArray(rooms) && rooms.some(rm => rm.name === 'K6 Room Indoor');
    },
  });
  console.log(`✅ Public rooms OK — ${publicRoomsRes.json('data') ? publicRoomsRes.json('data').length : 0} total ruangan`);
  sleep(0.5);

  // GET availability untuk branch (verifikasi meja baru muncul)
  const availRes = get(
    `${BASE.customer}/api/branches/${BRANCH_ID}/availability?date=${tomorrow()}&start_time=10:00&guests=2&room_id=${room1ID}`
  );
  check(availRes, {
    'availability room baru: 200':      (r) => r.status === 200,
    'availability room baru: ada slot': (r) => Array.isArray(r.json('data')) && r.json('data').length > 0,
  });
  console.log(`✅ Availability room baru OK — ${availRes.json('data') ? availRes.json('data').length : 0} tipe meja`);
  sleep(0.5);

  // ── Cleanup: null-kan room_id (soft-delete FK issue) → hapus ruangan ─────────
  // Table-type delete adalah soft-delete (deleted_at) — FK ke room masih ada.
  // Solusi: null-kan room_id via update sebelum hapus room.
  for (const id of tableIDs) {
    put(`${BASE.owner}/api/owner/table-types/${id}`, { room_id: null }, ownerToken);
    sleep(0.2);
  }
  console.log(`🧹 room_id di-null-kan untuk ${tableIDs.length} meja`);
  sleep(0.5);

  const del1 = del(`${BASE.owner}/api/owner/rooms/${room1ID}`, ownerToken);
  check(del1, { 'delete room 1: 200': (r) => r.status === 200 });
  sleep(0.3);

  const del2 = del(`${BASE.owner}/api/owner/rooms/${room2ID}`, ownerToken);
  check(del2, { 'delete room 2: 200': (r) => r.status === 200 });
  console.log(`🧹 2 ruangan dihapus`);
}
