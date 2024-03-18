import { beforeAll, describe, expect, it } from '@jest/globals'

import config from '../common/config'
import helpers from '../common/helpers'
import regex from '../common/regex'
import request from '../common/request'

describe('Test namespaces get delete list calls', () => {
	beforeAll(helpers.deleteAllNamespaces)

	it(`should create a new namespace foo1`, async () => {
		const res = await request(config.getDirektivHost())
			.post(`/api/v2/namespaces`)
			.send(makeDummyNamespace('foo1'))
		expect(res.statusCode).toEqual(200)
	})

	it(`should create a new namespace foo2`, async () => {
		const res = await request(config.getDirektivHost())
			.post(`/api/v2/namespaces`)
			.send(makeDummyNamespace('foo2'))
		expect(res.statusCode).toEqual(200)
	})

	it(`should get the new namespace foo1`, async () => {
		const res = await request(config.getDirektivHost())
			.get(`/api/v2/namespaces/foo1`)
		expect(res.statusCode).toEqual(200)

		expect(res.body.data).toEqual(expectDummyNamespace('foo1'))
	})

	it(`should get the new namespace foo2`, async () => {
		const res = await request(config.getDirektivHost())
			.get(`/api/v2/namespaces/foo2`)
		expect(res.statusCode).toEqual(200)
		expect(res.body.data).toEqual(expectDummyNamespace('foo2'))
	})

	it(`should list foo1 and foo2`, async () => {
		const res = await request(config.getDirektivHost())
			.get(`/api/v2/namespaces`)
		expect(res.statusCode).toEqual(200)
		expect(res.body).toEqual({
			data: [ expectDummyNamespace('foo1'), expectDummyNamespace('foo2') ],
		})
	})

	it(`should delete foo1`, async () => {
		const res = await request(config.getDirektivHost())
			.delete(`/api/v2/namespaces/foo1`)
		expect(res.statusCode).toEqual(200)
	})

	it(`should list foo1 and foo2`, async () => {
		const res = await request(config.getDirektivHost())
			.get(`/api/v2/namespaces`)
		expect(res.statusCode).toEqual(200)
		expect(res.body).toEqual({
			data: [ expectDummyNamespace('foo2') ],
		})
	})
})

function makeDummyNamespace (name) {
	return {
		name,
		data: btoa('value of' + name),
	}
}

function expectDummyNamespace (name) {
	return {
		name,
		createdAt: expect.stringMatching(regex.timestampRegex),
		updatedAt: expect.stringMatching(regex.timestampRegex),
	}
}
