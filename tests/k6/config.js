// Target URLs (public staging endpoints — safe as defaults).
// Override per-env with __ENV, e.g. `k6 run -e CUSTOMER_URL=https://... e2e_review.js`.
export const BASE = {
  customer: __ENV.CUSTOMER_URL || 'https://stg.app.wayt.fun',
  owner:    __ENV.OWNER_URL    || 'https://stg.owner.wayt.fun',
  admin:    __ENV.ADMIN_URL    || 'https://stg.admin.wayt.fun',
};

// Credentials MUST be injected via environment variables — never hardcode secrets here.
// Example: k6 run -e CUSTOMER_EMAIL=... -e CUSTOMER_PASSWORD=... e2e_review.js
// (or export them / source a local, untracked env file before running).
export const CREDS = {
  customer: {
    email:    __ENV.CUSTOMER_EMAIL    || '',
    password: __ENV.CUSTOMER_PASSWORD || '',
  },
  // Second customer used in E2E waiting list test — must exist in staging
  customer2: {
    email:    __ENV.CUSTOMER_2_EMAIL    || '',
    password: __ENV.CUSTOMER_2_PASSWORD || '',
  },
  owner: {
    email:    __ENV.OWNER_EMAIL    || '',
    password: __ENV.OWNER_PASSWORD || '',
  },
  // Test staff — harus di-insert manual ke DB (lihat setup note di e2e_staff_flow.js)
  staff: {
    email:    __ENV.STAFF_EMAIL    || '',
    password: __ENV.STAFF_PASSWORD || '',
  },
  admin: {
    username: __ENV.ADMIN_USERNAME || '',
    password: __ENV.ADMIN_PASSWORD || '',
  },
};
