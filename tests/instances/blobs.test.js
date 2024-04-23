import { beforeAll, describe, expect, it } from '@jest/globals'

import common from '../common'
import helpers from '../common/helpers'
import request from '../common/request'

const namespaceName = 'blobstest'

let id = ''

describe('Test wait success API behaviour', () => {
	beforeAll(common.helpers.deleteAllNamespaces)

	helpers.itShouldCreateNamespace(it, expect, namespaceName)

	helpers.itShouldCreateFileV2(it, expect, namespaceName,
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

	it(`should invoke the 'noop.yaml' workflow`, async () => {
		const req = await request(common.config.getDirektivHost()).post(`/api/v2/namespaces/${ namespaceName }/instances?path=noop.yaml`)
		expect(req.statusCode).toEqual(200)

		id = req.body.data.id

		await sleep(200)
	})

	it(`should get the instance's input data`, async () => {
		const req = await request(common.config.getDirektivHost()).get(`/api/v2/namespaces/${ namespaceName }/instances/${ id }/input`)

		expect(req.statusCode).toEqual(200)
		expect(req.body).toMatchObject({
			data: {},
		})
	})

	it(`should get the instance's output data`, async () => {
		const req = await request(common.config.getDirektivHost()).get(`/api/v2/namespaces/${ namespaceName }/instances/${ id }/output`)

		expect(req.statusCode).toEqual(200)
		expect(req.body).toMatchObject({
			data: {
				output: 'eyJyZXN1bHQiOiJ4In0=',
			},
		})
	})
})

function sleep (ms) {
	return new Promise(resolve => setTimeout(resolve, ms))
}
