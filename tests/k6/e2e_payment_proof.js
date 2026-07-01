/**
 * E2E: Upload Payment Proof
 *
 * Flow:
 * 1. Customer login & buat booking pending
 * 2. Upload bukti transfer (multipart PNG) → 200, proof_url tersimpan
 * 3. GET booking → verifikasi payment_proof_url terisi
 * 4. Negative: upload ekstensi tidak didukung (PDF) → 400
 * 5. Negative: upload tanpa auth → 401
 * 6. Cleanup (owner cancel)
 *
 * Requires: fixtures/test.png (buat dengan: python3 setup di README)
 */
import http from 'k6/http';
import { check, sleep, fail } from 'k6';
import { BASE, CREDS } from './config.js';
import { post, get, put, loginCustomer, loginOwner } from './helpers.js';

export const options = { vus: 1, iterations: 1 };

const BRANCH_ID     = 4;
const TABLE_TYPE_ID = 6;

// Baca file PNG dari disk (binary)
const PNG_FILE = open('./fixtures/test.png', 'b');

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

  // ── Buat booking ─────────────────────────────────────────────────────────────
  const createRes = post(`${BASE.customer}/api/bookings`, {
    branch_id:     BRANCH_ID,
    table_type_id: TABLE_TYPE_ID,
    booking_date:  dateOffset(3),
    start_time:    '17:00',
    guest_count:   2,
    notes:         'k6 e2e payment_proof',
  }, token);

  const created = check(createRes, {
    'create: 201': (r) => r.status === 201,
  });
  if (!created) {
    fail(`Create booking gagal: ${createRes.body}`);
  }
  const bookingID = createRes.json('data.id');
  console.log(`✅ Booking dibuat — id=${bookingID}`);
  sleep(0.5);

  // ── Upload bukti transfer (PNG valid) ─────────────────────────────────────────
  const uploadRes = http.post(
    `${BASE.customer}/api/bookings/${bookingID}/payment-proof`,
    { proof: http.file(PNG_FILE, 'bukti.png', 'image/png') },
    { headers: { 'Authorization': `Bearer ${token}` } }
  );

  check(uploadRes, {
    'upload: 200':                      (r) => r.status === 200,
    'upload: ada proof_url':            (r) => (r.json('data.proof_url') || '').length > 0,
    'upload: proof_url berisi supabase': (r) => (r.json('data.proof_url') || '').includes('supabase'),
  });
  if (uploadRes.status !== 200) {
    console.error(`❌ Upload gagal: ${uploadRes.body}`);
  } else {
    console.log(`✅ Upload OK — proof_url=${uploadRes.json('data.proof_url')}`);
  }
  sleep(0.5);

  // ── Verifikasi via GET booking ────────────────────────────────────────────────
  const getRes = get(`${BASE.customer}/api/bookings/${bookingID}`, token);
  check(getRes, {
    'get: 200':                   (r) => r.status === 200,
    'get: payment_proof_url ada': (r) => (r.json('data.payment_proof_url') || '').length > 0,
  });
  console.log(`✅ GET booking payment_proof_url=${getRes.json('data.payment_proof_url')}`);
  sleep(0.5);

  // ── Negative: ekstensi tidak diizinkan (PDF) ──────────────────────────────────
  const fakePdfBytes = new Uint8Array([0x25, 0x50, 0x44, 0x46]); // %PDF header
  const badExtRes = http.post(
    `${BASE.customer}/api/bookings/${bookingID}/payment-proof`,
    { proof: http.file(fakePdfBytes.buffer, 'document.pdf', 'application/pdf') },
    { headers: { 'Authorization': `Bearer ${token}` } }
  );
  check(badExtRes, {
    'upload bad ext (pdf): 400': (r) => r.status === 400,
  });
  console.log(`✅ Upload PDF → ${badExtRes.status} (expected 400)`);
  sleep(0.5);

  // ── Negative: tanpa Authorization header ──────────────────────────────────────
  const noAuthRes = http.post(
    `${BASE.customer}/api/bookings/${bookingID}/payment-proof`,
    { proof: http.file(PNG_FILE, 'bukti.png', 'image/png') }
  );
  check(noAuthRes, {
    'upload tanpa auth: 401': (r) => r.status === 401,
  });
  console.log(`✅ Upload tanpa auth → ${noAuthRes.status} (expected 401)`);
  sleep(0.5);

  // ── Cleanup ───────────────────────────────────────────────────────────────────
  const cancelRes = put(`${BASE.owner}/api/owner/bookings/${bookingID}/cancel`, {}, ownerToken);
  check(cancelRes, { 'cleanup: owner cancel 200': (r) => r.status === 200 });
  console.log(`🧹 Cleanup: booking #${bookingID} cancelled`);
}
