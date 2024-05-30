import { beforeAll, describe, expect, it } from '@jest/globals'
import { basename } from 'path'

import config from '../../common/config'
import helpers from '../../common/helpers'
import request from '../../common/request'

const namespace = basename(__filename)

describe('Test variable workflow links', () => {
	beforeAll(helpers.deleteAllNamespaces)
	helpers.itShouldCreateNamespace(it, expect, namespace)

	helpers.itShouldCreateYamlFile(it, expect, namespace, '/', 'wf1.yaml', 'workflow', `
direktiv_api: workflow/v1
description: A simple 'no-op' state that returns 'Hello world!'
states:
- id: step1
  type: noop
  transform:
    result: Hello world!
`)
	const foo = {
		name: 'foo',
		data: btoa('bar'),
		mimeType: 'mime',
		workflowPath: '/wf1.yaml',
	}

	let fooId
	it(`should create a new workflow variable foo`, async () => {
		const res = await request(config.getDirektivHost())
			.post(`/api/v2/namespaces/${ namespace }/variables`)
			.send(foo)
		expect(res.statusCode).toEqual(200)
		fooId = res.body.data.id
	})

	helpers.itShouldUpdateFilePathV2(it, expect, namespace, '/wf1.yaml', '/wf2.yaml')

	it(`should read new path in variable foo`, async () => {
		const res = await request(config.getDirektivHost())
			.get(`/api/v2/namespaces/${ namespace }/variables/${ fooId }`)
			.send(foo)
		expect(res.body.data.reference).toEqual('/wf2.yaml')
	})

	helpers.itShouldDeleteFile(it, expect, namespace, '/wf2.yaml')

	it(`should read 404 variable foo`, async () => {
		const res = await request(config.getDirektivHost())
			.get(`/api/v2/namespaces/${ namespace }/variables/${ fooId }`)
			.send(foo)
		expect(res.statusCode).toEqual(404)
	})
})
