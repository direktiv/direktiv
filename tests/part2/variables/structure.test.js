import { beforeAll, describe, expect, it } from '@jest/globals'
import { basename } from 'path'

import common from '../../common'
import config from '../../common/config'
import helpers from '../../common/helpers'
import regex from '../../common/regex'
import request from '../../common/request'

const namespace = basename(__filename)

describe('Test variable get delete list calls', () => {
	beforeAll(helpers.deleteAllNamespaces)
	helpers.itShouldCreateNamespace(it, expect, namespace)

	let createRes
	it(`should create a new variable foo`, async () => {
		createRes = await request(config.getDirektivHost())
			.post(`/api/v2/namespaces/${ namespace }/variables`)
			.send({
				name: 'foo',
				data: btoa('foo'),
				mimeType: 'mime_foo',
			})
		expect(createRes.statusCode).toEqual(200)
	})

	it(`should get the new variable foo`, async () => {
		const res = await request(config.getDirektivHost())
			.get(`/api/v2/namespaces/${ namespace }/variables/${ createRes.body.data.id }`)
		expect(res.statusCode).toEqual(200)
		expect(res.body.data).toEqual({
			id: expect.stringMatching(common.regex.uuidRegex),

			name: 'foo',
			data: btoa('foo'),
			mimeType: 'mime_foo',
			size: 3,
			type: 'namespace-variable',
			reference: namespace,

			createdAt: expect.stringMatching(regex.timestampRegex),
			updatedAt: expect.stringMatching(regex.timestampRegex),
		})
	})

	it(`should list the new variable foo`, async () => {
		const res = await request(config.getDirektivHost())
			.get(`/api/v2/namespaces/${ namespace }/variables`)
		expect(res.statusCode).toEqual(200)
		expect(res.body.data.length).toEqual(1)
		expect(res.body.data[0]).toEqual({
			id: expect.stringMatching(common.regex.uuidRegex),

			name: 'foo',

			mimeType: 'mime_foo',
			size: 3,
			type: 'namespace-variable',
			reference: namespace,

			createdAt: expect.stringMatching(regex.timestampRegex),
			updatedAt: expect.stringMatching(regex.timestampRegex),
		})
	})
})
