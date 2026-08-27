import http from 'k6/http';
import { check, sleep } from 'k6';
import { Trend, Rate, Counter } from 'k6/metrics';

const stressLatency = new Trend('stress_query_latency');
const stressTtfb = new Trend('stress_ttfb');
const stressSuccessRate = new Rate('stress_success_rate');
const stressReqCount = new Counter('stress_req_count');

export const options = {
  stages: [
    { duration: '5s', target: 10 },   // Warm-up
    { duration: '10s', target: 30 },  // Moderate load
    { duration: '10s', target: 60 },  // High load
    { duration: '10s', target: 100 }, // Peak stress
    { duration: '5s', target: 0 },    // Ramp-down
  ],
  thresholds: {
    'stress_success_rate': ['rate>0.95'],
  },
};

const GRAPHQL_URL = 'http://localhost:8080/query';

const sampleQueries = [
  JSON.stringify({ query: '{ getAllCompanies { id companyName } }' }),
  JSON.stringify({ query: '{ getCompanyByID(companyId: "4") { id companyName } }' }),
  JSON.stringify({ query: '{ getAllPosts { id title content } }' }),
];

export default function () {
  const randomIp = `10.${Math.floor(Math.random() * 200)}.${Math.floor(Math.random() * 250)}.${Math.floor(Math.random() * 250)}`;
  const q = sampleQueries[Math.floor(Math.random() * sampleQueries.length)];

  const params = {
    headers: {
      'Content-Type': 'application/json',
      'X-Forwarded-For': randomIp,
      'True-Client-Ip': randomIp,
    },
  };

  const res = http.post(GRAPHQL_URL, q, params);

  stressReqCount.add(1);
  stressLatency.add(res.timings.duration);
  stressTtfb.add(res.timings.waiting);

  const ok = check(res, {
    'status is 200': (r) => r.status === 200,
  });

  stressSuccessRate.add(ok ? 1 : 0);
  sleep(0.01);
}
