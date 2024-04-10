import { beforeAll, describe, expect, it } from '@jest/globals'
import { basename } from 'path'

import common from '../common'
import config from '../common/config'
import helpers from '../common/helpers'
import regex from '../common/regex'
import request from '../common/request'

const namespace = basename(__filename)

describe('Test variable list calls', () => {
	beforeAll(helpers.deleteAllNamespaces)
	helpers.itShouldCreateNamespace(it, expect, namespace)

	helpers.itShouldCreateFileV2(it, expect, namespace, '/', 'wf.yaml', 'workflow', 'text',
		btoa(helpers.dummyWorkflow('wf.yaml')))

	helpers.itShouldCreateFileV2()

	helpers.itShouldCreateVariableV2(it, expect, namespace, {
		name: 'foo1',
		data: btoa('foo1'),
		mimeType: 'mime_foo1',
	})

	helpers.itShouldCreateVariableV2(it, expect, namespace, {
		name: 'foo2',
		data: btoa('foo2'),
		mimeType: 'mime_foo2',
		workflowPath: '/wf.yaml',
	})

	it(`should list variable foo1`, async () => {
		const res = await request(config.getDirektivHost())
			.get(`/api/v2/namespaces/${ namespace }/variables`)
		expect(res.statusCode).toEqual(200)
		expect(res.body.data.length).toEqual(1)
		expect(res.body.data[0]).toEqual({
			id: expect.stringMatching(common.regex.uuidRegex),

			name: 'foo1',
			mimeType: 'mime_foo1',
			size: 4,
			type: 'namespace-variable',
			reference: namespace,

			createdAt: expect.stringMatching(regex.timestampRegex),
			updatedAt: expect.stringMatching(regex.timestampRegex),
		})
	})

	it(`should list variable foo2`, async () => {
		const res = await request(config.getDirektivHost())
			.get(`/api/v2/namespaces/${ namespace }/variables?workflowPath=/wf.yaml`)
		expect(res.statusCode).toEqual(200)
		console.log(res.body.data)
		expect(res.body.data.length).toEqual(1)
		expect(res.body.data[0]).toEqual({
			id: expect.stringMatching(common.regex.uuidRegex),

			name: 'foo2',
			mimeType: 'mime_foo2',
			size: 4,
			type: 'workflow-variable',
			reference: '/wf.yaml',

			createdAt: expect.stringMatching(regex.timestampRegex),
			updatedAt: expect.stringMatching(regex.timestampRegex),
		})
	})
})
