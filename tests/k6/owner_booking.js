/**
 * Flow: Owner manages bookings
 * 1. Login as owner
 * 2. Get my restaurant info
 * 3. List bookings for branch
 * 4. Create a booking (via customer API) then confirm → check-in → complete it
 */
import { check, sleep } from 'k6';
import { BASE, CREDS } from './config.js';
import { post, get, put, del, loginCustomer, loginOwner, tomorrow } from './helpers.js';

export const options = { vus: 1, iterations: 1 };

export default function () {
  // --- Owner login ---
  const ownerToken = loginOwner(BASE.owner, CREDS.owner.email, CREDS.owner.password);
  console.log('✅ Owner login OK');
  sleep(0.5);

  // --- Get restaurant info ---
  const restaurantRes = get(`${BASE.owner}/api/owner/restaurant`, ownerToken);
  check(restaurantRes, {
    'owner: get restaurant 200': (r) => r.status === 200,
    'owner: has restaurant id':  (r) => r.json('data.id') > 0,
  });
  console.log(`✅ Restaurant: id=${restaurantRes.json('data.id')} name=${restaurantRes.json('data.name')}`);
  sleep(0.5);

  // --- List bookings for branch 4 ---
  const listRes = get(`${BASE.owner}/api/owner/bookings?branch_id=4`, ownerToken);
  check(listRes, {
    'owner: list bookings 200': (r) => r.status === 200,
  });
  console.log(`✅ List bookings OK`);
  sleep(0.5);

  // --- Create a test booking as customer, then manage it as owner ---
  const customerToken = loginCustomer(BASE.customer, CREDS.customer.email, CREDS.customer.password);
  sleep(0.5);

  const date = tomorrow();
  const createRes = post(`${BASE.customer}/api/bookings`, {
    branch_id:     4,
    table_type_id: 6,
    booking_date:  date,
    start_time:    '13:00',
    guest_count:   2,
    notes:         'k6 owner flow test',
  }, customerToken);

  const bookingCreated = check(createRes, {
    'setup: create booking 201': (r) => r.status === 201,
    'setup: has booking id':     (r) => r.json('data.id') > 0,
  });
  if (!bookingCreated) {
    console.error(`❌ Setup booking failed: ${createRes.body}`);
    return;
  }

  const bookingID = createRes.json('data.id');
  console.log(`✅ Test booking created — id=${bookingID} status=${createRes.json('data.status')}`);
  sleep(0.5);

  // --- Owner: Confirm booking ---
  const confirmRes = put(`${BASE.owner}/api/owner/bookings/${bookingID}/confirm`, {}, ownerToken);
  check(confirmRes, {
    'owner: confirm booking 200': (r) => r.status === 200,
  });
  if (confirmRes.status !== 200) {
    console.error(`❌ Confirm failed: ${confirmRes.body}`);
  } else {
    console.log(`✅ Booking #${bookingID} confirmed`);
  }
  sleep(0.5);

  // --- Owner: Check-in ---
  const checkinRes = put(`${BASE.owner}/api/owner/bookings/${bookingID}/checkin`, {}, ownerToken);
  check(checkinRes, {
    'owner: check-in 200': (r) => r.status === 200,
  });
  if (checkinRes.status !== 200) {
    console.error(`❌ Check-in failed: ${checkinRes.body}`);
  } else {
    console.log(`✅ Booking #${bookingID} checked in`);
  }
  sleep(0.5);

  // --- Owner: Complete ---
  const completeRes = put(`${BASE.owner}/api/owner/bookings/${bookingID}/complete`, {
    notes:      'k6 test completed',
    total_bill: 150000,
  }, ownerToken);
  check(completeRes, {
    'owner: complete 200': (r) => r.status === 200,
  });
  if (completeRes.status !== 200) {
    console.error(`❌ Complete failed: ${completeRes.body}`);
  } else {
    console.log(`✅ Booking #${bookingID} completed — full owner flow done`);
  }
}
