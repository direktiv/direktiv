import { beforeAll, describe, expect, it } from '@jest/globals'
import { basename } from 'path'
import request from 'supertest'

import common from '../common'
import config from '../common/config'
import helpers from '../common/helpers'
import regex from '../common/regex'

const namespace = basename(__filename)


describe('Test workflow variable operations', () => {
	beforeAll(helpers.deleteAllNamespaces)
	helpers.itShouldCreateNamespace(it, expect, namespace)

	helpers.itShouldCreateFileV2(it, expect, namespace, '/', 'foo1', 'workflow', 'text',
		btoa(helpers.dummyWorkflow('foo1')))

	// todo: check zero list.

	it(`should create a new variable`, async () => {
		const res = await request(config.getDirektivHost())
			.post(`/api/v2/namespaces/${ namespace }/variables`)
			.send({
				name: 'foo1',
				data: 'bar1',
				mimeType: 'mime1',
			})
		expect(res.statusCode).toEqual(200)
		expect(res.body.data).toMatchObject({
			id: expect.stringMatching(common.regex.uuidRegex),
			name: 'foo1',
			data: 'bar1',
			mimeType: 'mime1',
			createdAt: expect.stringMatching(regex.timestampRegex),
			updatedAt: expect.stringMatching(regex.timestampRegex),
		})
	})

	it(`should create a new variable`, async () => {
		const res = await request(config.getDirektivHost())
			.post(`/api/v2/namespaces/${ namespace }/variables`)
			.send({
				name: 'foo1',
				data: 'bar1',
				mimeType: 'mime1',
			})
		expect(res.statusCode).toEqual(200)
		expect(res.body.data).toMatchObject({
			id: expect.stringMatching(common.regex.uuidRegex),
			name: 'foo1',
			data: 'bar1',
			mimeType: 'mime1',
			createdAt: expect.stringMatching(regex.timestampRegex),
			updatedAt: expect.stringMatching(regex.timestampRegex),
		})
	})
})
