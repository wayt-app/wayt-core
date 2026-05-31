export const BASE = {
  customer: __ENV.CUSTOMER_URL || 'https://stg.app.wayt.fun',
  owner:    __ENV.OWNER_URL    || 'https://stg.owner.wayt.fun',
  admin:    __ENV.ADMIN_URL    || 'https://stg.admin.wayt.fun',
};

export const CREDS = {
  customer: {
    email:    __ENV.CUSTOMER_EMAIL    || 'syarif.hidayatull4h@gmail.com',
    password: __ENV.CUSTOMER_PASSWORD || 'password1',
  },
  owner: {
    email:    __ENV.OWNER_EMAIL    || 'syarif.hidayatull4h@gmail.com',
    password: __ENV.OWNER_PASSWORD || 'password1',
  },
  admin: {
    username: __ENV.ADMIN_USERNAME || 'superadmin',
    password: __ENV.ADMIN_PASSWORD || 'Password0!',
  },
};
