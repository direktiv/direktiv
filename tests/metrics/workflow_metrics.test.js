import { beforeAll, describe, expect, it } from '@jest/globals'
import { basename } from 'path'

import common from '../common'
import config from '../common/config'
import helpers from '../common/helpers'
import request from '../common/request'

const namespace = basename(__filename)

describe('Test workflow metrics', () => {
	beforeAll(helpers.deleteAllNamespaces)

	helpers.itShouldCreateNamespace(it, expect, namespace)

	it(`should read no results`, async () => {
		const res = await request(config.getDirektivHost())
			.get(`/api/v2/namespaces/${ namespace }/metrics/instances?workflowPath=%2Ffoo1.yaml`)

		expect(res.statusCode).toEqual(200)
		expect(res.body).toEqual({
			data: {
				cancelled: 0,
				crashed: 0,
				failed: 0,
				pending: 0,
				complete: 0,
				total: 0,
			},
		})
	})

	helpers.itShouldCreateFileV2(it, expect, namespace,
		'/',
		'foo1.yaml',
		'workflow',
		'text/plain',
		btoa(`
direktiv_api: workflow/v1
states:
- id: a
  type: noop
`))

	it(`should invoke the 'foo1.yaml' workflow`, async () => {
		const req = await request(common.config.getDirektivHost()).post(`/api/v2/namespaces/${ namespace }/instances?path=foo1.yaml&wait=true`)

		expect(req.statusCode).toEqual(200)
	})

	it(`should read one result`, async () => {
		const res = await request(config.getDirektivHost())
			.get(`/api/v2/namespaces/${ namespace }/metrics/instances?workflowPath=%2Ffoo1.yaml`)
		expect(res.statusCode).toEqual(200)
		expect(res.body).toEqual({
			data: {
				cancelled: 0,
				crashed: 0,
				failed: 0,
				pending: 0,
				complete: 1,
				total: 1,
			},
		})
	})
})
