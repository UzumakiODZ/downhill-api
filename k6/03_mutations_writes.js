import http from 'k6/http';
import { check, sleep } from 'k6';
import { Trend, Rate, Counter } from 'k6/metrics';

const mutationDuration = new Trend('mutation_duration');
const mutationTtfb = new Trend('mutation_ttfb');
const mutationSuccessRate = new Rate('mutation_success_rate');
const createdEntities = new Counter('created_entities_count');

export const options = {
  scenarios: {
    write_mutations: {
      executor: 'constant-vus',
      vus: 5,
      duration: '15s',
    },
  },
};

const GRAPHQL_URL = 'http://localhost:8080/query';

export default function () {
  const randomIp = `172.16.${Math.floor(Math.random() * 250) + 1}.${Math.floor(Math.random() * 250) + 1}`;
  const timestamp = Date.now() + Math.floor(Math.random() * 10000);
  const companyName = `PerfCorp_${timestamp}`;

  const mutation = JSON.stringify({
    query: `mutation CreateCompany($compInput: CreateCompanyInput!) {
      createCompany(input: $compInput) {
        id
        companyName
      }
    }`,
    variables: {
      compInput: {
        companyName: companyName,
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

  mutationTtfb.add(res.timings.waiting);
  mutationDuration.add(res.timings.duration);

  let success = false;
  try {
    const json = JSON.parse(res.body);
    if (json.data && json.data.createCompany && json.data.createCompany.id) {
      success = true;
      createdEntities.add(1);
    }
  } catch (e) {}

  const isOk = check(res, {
    'status is 200': (r) => r.status === 200,
    'mutation succeeded': () => success,
  });

  mutationSuccessRate.add(isOk ? 1 : 0);
  sleep(0.1);
}
