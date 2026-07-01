/**
 * One-time setup: Register customer2 untuk E2E tests
 *
 * Jalankan sekali saja:
 *   k6 run tests/k6/setup_customer2.js
 *
 * Setelah selesai:
 *   Cek email syarif.hidayatull4h@gmail.com → klik link verifikasi
 *   Lalu jalankan e2e_waiting_list.js (customer2 sudah siap dipakai)
 */
import { check, sleep } from 'k6';
import { BASE } from './config.js';
import { post } from './helpers.js';

export const options = { vus: 1, iterations: 1 };

const CUSTOMER2_EMAIL    = __ENV.CUSTOMER_2_EMAIL    || 'syarif.hidayatull4h+test2@gmail.com';
const CUSTOMER2_PASSWORD = __ENV.CUSTOMER_2_PASSWORD || 'password1';

export default function () {
  const res = post(`${BASE.customer}/api/customers/register`, {
    name:     'K6 Test Customer 2',
    email:    CUSTOMER2_EMAIL,
    password: CUSTOMER2_PASSWORD,
    phone:    '081200000002',
  });

  const ok = check(res, {
    'register customer2: 201': (r) => r.status === 201,
  });

  if (ok) {
    console.log(`✅ Customer2 terdaftar — email: ${CUSTOMER2_EMAIL}`);
    console.log('📧 Cek inbox Gmail → klik link verifikasi → customer2 siap dipakai');
  } else if (res.status === 400 && res.body.includes('sudah terdaftar')) {
    console.log(`ℹ️  Customer2 sudah terdaftar sebelumnya — email: ${CUSTOMER2_EMAIL}`);
    console.log('   Kalau belum bisa login, cek email untuk link verifikasi');
  } else {
    console.error(`❌ Register gagal: ${res.body}`);
  }
}
