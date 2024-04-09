import { beforeAll, describe, expect, it } from '@jest/globals'
import { basename } from 'path'

import common from '../common'
import config from '../common/config'
import helpers from '../common/helpers'
import regex from '../common/regex'
import request from '../common/request'

const namespace = basename(__filename)

describe('Test no name duplicated', () => {
	beforeAll(helpers.deleteAllNamespaces)
	helpers.itShouldCreateNamespace(it, expect, namespace)

	helpers.itShouldCreateFileV2(it, expect, namespace, '/', 'wf1.yaml', 'workflow', 'text',
		btoa(helpers.dummyWorkflow('wf1.yaml')))

	const foo1 =  {
		name: 'foo1',
			data: btoa('bar1'),
			mimeType: 'mime1',
	}

	it(`should create a new variable foo1`, async () => {
		const res = await request(config.getDirektivHost())
			.post(`/api/v2/namespaces/${ namespace }/variables`)
			.send(foo1)
		expect(res.statusCode).toEqual(200)
	})

	it(`should not duplicate a new variable foo1`, async () => {
		const res = await request(config.getDirektivHost())
			.post(`/api/v2/namespaces/${ namespace }/variables`)
			.send(foo1)
		expect(res.statusCode).toEqual(400)
		expect(res.body).toEqual(
			{
				error: {
					code: "resource_already_exists",
					message: "resource already exists",
	},
			}
		)
	})
})

