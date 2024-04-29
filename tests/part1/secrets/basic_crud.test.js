import { beforeAll, describe, expect, it } from '@jest/globals'
import { basename } from 'path'

import config from '../../common/config'
import helpers from '../../common/helpers'
import regex from '../../common/regex'
import request from '../../common/request'

const namespace = basename(__filename)

describe('Test secrets get delete list calls', () => {
	beforeAll(helpers.deleteAllNamespaces)
	helpers.itShouldCreateNamespace(it, expect, namespace)

	it(`should create a new secret foo1`, async () => {
		const res = await request(config.getDirektivHost())
			.post(`/api/v2/namespaces/${ namespace }/secrets`)
			.send(makeDummySecret('foo1'))
		expect(res.statusCode).toEqual(200)
	})

	it(`should create a new secret foo2`, async () => {
		const res = await request(config.getDirektivHost())
			.post(`/api/v2/namespaces/${ namespace }/secrets`)
			.send(makeDummySecret('foo2'))
		expect(res.statusCode).toEqual(200)
	})

	it(`should get the new secret foo1`, async () => {
		const res = await request(config.getDirektivHost())
			.get(`/api/v2/namespaces/${ namespace }/secrets/foo1`)
		expect(res.statusCode).toEqual(200)

		expect(res.body.data).toEqual(expectDummySecret('foo1'))
	})

	it(`should get the new secret foo2`, async () => {
		const res = await request(config.getDirektivHost())
			.get(`/api/v2/namespaces/${ namespace }/secrets/foo2`)
		expect(res.statusCode).toEqual(200)
		expect(res.body.data).toEqual(expectDummySecret('foo2'))
	})

	it(`should list foo1 and foo2`, async () => {
		const res = await request(config.getDirektivHost())
			.get(`/api/v2/namespaces/${ namespace }/secrets`)
		expect(res.statusCode).toEqual(200)
		expect(res.body).toEqual({
			data: [ expectDummySecret('foo1'), expectDummySecret('foo2') ],
		})
	})

	it(`should delete foo1`, async () => {
		const res = await request(config.getDirektivHost())
			.delete(`/api/v2/namespaces/${ namespace }/secrets/foo1`)
		expect(res.statusCode).toEqual(200)
	})

	it(`should list foo1 and foo2`, async () => {
		const res = await request(config.getDirektivHost())
			.get(`/api/v2/namespaces/${ namespace }/secrets`)
		expect(res.statusCode).toEqual(200)
		expect(res.body).toEqual({
			data: [ expectDummySecret('foo2') ],
		})
	})
})

function makeDummySecret (name) {
	return {
		name,
		data: btoa('value of' + name),
	}
}

function expectDummySecret (name) {
	return {
		name,
		initialized: true,
		createdAt: expect.stringMatching(regex.timestampRegex),
		updatedAt: expect.stringMatching(regex.timestampRegex),
	}
}
