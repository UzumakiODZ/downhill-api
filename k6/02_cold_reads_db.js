import http from 'k6/http';
import { check, sleep } from 'k6';
import { Trend, Rate, Counter } from 'k6/metrics';

const dbQueryDuration = new Trend('db_query_duration');
const dbTtfbTrend = new Trend('db_ttfb');
const dbSuccessRate = new Rate('db_success_rate');

export const options = {
  scenarios: {
    cold_db_reads: {
      executor: 'constant-vus',
      vus: 5,
      duration: '15s',
    },
  },
};

const GRAPHQL_URL = 'http://localhost:8080/query';

export default function (data) {
  const randomIp = `10.0.${Math.floor(Math.random() * 250) + 1}.${Math.floor(Math.random() * 250) + 1}`;
  // Using varying company IDs / non-existent or fresh IDs to trigger DB queries
  const randomCompanyId = Math.floor(Math.random() * 20) + 1;

  const query = JSON.stringify({
    query: `query GetCompanyAndRoles($id: ID!) {
      getCompanyByID(companyId: $id) {
        id
        companyName
      }
      getRolesByCompany(companyId: $id) {
        id
        roleName
        cgpa
        ctc
      }
    }`,
    variables: { id: String(randomCompanyId) },
  });

  const params = {
    headers: {
      'Content-Type': 'application/json',
      'X-Forwarded-For': randomIp,
      'True-Client-Ip': randomIp,
    },
  };

  const res = http.post(GRAPHQL_URL, query, params);

  dbTtfbTrend.add(res.timings.waiting);
  dbQueryDuration.add(res.timings.duration);

  const isOk = check(res, {
    'status is 200 or 404 handled': (r) => r.status === 200,
  });

  dbSuccessRate.add(isOk ? 1 : 0);
  sleep(0.1);
}
