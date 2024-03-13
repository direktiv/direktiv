import { beforeAll, describe, expect, it } from '@jest/globals'
import { basename } from 'path'

import common from '../common'
import config from '../common/config'
import helpers from '../common/helpers'
import regex from '../common/regex'
import request from '../common/request'

const namespace = basename(__filename)

describe('Test log api calls', () => {
	beforeAll(helpers.deleteAllNamespaces)
	helpers.itShouldCreateNamespace(it, expect, namespace)

	beforeAll(common.helpers.deleteAllNamespaces)

	helpers.itShouldCreateFileV2(it, expect, namespace,
		'',
		'noop.yaml',
		'workflow',
		'text/plain',
		btoa(`
states:
- id: a
  type: noop
  transform:
    result: x`))

	it(`should contain instance log entries`, async () => {
		const req = await request(common.config.getDirektivHost()).get(`/api/namespaces/${ namespace }/tree/noop.yaml?op=wait`)
		expect(req.statusCode).toEqual(200)
		const req1 = await request(common.config.getDirektivHost()).get(`/api/namespaces/${ namespace }/instances`)
		expect(req.statusCode).toEqual(200)

		const req2 = await requestWithRetry(`/api/v2/namespaces/${namespace}/logs?instance=${req1.body.instances.results[0].id}`);
		expect(req2.statusCode).toEqual(200);
		if (req.body.data == null) {
			expect(false).toBeTruthy
		}
		if (req.body.data != null) {
			expect(req.body.data.length).toBeGreaterThanOrEqual(1)
		}
		
	})
	it(`should contain namespace log entries`, async () => {
		const req = await requestWithRetry(`/api/v2/namespaces/${ namespace }/logs`)
		expect(req.statusCode).toEqual(200)
		if (req.body.data == null) {
			expect(false).toBeTruthy
		}
		if (req.body.data != null) {
			expect(req.body.data.length).toBeGreaterThanOrEqual(1)
		}
	})
})

const requestWithRetry = (url, retries = 10) => {
	return new Promise((resolve, reject) => {
	  const attempt = () => {
		request(common.config.getDirektivHost()).get(url)
		  .then(response => {
			if (response.statusCode === 200) {
			  resolve(response);
			} else {
			  if (retries > 0) {
				console.log(`Retrying... ${retries} attempts left`);
				attempt(retries - 1);
			  } else {
				reject(`Failed after several attempts`);
			  }
			}
		  })
		  .catch(error => {
			if (retries > 0) {
			  console.log(`Retrying due to error... ${retries} attempts left`);
			  attempt(retries - 1);
			} else {
			  reject(error);
			}
		  });
	  };
	  attempt();
	});
  };