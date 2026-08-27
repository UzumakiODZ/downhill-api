import http from 'k6/http';
import { check, sleep } from 'k6';
import { Trend, Rate, Counter } from 'k6/metrics';

// Custom performance metrics
const queryDuration = new Trend('graphql_query_duration');
const ttfbTrend = new Trend('graphql_ttfb');
const successRate = new Rate('graphql_success_rate');
const requestsCount = new Counter('graphql_total_requests');

export const options = {
  scenarios: {
    cached_reads: {
      executor: 'constant-vus',
      vus: 20,
      duration: '20s',
    },
  },
  thresholds: {
    'http_req_duration': ['p(95)<100', 'p(99)<200'],
    'graphql_success_rate': ['rate>0.99'],
  },
};

const GRAPHQL_URL = 'http://localhost:8080/query';

const queries = [
  {
    name: 'GetAllCompanies',
    body: JSON.stringify({
      query: `query GetAllCompanies {
        getAllCompanies {
          id
          companyName
        }
      }`,
    }),
  },
  {
    name: 'GetCompanyByID',
    body: JSON.stringify({
      query: `query GetCompanyByID {
        getCompanyByID(companyId: "4") {
          id
          companyName
        }
      }`,
    }),
  },
  {
    name: 'GetAllPosts',
    body: JSON.stringify({
      query: `query GetAllPosts {
        getAllPosts {
          id
          title
          content
        }
      }`,
    }),
  },
  {
    name: 'GetRolesByCompany',
    body: JSON.stringify({
      query: `query GetRolesByCompany {
        getRolesByCompany(companyId: "4") {
          id
          roleName
          year
          companyId
          offerType
          cgpa
          ctc
          base
        }
      }`,
    }),
  },
];

export default function () {
  const randomIp = `192.168.${Math.floor(Math.random() * 250) + 1}.${Math.floor(Math.random() * 250) + 1}`;
  const queryObj = queries[Math.floor(Math.random() * queries.length)];

  const params = {
    headers: {
      'Content-Type': 'application/json',
      'X-Forwarded-For': randomIp,
      'True-Client-Ip': randomIp,
    },
    tags: { name: queryObj.name },
  };

  const res = http.post(GRAPHQL_URL, queryObj.body, params);

  requestsCount.add(1);
  ttfbTrend.add(res.timings.waiting);
  queryDuration.add(res.timings.duration);

  const isOk = check(res, {
    'status is 200': (r) => r.status === 200,
    'has data field': (r) => {
      try {
        const body = JSON.parse(r.body);
        return body.data !== undefined && body.errors === undefined;
      } catch (e) {
        return false;
      }
    },
  });

  successRate.add(isOk ? 1 : 0);
  sleep(0.02);
}
