import { beforeAll, describe, expect, it } from '@jest/globals'
import { basename } from 'path'

import config from '../common/config'
import helpers from '../common/helpers'
import request from '../common/request'

const namespace = basename(__filename)

describe('Test no namespace variable name duplicated', () => {
	beforeAll(helpers.deleteAllNamespaces)
	helpers.itShouldCreateNamespace(it, expect, namespace)

	let foo = {
		name: 'foo',
		data: btoa('bar'),
		mimeType: 'mime',
	}

	it(`should create a new namespace variable foo`, async () => {
		const res = await request(config.getDirektivHost())
			.post(`/api/v2/namespaces/${ namespace }/variables`)
			.send(foo)
		expect(res.statusCode).toEqual(200)
	})

	it(`should not duplicate a namespace variable foo`, async () => {
		const res = await request(config.getDirektivHost())
			.post(`/api/v2/namespaces/${ namespace }/variables`)
			.send(foo)
		expect(res.statusCode).toEqual(400)
		expect(res.body).toEqual(
			{
				error: {
					code: 'resource_already_exists',
					message: 'resource already exists',
				},
			},
		)
	})

	helpers.itShouldCreateFileV2(it, expect, namespace, '/', 'wf1.yaml', 'workflow', 'text',
		btoa(helpers.dummyWorkflow('wf1.yaml')))

	foo = {
		name: 'foo',
		data: btoa('bar'),
		mimeType: 'mime',
		workflowPath: '/wf1.yaml'
	}

	it(`should allow create workflow variable foo`, async () => {
		const res = await request(config.getDirektivHost())
			.post(`/api/v2/namespaces/${ namespace }/variables`)
			.send(foo)
		expect(res.statusCode).toEqual(200)
	})

	it(`should not duplicate workflow variable foo`, async () => {
		const res = await request(config.getDirektivHost())
			.post(`/api/v2/namespaces/${ namespace }/variables`)
			.send(foo)
		expect(res.statusCode).toEqual(400)
		expect(res.body).toEqual(
			{
				error: {
					code: 'resource_already_exists',
					message: 'resource already exists',
				},
			},
		)
	})
})
