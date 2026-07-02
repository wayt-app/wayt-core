/**
 * E2E: Notification Flow
 *
 * Flow:
 * 1. Mark semua notifikasi existing sebagai read (clean slate)
 * 2. Verifikasi unread kosong setelah mark
 * 3. Buat booking baru → trigger notifikasi "Menunggu Konfirmasi"
 * 4. GET /notifications?unread_only=true → ada notifikasi baru
 * 5. Verifikasi konten notifikasi (title, is_read=false)
 * 6. Mark all read lagi → cek unread kosong
 * 7. Cleanup (cancel booking)
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
  // ── Login ────────────────────────────────────────────────────────────────────
  const token = loginCustomer(BASE.customer, CREDS.customer.email, CREDS.customer.password);
  console.log('✅ Customer login OK');
  sleep(0.5);

  const ownerToken = loginOwner(BASE.owner, CREDS.owner.email, CREDS.owner.password);
  console.log('✅ Owner login OK');
  sleep(0.5);

  // ── Clean slate: mark semua notif existing sebagai read ───────────────────────
  const markRes = put(`${BASE.customer}/api/notifications/read`, {}, token);
  check(markRes, { 'mark all read: 200': (r) => r.status === 200 });
  console.log('✅ Semua notifikasi ditandai read');
  sleep(0.5);

  // ── Verifikasi unread kosong ──────────────────────────────────────────────────
  const emptyRes = get(`${BASE.customer}/api/notifications?unread_only=true`, token);
  check(emptyRes, {
    'unread kosong setelah mark: 200': (r) => r.status === 200,
    'unread kosong setelah mark: []':  (r) => {
      const d = r.json('data');
      return Array.isArray(d) && d.length === 0;
    },
  });
  console.log(`✅ Unread kosong — ${emptyRes.json('data') ? emptyRes.json('data').length : '?'} notifikasi`);
  sleep(0.5);

  // ── Buat booking → trigger notifikasi ────────────────────────────────────────
  const bookingRes = post(`${BASE.customer}/api/bookings`, {
    branch_id:     BRANCH_ID,
    table_type_id: TABLE_TYPE_ID,
    booking_date:  dateOffset(3),
    start_time:    '10:00',
    guest_count:   2,
    notes:         'k6 e2e notification test',
  }, token);

  const bookingCreated = check(bookingRes, {
    'booking: 201': (r) => r.status === 201,
  });
  if (!bookingCreated) fail(`Booking gagal: ${bookingRes.body}`);
  const bookingID = bookingRes.json('data.id');
  console.log(`✅ Booking dibuat — id=${bookingID} (trigger notifikasi)`);
  sleep(1.5); // Tunggu goroutine notifikasi selesai

  // ── GET unread notifications → harus ada notif baru ──────────────────────────
  const unreadRes = get(`${BASE.customer}/api/notifications?unread_only=true`, token);
  const notifs = unreadRes.json('data');
  const unreadCount = Array.isArray(notifs) ? notifs.length : 0;

  check(unreadRes, {
    'unread: 200':           (r) => r.status === 200,
    'unread: ada notifikasi': (r) => Array.isArray(r.json('data')) && r.json('data').length > 0,
  });
  console.log(`✅ Unread count: ${unreadCount}`);

  // Verifikasi konten notifikasi pertama
  if (unreadCount > 0) {
    const latest = notifs[0];
    check({ latest }, {
      'notif: is_read=false':  (v) => v.latest.is_read === false,
      'notif: ada title':      (v) => (v.latest.title || '').length > 0,
      'notif: ada message':    (v) => (v.latest.message || '').length > 0,
    });
    console.log(`✅ Notifikasi terbaru: "${latest.title}" — ${latest.message.slice(0, 60)}...`);
  }
  sleep(0.5);

  // ── Mark all read → unread harus kosong ──────────────────────────────────────
  const markRes2 = put(`${BASE.customer}/api/notifications/read`, {}, token);
  check(markRes2, { 'mark all read 2: 200': (r) => r.status === 200 });
  sleep(0.5);

  const afterMarkRes = get(`${BASE.customer}/api/notifications?unread_only=true`, token);
  check(afterMarkRes, {
    'setelah mark: unread kosong': (r) => {
      const d = r.json('data');
      return Array.isArray(d) && d.length === 0;
    },
  });
  console.log('✅ Setelah mark all read — unread kosong');
  sleep(0.5);

  // ── GET all notifications (non-unread-only) ───────────────────────────────────
  const allRes = get(`${BASE.customer}/api/notifications`, token);
  check(allRes, {
    'all notifs: 200':           (r) => r.status === 200,
    'all notifs: ada data':      (r) => Array.isArray(r.json('data')) && r.json('data').length > 0,
    'all notifs: semua is_read': (r) => {
      const d = r.json('data');
      return Array.isArray(d) && d.every(n => n.is_read === true);
    },
  });
  console.log(`✅ All notifications — ${allRes.json('data') ? allRes.json('data').length : 0} total, semua is_read=true`);
  sleep(0.5);

  // ── Bug #2: owner harus dapat notif in-app saat customer membatalkan ─────────
  // Clean slate notif owner (buang "Booking Baru" dari create di atas)
  const ownerMark = put(`${BASE.owner}/api/owner/notifications/read`, {}, ownerToken);
  check(ownerMark, { 'owner mark read: 200': (r) => r.status === 200 });
  sleep(0.5);

  // Customer membatalkan booking
  const cancelRes = del(`${BASE.customer}/api/bookings/${bookingID}`, token);
  check(cancelRes, { 'cancel booking: 200': (r) => r.status === 200 });
  console.log(`🧹 Booking #${bookingID} dibatalkan oleh customer`);
  sleep(1.5); // tunggu goroutine sendInAppNotif selesai

  // Owner harus punya notifikasi "Booking Dibatalkan" untuk booking ini
  const ownerNotifRes = get(`${BASE.owner}/api/owner/notifications?unread_only=true`, ownerToken);
  const ownerNotifs = ownerNotifRes.json('data');
  const cancelNotif = Array.isArray(ownerNotifs)
    ? ownerNotifs.find(n => (n.title || '').toLowerCase().includes('dibatalkan') && (n.message || '').includes('#' + bookingID))
    : null;
  check({ ownerNotifRes, cancelNotif }, {
    'owner notif cancel: 200':       (v) => v.ownerNotifRes.status === 200,
    'owner notif cancel: ada notif': (v) => v.cancelNotif != null,
  });
  console.log(cancelNotif
    ? `✅ Owner dapat notif cancel: "${cancelNotif.title}" — ${cancelNotif.message.slice(0, 60)}`
    : `❌ Owner TIDAK dapat notif cancel untuk #${bookingID}`);
}
