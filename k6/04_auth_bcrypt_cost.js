import http from 'k6/http';
import { check, sleep } from 'k6';
import { Trend, Rate } from 'k6/metrics';

const authDuration = new Trend('auth_bcrypt_duration');
const authTtfb = new Trend('auth_bcrypt_ttfb');
const authSuccessRate = new Rate('auth_success_rate');

export const options = {
  scenarios: {
    auth_registration: {
      executor: 'constant-vus',
      vus: 2,
      duration: '15s',
    },
  },
};

const GRAPHQL_URL = 'http://localhost:8080/query';

export default function () {
  const randomIp = `192.168.${Math.floor(Math.random() * 250) + 1}.${Math.floor(Math.random() * 250) + 1}`;
  const randomNum = Math.floor(Math.random() * 8999) + 1000;
  const regId = `2024${randomNum}`;
  const email = `test${randomNum}mitblr${randomNum}@learner.manipal.edu`;
  const password = `SecretPass!${randomNum}`;

  const mutation = JSON.stringify({
    query: `mutation CreateUser($input: CreateUserInput!) {
      createUser(input: $input) {
        token
        user {
          id
          username
        }
      }
    }`,
    variables: {
      input: {
        email: email,
        regId: regId,
        password: password,
      },
    },
  });

  const params = {
    headers: {
      'Content-Type': 'application/json',
      'X-Forwarded-For': randomIp,
      'True-Client-Ip': randomIp,
    },
  };

  const res = http.post(GRAPHQL_URL, mutation, params);

  authTtfb.add(res.timings.waiting);
  authDuration.add(res.timings.duration);

  let success = false;
  try {
    const json = JSON.parse(res.body);
    if (json.data && json.data.createUser && json.data.createUser.token) {
      success = true;
    }
  } catch (e) {}

  const isOk = check(res, {
    'status is 200': (r) => r.status === 200,
    'user created & token generated': () => success,
  });

  authSuccessRate.add(isOk ? 1 : 0);
  sleep(0.5);
}
