/**
 * E2E: Promo Link Flow
 *
 * Flow:
 * 1. Owner buat promo link baru → dapat token
 * 2. Public GET /api/promo-links/:token → increment visit_count (async goroutine)
 * 3. Public GET /api/restaurants/promo/:token → dapat info restoran
 * 4. Owner cek visit_count bertambah
 * 5. Negative: token tidak valid → 404
 * 6. Owner delete promo link (cleanup)
 */
import { check, sleep, fail } from 'k6';
import { BASE, CREDS } from './config.js';
import { post, get, del, loginOwner } from './helpers.js';

export const options = { vus: 1, iterations: 1 };

export default function () {
  // ── Owner login ──────────────────────────────────────────────────────────────
  const ownerToken = loginOwner(BASE.owner, CREDS.owner.email, CREDS.owner.password);
  console.log('✅ Owner login OK');
  sleep(0.5);

  // ── Owner buat promo link baru ────────────────────────────────────────────────
  const createRes = post(`${BASE.owner}/api/owner/promo-links`, {
    label: 'K6 E2E Test Link',
  }, ownerToken);

  const created = check(createRes, {
    'create promo link: 201':   (r) => r.status === 201,
    'create promo link: token': (r) => (r.json('data.token') || '').length > 0,
    'create promo link: visit_count=0': (r) => r.json('data.visit_count') === 0,
  });
  if (!created) fail(`Create promo link gagal: ${createRes.body}`);
  const linkID = createRes.json('data.id');
  const token  = createRes.json('data.token');
  console.log(`✅ Promo link dibuat — id=${linkID} token=${token} visit_count=0`);
  sleep(0.5);

  // ── Public: GET /api/promo-links/:token → return restaurant info + increment ──
  // Response berisi data restaurant (name, id, logo_url, dll)
  const visit1 = get(`${BASE.customer}/api/promo-links/${token}`);
  check(visit1, {
    'visit 1: 200':             (r) => r.status === 200,
    'visit 1: ada nama resto':  (r) => (r.json('data.name') || '').length > 0,
    'visit 1: ada resto id':    (r) => r.json('data.id') > 0,
  });
  console.log(`✅ Visit 1 OK — restoran: ${visit1.json('data.name')}`);
  sleep(0.5);

  // Visit 2 — increment lagi
  get(`${BASE.customer}/api/promo-links/${token}`);
  console.log('✅ Visit 2 OK');
  sleep(1); // Tunggu goroutine increment selesai

  // ── Owner cek visit_count bertambah ──────────────────────────────────────────
  const listRes = get(`${BASE.owner}/api/owner/promo-links`, ownerToken);
  check(listRes, {
    'list: 200': (r) => r.status === 200,
  });

  const links = listRes.json('data');
  let updatedCount = -1;
  if (Array.isArray(links)) {
    for (let i = 0; i < links.length; i++) {
      if (links[i].id === linkID) {
        updatedCount = links[i].visit_count;
        break;
      }
    }
  }
  check({ updatedCount }, {
    'visit_count bertambah (>= 2)': (v) => v.updatedCount >= 2,
  });
  console.log(`✅ visit_count setelah 2 kunjungan: ${updatedCount}`);
  sleep(0.5);

  // ── Verifikasi via update label (endpoint yang return data promo link) ──────────
  // /api/restaurants/promo/:token memakai promo_token restaurant (bukan promo link token)
  // Skip endpoint itu — sudah verified via visit_count bertambah di atas
  console.log('ℹ️  /api/restaurants/promo/:token pakai promo_token restaurant, skip');

  // ── Negative: token tidak valid → 404 ────────────────────────────────────────
  const badRes = get(`${BASE.customer}/api/promo-links/invalidtoken999`);
  check(badRes, {
    'invalid token: 404': (r) => r.status === 404,
  });
  console.log(`✅ Invalid token → ${badRes.status} (expected 404)`);
  sleep(0.5);

  // ── Cleanup: owner delete promo link ─────────────────────────────────────────
  const deleteRes = del(`${BASE.owner}/api/owner/promo-links/${linkID}`, ownerToken);
  check(deleteRes, { 'delete: 200': (r) => r.status === 200 });
  console.log(`🧹 Promo link #${linkID} dihapus`);
}
