/**
 * E2E: Waiting List & Auto-Promote
 *
 * Flow:
 * 1. Login semua actor di awal (customer A, customer B, owner)
 * 2. Cek availability → isi semua sisa slot dengan booking dari customer A
 * 3. Customer B buat booking di slot yang sama → waiting_list
 * 4. Owner cancel salah satu booking customer A
 * 5. Assert booking B promoted dari waiting_list → pending
 *
 * Prerequisites:
 *   - Staging branch_id=4, table_type_id=6 (Meja Kecil, capacity=2)
 *   - Customer2 terdaftar & terverifikasi: jalankan setup_customer2.js dulu
 */
import { check, sleep, fail } from 'k6';
import { BASE, CREDS } from './config.js';
import { post, get, put, del, loginCustomer, loginOwner } from './helpers.js';

export const options = { vus: 1, iterations: 1 };

const BRANCH_ID     = 4;
const TABLE_TYPE_ID = 6;   // Meja Kecil, capacity=2, room_id=4
const SLOT_TIME     = '20:00';

function bookingDate() {
  const d = new Date();
  d.setDate(d.getDate() + 3);
  return d.toISOString().split('T')[0];
}

export default function () {
  const date = bookingDate();

  // ── Login semua actor di awal ────────────────────────────────────────────────
  const tokenA = loginCustomer(BASE.customer, CREDS.customer.email, CREDS.customer.password);
  console.log('✅ Customer A login OK');
  sleep(0.5);

  const tokenB = loginCustomer(BASE.customer, CREDS.customer2.email, CREDS.customer2.password);
  console.log('✅ Customer B login OK');
  sleep(0.5);

  const ownerToken = loginOwner(BASE.owner, CREDS.owner.email, CREDS.owner.password);
  console.log('✅ Owner login OK');
  sleep(0.5);

  // ── Cek berapa sisa slot tersedia ────────────────────────────────────────────
  const availRes = get(
    `${BASE.customer}/api/branches/${BRANCH_ID}/availability?date=${date}&start_time=${SLOT_TIME}&guests=2`
  );
  check(availRes, { 'availability: 200': (r) => r.status === 200 });

  const slots = availRes.json('data');
  let availableCount = 0;
  if (Array.isArray(slots)) {
    for (let i = 0; i < slots.length; i++) {
      if (slots[i].table_type_id === TABLE_TYPE_ID) {
        availableCount = slots[i].available;
        break;
      }
    }
  }
  console.log(`ℹ️  Sisa slot tersedia untuk table_type_id=${TABLE_TYPE_ID}: ${availableCount}`);

  if (availableCount === 0) {
    fail(`Tidak ada slot tersedia — slot sudah penuh sebelum test. Coba jam lain atau bersihkan data staging.`);
  }

  // ── Customer A: isi semua sisa slot ─────────────────────────────────────────
  const fillerIDs = [];
  for (let i = 0; i < availableCount; i++) {
    const res = post(`${BASE.customer}/api/bookings`, {
      branch_id:     BRANCH_ID,
      table_type_id: TABLE_TYPE_ID,
      booking_date:  date,
      start_time:    SLOT_TIME,
      guest_count:   2,
      notes:         `k6 e2e waiting_list — filler ${i + 1}/${availableCount}`,
    }, tokenA);
    check(res, {
      [`filler ${i + 1}: created 201`]:      (r) => r.status === 201,
      [`filler ${i + 1}: not waiting_list`]: (r) => r.json('data.status') !== 'waiting_list',
    });
    if (res.status !== 201) {
      console.error(`❌ Filler ${i + 1} gagal: ${res.body}`);
      break;
    }
    fillerIDs.push(res.json('data.id'));
    console.log(`✅ Filler ${i + 1}/${availableCount} — id=${res.json('data.id')} status=${res.json('data.status')}`);
    sleep(0.3);
  }

  if (fillerIDs.length < availableCount) {
    // Cleanup dan fail
    for (let id of fillerIDs) del(`${BASE.customer}/api/bookings/${id}`, tokenA);
    fail(`Tidak bisa mengisi semua slot (${fillerIDs.length}/${availableCount} berhasil)`);
  }

  // ── Customer B: buat booking → harus waiting_list ────────────────────────────
  const resB = post(`${BASE.customer}/api/bookings`, {
    branch_id:     BRANCH_ID,
    table_type_id: TABLE_TYPE_ID,
    booking_date:  date,
    start_time:    SLOT_TIME,
    guest_count:   2,
    notes:         'k6 e2e waiting_list — booking B (harus waiting_list)',
  }, tokenB);

  const bookingBOK = check(resB, {
    'booking B: created 201':     (r) => r.status === 201,
    'booking B: is waiting_list': (r) => r.json('data.status') === 'waiting_list',
  });
  if (!bookingBOK) {
    console.error(`❌ Booking B tidak waiting_list: ${resB.body}`);
    for (let id of fillerIDs) del(`${BASE.customer}/api/bookings/${id}`, tokenA);
    fail('Booking B tidak masuk waiting_list');
  }
  const bookingBID = resB.json('data.id');
  console.log(`✅ Booking B — id=${bookingBID} status=waiting_list ✓`);
  sleep(0.5);

  // ── Owner: cancel filler pertama ─────────────────────────────────────────────
  const cancelTarget = fillerIDs[0];
  const cancelRes = put(`${BASE.owner}/api/owner/bookings/${cancelTarget}/cancel`, {}, ownerToken);
  check(cancelRes, {
    'owner: cancel filler 200': (r) => r.status === 200,
  });
  if (cancelRes.status !== 200) {
    console.error(`❌ Owner cancel gagal: ${cancelRes.body}`);
  } else {
    console.log(`✅ Owner cancelled filler #${cancelTarget}`);
  }
  sleep(1.5); // Beri waktu autoPromote goroutine jalan

  // ── Assert: booking B terpromote → pending ───────────────────────────────────
  const checkBRes = get(`${BASE.customer}/api/bookings/${bookingBID}`, tokenB);
  const statusBFinal = checkBRes.json('data.status');
  check(checkBRes, {
    'booking B: promoted to pending': (r) => r.json('data.status') === 'pending',
  });
  console.log(`✅ Booking B final status: ${statusBFinal}`);

  if (statusBFinal === 'waiting_list') {
    console.error('❌ Booking B masih waiting_list — autoPromote tidak jalan?');
  }

  // ── Cleanup ───────────────────────────────────────────────────────────────────
  sleep(0.3);
  for (let i = 1; i < fillerIDs.length; i++) {
    const r = del(`${BASE.customer}/api/bookings/${fillerIDs[i]}`, tokenA);
    check(r, { [`cleanup: filler ${i + 1}`]: (rr) => rr.status === 200 });
  }
  const cleanupB = del(`${BASE.customer}/api/bookings/${bookingBID}`, tokenB);
  check(cleanupB, { 'cleanup: booking B': (r) => r.status === 200 });
  console.log(`🧹 Cleanup selesai`);
}
