import { beforeAll, describe, expect, it } from '@jest/globals'
import { basename } from 'path'

import common from '../../common'
import config from '../../common/config'
import helpers from '../../common/helpers'
import request from '../../common/request'

const namespace = basename(__filename.replaceAll('.', '-'))

describe('Test workflow metrics', () => {
	beforeAll(helpers.deleteAllNamespaces)

	helpers.itShouldCreateNamespace(it, expect, namespace)

	it(`should read no results`, async () => {
		const res = await request(config.getDirektivBaseUrl())
			.get(`/api/v2/namespaces/${ namespace }/metrics/instances?workflowPath=/foo1.wf.ts`)

		expect(res.statusCode).toEqual(404)
		expect(res.body.error).toEqual({
			     code: 'not_found',
			     message: 'requested resource is not found',
		})
	})

	helpers.itShouldCreateFile(it, expect, namespace,
		'/',
		'foo1.wf.ts',
		'workflow',
		'application/x-typescript',
		btoa(`
function stateOne(payload) {
	print("RUN STATE FIRST");
	payload.bar = "foo";
	return finish(payload);
}
`))

	it(`should invoke the '/foo1.wf.ts' workflow`, async () => {
		const res = await request(common.config.getDirektivBaseUrl())
			.post(`/api/v2/namespaces/${ namespace }/instances?path=foo1.wf.ts&wait=true`)
			.send({ foo: 'bar' })
		expect(res.statusCode).toEqual(200)
	})

	it(`should read one result`, async () => {
		const res = await request(config.getDirektivBaseUrl())
			.get(`/api/v2/namespaces/${ namespace }/metrics/instances?workflowPath=/foo1.wf.ts`)
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
