import { beforeAll, describe, expect, it } from '@jest/globals'
import { basename } from 'path'

import config from '../../common/config'
import helpers from '../../common/helpers'
import regex from '../../common/regex'
import request from '../../common/request'

const namespace = basename(__filename)

describe('Test apitokens get delete list calls', () => {
	beforeAll(helpers.deleteAllNamespaces)
	helpers.itShouldCreateNamespace(it, expect, namespace)

	it(`should create a new apitoken foo1`, async () => {
		const res = await request(config.getDirektivHost())
			.post(`/api/v2/namespaces/${ namespace }/apitokens`)
			.send(makeDummyAPIToken('foo1'))
		expect(res.statusCode).toEqual(200)
	})

	it(`should create a new apitoken foo2`, async () => {
		const res = await request(config.getDirektivHost())
			.post(`/api/v2/namespaces/${ namespace }/apitokens`)
			.send(makeDummyAPIToken('foo2'))
		expect(res.statusCode).toEqual(200)
	})

	it(`should get the new apitoken foo1`, async () => {
		const res = await request(config.getDirektivHost())
			.get(`/api/v2/namespaces/${ namespace }/apitokens/foo1`)
		expect(res.statusCode).toEqual(200)

		expect(res.body.data).toEqual(expectDummyAPIToken('foo1'))
	})

	it(`should get the new apitoken foo2`, async () => {
		const res = await request(config.getDirektivHost())
			.get(`/api/v2/namespaces/${ namespace }/apitokens/foo2`)
		expect(res.statusCode).toEqual(200)
		expect(res.body.data).toEqual(expectDummyAPIToken('foo2'))
	})

	it(`should list foo1 and foo2`, async () => {
		const res = await request(config.getDirektivHost())
			.get(`/api/v2/namespaces/${ namespace }/apitokens`)
		expect(res.statusCode).toEqual(200)
		expect(res.body).toEqual({
			data: [ expectDummyAPIToken('foo1'), expectDummyAPIToken('foo2') ],
		})
	})

	it(`should delete foo1`, async () => {
		const res = await request(config.getDirektivHost())
			.delete(`/api/v2/namespaces/${ namespace }/apitokens/foo1`)
		expect(res.statusCode).toEqual(200)
	})

	it(`should list foo1 and foo2`, async () => {
		const res = await request(config.getDirektivHost())
			.get(`/api/v2/namespaces/${ namespace }/apitokens`)
		expect(res.statusCode).toEqual(200)
		expect(res.body).toEqual({
			data: [ expectDummyAPIToken('foo2') ],
		})
	})
})

function makeDummyAPIToken (name) {
	return {
		name,
		description: name + ' description',
		topics: [name + '_topic1', name + '_topic2'],
		methods: [name + '_method1', name + '_method2'],
	}
}

function expectDummyAPIToken (name) {
	return {
		name,
		description: name + ' description',
		prefix: expect.anything(),
		topics: [name + '_topic1', name + '_topic2'],
		methods: [name + '_method1', name + '_method2'],
		createdAt: expect.stringMatching(regex.timestampRegex),
		updatedAt: expect.stringMatching(regex.timestampRegex),
	}
}
