/**
 * Flow: Customer booking end-to-end
 * 1. Login
 * 2. Check availability
 * 3. Create booking
 * 4. Get booking by ID
 * 5. Get my bookings (list)
 * 6. Cancel booking (cleanup)
 */
import { check, sleep } from 'k6';
import { BASE, CREDS } from './config.js';
import { post, get, del, expectOK, loginCustomer, tomorrow } from './helpers.js';

export const options = { vus: 1, iterations: 1 };

export default function () {
  const base = BASE.customer;
  const { email, password } = CREDS.customer;

  // 1. Login
  const token = loginCustomer(base, email, password);
  console.log('✅ Customer login OK');
  sleep(0.5);

  // 2. Check availability
  const date = tomorrow();
  const availRes = get(`${base}/api/branches/4/availability?date=${date}&start_time=12:00&guests=2`);
  check(availRes, {
    'availability: status 200':    (r) => r.status === 200,
    'availability: has results':   (r) => Array.isArray(r.json('data')) && r.json('data').length > 0,
    'availability: table_type_id': (r) => r.json('data.0.table_type_id') > 0,
  });
  console.log(`✅ Availability check OK — available: ${availRes.json('data.0.available')}`);
  sleep(0.5);

  // 3. Create booking
  const createRes = post(`${base}/api/bookings`, {
    branch_id:     4,
    table_type_id: 6,
    booking_date:  date,
    start_time:    '12:00',
    guest_count:   2,
    notes:         'k6 automated test',
  }, token);

  const bookingCreated = check(createRes, {
    'create booking: status 201': (r) => r.status === 201,
    'create booking: has id':     (r) => r.json('data.id') > 0,
  });
  if (!bookingCreated) {
    console.error(`❌ Create booking failed: ${createRes.body}`);
    return;
  }

  const bookingID = createRes.json('data.id');
  const bookingStatus = createRes.json('data.status');
  console.log(`✅ Booking created — id=${bookingID} status=${bookingStatus}`);
  sleep(0.5);

  // 4. Get booking by ID
  const getRes = get(`${base}/api/bookings/${bookingID}`, token);
  check(getRes, {
    'get booking: status 200':        (r) => r.status === 200,
    'get booking: correct id':        (r) => r.json('data.id') === bookingID,
    'get booking: has branch_id':     (r) => r.json('data.branch_id') === 4,
    'get booking: has guest_count':   (r) => r.json('data.guest_count') === 2,
  });
  console.log('✅ Get booking by ID OK');
  sleep(0.5);

  // 5. Get my bookings (list)
  const listRes = get(`${base}/api/bookings`, token);
  check(listRes, {
    'my bookings: status 200':   (r) => r.status === 200,
    'my bookings: non-empty':    (r) => {
      const data = r.json('data');
      return data !== null && (Array.isArray(data) ? data.length > 0 : Object.keys(data).length > 0);
    },
  });
  console.log('✅ My bookings list OK');
  sleep(0.5);

  // 6. Cancel booking (cleanup)
  const cancelRes = del(`${base}/api/bookings/${bookingID}`, token);
  check(cancelRes, {
    'cancel booking: status 200': (r) => r.status === 200,
  });
  if (cancelRes.status !== 200) {
    console.error(`❌ Cancel booking failed: ${cancelRes.body}`);
  } else {
    console.log(`✅ Booking #${bookingID} cancelled (cleanup)`);
  }
}
