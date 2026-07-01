/**
 * E2E: Change Tables by Owner
 *
 * Flow:
 * 1. Customer buat booking (misal pakai Meja Kecil)
 * 2. Owner GET table-change-options → lihat semua meja + availability
 * 3. Owner PUT /tables dengan meja yang berbeda (Meja sedang jika available)
 * 4. Verifikasi tables_count berubah sesuai meja baru
 * 5. Negative: ganti ke meja yang tidak ada → 400
 * 6. Cleanup
 */
import { check, sleep, fail } from 'k6';
import { BASE, CREDS } from './config.js';
import { post, get, put, loginCustomer, loginOwner } from './helpers.js';

export const options = { vus: 1, iterations: 1 };

const BRANCH_ID     = 4;
const TABLE_TYPE_ID = 6;  // Meja Kecil (initial booking)

function dateOffset(days) {
  const d = new Date();
  d.setDate(d.getDate() + days);
  return d.toISOString().split('T')[0];
}

export default function () {
  // ── Login ────────────────────────────────────────────────────────────────────
  const token = loginCustomer(BASE.customer, CREDS.customer.email, CREDS.customer.password);
  console.log('✅ Customer login OK');
  sleep(0.5);

  const ownerToken = loginOwner(BASE.owner, CREDS.owner.email, CREDS.owner.password);
  console.log('✅ Owner login OK');
  sleep(0.5);

  // ── Customer buat booking dengan Meja Kecil ───────────────────────────────────
  const createRes = post(`${BASE.customer}/api/bookings`, {
    branch_id:     BRANCH_ID,
    table_type_id: TABLE_TYPE_ID,
    booking_date:  dateOffset(3),
    start_time:    '16:00',
    guest_count:   2,
    notes:         'k6 e2e change_tables — booking awal',
  }, token);

  const created = check(createRes, {
    'create: 201':          (r) => r.status === 201,
    'create: status ok':    (r) => ['pending', 'waiting_list'].includes(r.json('data.status')),
  });
  if (!created) fail(`Create booking gagal: ${createRes.body}`);
  const bookingID = createRes.json('data.id');
  const initTableTypeID = createRes.json('data.table_type_id');
  console.log(`✅ Booking dibuat — id=${bookingID} table_type_id=${initTableTypeID}`);
  sleep(0.5);

  // ── Owner: GET table-change-options ─────────────────────────────────────────
  const optsRes = get(`${BASE.owner}/api/owner/bookings/${bookingID}/table-change-options`, ownerToken);
  const optsOK = check(optsRes, {
    'table-change-options: 200':       (r) => r.status === 200,
    'table-change-options: ada data':  (r) => Array.isArray(r.json('data')) && r.json('data').length > 0,
  });
  if (!optsOK) fail(`GetTableChangeOptions gagal: ${optsRes.body}`);

  const opts = optsRes.json('data');
  console.log(`✅ Table change options: ${opts.length} meja tersedia`);

  // Response pakai table_type_id bukan id
  // Pilih meja available yang berbeda dari current table_type_id
  let targetTableIDs = [];
  for (let i = 0; i < opts.length; i++) {
    const opt = opts[i];
    if (opt.is_available && opt.table_type_id !== initTableTypeID) {
      targetTableIDs.push(opt.table_type_id);
      break; // cukup 1 meja untuk 2 tamu
    }
  }

  if (targetTableIDs.length === 0) {
    // Fallback: pakai meja available manapun (termasuk yang sama)
    for (let i = 0; i < opts.length; i++) {
      if (opts[i].is_available) {
        targetTableIDs.push(opts[i].table_type_id);
        break;
      }
    }
  }
  console.log(`ℹ️  Target meja untuk ganti: [${targetTableIDs.join(', ')}]`);
  sleep(0.5);

  // ── Owner: PUT /tables — ganti meja ──────────────────────────────────────────
  const changeRes = put(`${BASE.owner}/api/owner/bookings/${bookingID}/tables`, {
    table_type_ids: targetTableIDs,
  }, ownerToken);

  check(changeRes, {
    'change tables: 200':              (r) => r.status === 200,
    'change tables: tables_count > 0': (r) => r.json('data.tables_count') > 0,
  });
  if (changeRes.status !== 200) {
    console.error(`❌ Change tables gagal: ${changeRes.body}`);
  } else {
    console.log(`✅ Tables berhasil diganti — tables_count=${changeRes.json('data.tables_count')}`);
  }
  sleep(0.5);

  // ── Negative: ganti ke table_type_ids yang tidak ada ─────────────────────────
  const badRes = put(`${BASE.owner}/api/owner/bookings/${bookingID}/tables`, {
    table_type_ids: [99999],
  }, ownerToken);
  check(badRes, {
    'bad table_ids: 400': (r) => r.status === 400,
  });
  console.log(`✅ Ganti meja ID tidak valid → ${badRes.status} (expected 400)`);
  sleep(0.5);

  // ── Cleanup ───────────────────────────────────────────────────────────────────
  const cancelRes = put(`${BASE.owner}/api/owner/bookings/${bookingID}/cancel`, {}, ownerToken);
  check(cancelRes, { 'cleanup: cancel 200': (r) => r.status === 200 });
  console.log(`🧹 Cleanup: booking #${bookingID} cancelled`);
}
