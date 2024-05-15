import { beforeAll, describe, expect, it } from '@jest/globals'
import { basename } from 'path'

import common from '../../common'
import helpers from '../../common/helpers'
import request from '../../common/request'

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
	it(`generate some logs`, async () => {
		const res = await request(common.config.getDirektivHost()).post(`/api/v2/namespaces/${ namespace }/instances?path=noop.yaml&wait=true`)
		expect(res.statusCode).toEqual(200)
	})
	// retry70(`should contain instance log entries`, async () => {
	// 	const instRes = await request(common.config.getDirektivHost()).get(`/api/namespaces/${ namespace }/instances`)
	// 	expect(instRes.statusCode).toEqual(200)
	//
	// 	const logRes = await request(common.config.getDirektivHost()).get(`/api/v2/namespaces/${ namespace }/logs?instance=${ instRes.body.instances.results[0].id }`)
	// 	expect(logRes.statusCode).toEqual(200)
	// 	expect(logRes.body.data).not.toBeNull()
	// 	expect(logRes.body.data.length).toBeGreaterThanOrEqual(1)
	// },
	// )
	// retry50(`should contain namespace log entries`, async () => {
	// 	const logRes = await request(common.config.getDirektivHost()).get(`/api/v2/namespaces/${ namespace }/logs`)
	// 	expect(logRes.statusCode).toEqual(200)
	// 	expect(logRes.body.data).not.toBeNull()
	// 	expect(logRes.body.data).not.toBeNull()
	// 	expect(logRes.body.data.length).toBeGreaterThanOrEqual(1)
	// })
})
