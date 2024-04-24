import { beforeAll, describe, expect, it } from '@jest/globals'

import common from '../common'
import helpers from '../common/helpers'
import request from '../common/request'

const namespaceName = 'datatest'

describe('Test instance data behaviour', () => {
	beforeAll(common.helpers.deleteAllNamespaces)

	helpers.itShouldCreateNamespace(it, expect, namespaceName)

	helpers.itShouldCreateFileV2(it, expect, namespaceName,
		'',
		'data.yaml',
		'workflow',
		'text/plain',
		btoa(`
states:
- id: a
  type: noop
`))

	it(`should invoke the '/data.yaml' workflow with no input`, async () => {
		const req = await request(common.config.getDirektivHost()).get(`/api/namespaces/${ namespaceName }/tree/data.yaml?op=wait`)
		expect(req.statusCode).toEqual(200)
		expect(req.body).toMatchObject({})
	})

	it(`should invoke the '/data.yaml' workflow with a simple object input`, async () => {
		const req = await request(common.config.getDirektivHost()).get(`/api/namespaces/${ namespaceName }/tree/data.yaml?op=wait`)
			.send(`{"x": 5}`)
		expect(req.statusCode).toEqual(200)
		expect(req.body).toMatchObject({
			x: 5,
		})
	})

	it(`should invoke the '/data.yaml' workflow with a json non-object input`, async () => {
		const req = await request(common.config.getDirektivHost()).get(`/api/namespaces/${ namespaceName }/tree/data.yaml?op=wait`)
			.send(`[1, 2, 3]`)
		expect(req.statusCode).toEqual(200)
		expect(req.body).toMatchObject({
			input: [ 1, 2, 3 ],
		})
	})

	it(`should invoke the '/data.yaml' workflow with a non-json input`, async () => {
		const req = await request(common.config.getDirektivHost()).get(`/api/namespaces/${ namespaceName }/tree/data.yaml?op=wait`)
			.send(`Hello, world!`)
		expect(req.statusCode).toEqual(200)
		expect(req.body).toMatchObject({
			input: 'SGVsbG8sIHdvcmxkIQ==',
		})
	})
})
