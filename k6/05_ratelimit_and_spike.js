import http from 'k6/http';
import { check, sleep } from 'k6';
import { Trend, Rate, Counter } from 'k6/metrics';

const rateLimit429Count = new Counter('ratelimit_429_count');
const rateLimit200Count = new Counter('ratelimit_200_count');
const rateLimitRejectionLatency = new Trend('ratelimit_429_latency');

export const options = {
  scenarios: {
    single_ip_burst: {
      executor: 'per-vu-iterations',
      vus: 1,
      iterations: 20,
      maxDuration: '10s',
    },
  },
};

const GRAPHQL_URL = 'http://localhost:8080/query';

export default function () {
  const staticIp = '203.0.113.42'; // Fixed IP to test rate limiting per IP

  const query = JSON.stringify({
    query: `query { getAllCompanies { id companyName } }`,
  });

  const params = {
    headers: {
      'Content-Type': 'application/json',
      'X-Forwarded-For': staticIp,
      'True-Client-Ip': staticIp,
    },
  };

  const res = http.post(GRAPHQL_URL, query, params);

  if (res.status === 429) {
    rateLimit429Count.add(1);
    rateLimitRejectionLatency.add(res.timings.duration);
    check(res, {
      'is 429 Too Many Requests': (r) => r.status === 429,
      'has RateLimit-RetryAfter header': (r) => r.headers['Ratelimit-Retryafter'] !== undefined || r.headers['RateLimit-RetryAfter'] !== undefined,
    });
  } else if (res.status === 200) {
    rateLimit200Count.add(1);
    check(res, {
      'is 200 OK': (r) => r.status === 200,
      'has RateLimit-Remaining header': (r) => r.headers['Ratelimit-Remaining'] !== undefined || r.headers['RateLimit-Remaining'] !== undefined,
    });
  }

  sleep(0.01); // rapid burst
}
