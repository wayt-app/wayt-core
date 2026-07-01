/**
 * E2E: Review Flow
 *
 * Flow:
 * 1. Customer buat booking → owner confirm → check-in → complete
 * 2. Customer submit review (rating + comment)
 * 3. Customer GET /bookings/:id/review → verifikasi review data
 * 4. Owner GET /reviews → review muncul
 * 5. Owner GET /reviews/stats → ada rating data
 * 6. Negative: submit review kedua kali → 400 (one review per booking)
 * 7. Negative: submit review untuk booking yang belum complete → 400
 */
import { check, sleep, fail } from 'k6';
import { BASE, CREDS } from './config.js';
import { post, get, put, loginCustomer, loginOwner } from './helpers.js';

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
  const token = loginCustomer(BASE.customer, CREDS.customer.email, CREDS.customer.password);
  console.log('✅ Customer login OK');
  sleep(0.5);

  const ownerToken = loginOwner(BASE.owner, CREDS.owner.email, CREDS.owner.password);
  console.log('✅ Owner login OK');
  sleep(0.5);

  // ── Buat booking untuk di-review ──────────────────────────────────────────────
  const bookingRes = post(`${BASE.customer}/api/bookings`, {
    branch_id:     BRANCH_ID,
    table_type_id: TABLE_TYPE_ID,
    booking_date:  dateOffset(3),
    start_time:    '11:00',
    guest_count:   2,
    notes:         'k6 e2e review flow',
  }, token);

  const bookingCreated = check(bookingRes, {
    'booking: 201': (r) => r.status === 201,
  });
  if (!bookingCreated) fail(`Booking gagal: ${bookingRes.body}`);
  const bookingID = bookingRes.json('data.id');
  console.log(`✅ Booking dibuat — id=${bookingID}`);
  sleep(0.5);

  // ── Negative: submit review sebelum booking complete → 400 ───────────────────
  const earlyReviewRes = post(`${BASE.customer}/api/bookings/${bookingID}/review`, {
    rating:  5,
    comment: 'k6 early review — should fail',
  }, token);
  check(earlyReviewRes, {
    'review sebelum complete: 400': (r) => r.status === 400,
  });
  console.log(`✅ Review sebelum complete → ${earlyReviewRes.status} (expected 400)`);
  sleep(0.5);

  // ── Owner: confirm → check-in → complete ─────────────────────────────────────
  const confirmRes = put(`${BASE.owner}/api/owner/bookings/${bookingID}/confirm`, {}, ownerToken);
  check(confirmRes, { 'confirm: 200': (r) => r.status === 200 });
  sleep(0.3);

  const checkinRes = put(`${BASE.owner}/api/owner/bookings/${bookingID}/checkin`, {}, ownerToken);
  check(checkinRes, { 'checkin: 200': (r) => r.status === 200 });
  sleep(0.3);

  const completeRes = put(`${BASE.owner}/api/owner/bookings/${bookingID}/complete`, {
    notes:      'k6 test selesai',
    total_bill: 120000,
  }, ownerToken);
  check(completeRes, { 'complete: 200': (r) => r.status === 200 });
  console.log(`✅ Booking #${bookingID} completed`);
  sleep(0.5);

  // ── Customer submit review ────────────────────────────────────────────────────
  const reviewRes = post(`${BASE.customer}/api/bookings/${bookingID}/review`, {
    rating:  4,
    comment: 'Makanan enak, pelayanan bagus! — k6 e2e test',
  }, token);

  check(reviewRes, {
    'submit review: 201':        (r) => r.status === 201,
    'submit review: rating=4':   (r) => r.json('data.rating') === 4,
    'submit review: ada comment': (r) => (r.json('data.comment') || '').length > 0,
    'submit review: booking_id cocok': (r) => r.json('data.booking_id') === bookingID,
  });
  if (reviewRes.status !== 201) {
    console.error(`❌ Submit review gagal: ${reviewRes.body}`);
  } else {
    console.log(`✅ Review submitted — id=${reviewRes.json('data.id')} rating=${reviewRes.json('data.rating')}`);
  }
  sleep(0.5);

  // ── Customer GET /bookings/:id/review ─────────────────────────────────────────
  const getReviewRes = get(`${BASE.customer}/api/bookings/${bookingID}/review`, token);
  check(getReviewRes, {
    'get review: 200':        (r) => r.status === 200,
    'get review: rating=4':   (r) => r.json('data.rating') === 4,
  });
  console.log(`✅ GET review OK — rating=${getReviewRes.json('data.rating')}`);
  sleep(0.5);

  // ── Negative: submit review dua kali → 400 ───────────────────────────────────
  const dupRes = post(`${BASE.customer}/api/bookings/${bookingID}/review`, {
    rating:  3,
    comment: 'duplicate review — should fail',
  }, token);
  check(dupRes, {
    'duplicate review: 400': (r) => r.status === 400,
  });
  console.log(`✅ Duplicate review → ${dupRes.status} (expected 400)`);
  sleep(0.5);

  // ── Owner GET /reviews ────────────────────────────────────────────────────────
  const ownerListRes = get(`${BASE.owner}/api/owner/reviews`, ownerToken);
  check(ownerListRes, {
    'owner reviews: 200':        (r) => r.status === 200,
    'owner reviews: ada review': (r) => r.json('data.total') > 0,
  });
  console.log('✅ Owner reviews list OK');
  sleep(0.5);

  // ── Owner GET /reviews/stats ──────────────────────────────────────────────────
  const statsRes = get(`${BASE.owner}/api/owner/reviews/stats`, ownerToken);
  check(statsRes, {
    'review stats: 200':           (r) => r.status === 200,
    'review stats: total_reviews > 0': (r) => r.json('data.total_reviews') > 0,
    'review stats: avg_rating > 0':    (r) => r.json('data.avg_rating') > 0,
  });
  console.log(`✅ Review stats OK — total=${statsRes.json('data.total_reviews')} avg=${statsRes.json('data.avg_rating')}`);
}
