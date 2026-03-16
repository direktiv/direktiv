import { beforeAll, describe, expect, it } from '@jest/globals'

import { basename } from 'path'
import config from '../../common/config'
import { fileURLToPath } from 'url'
import helpers from '../../common/helpers'
import regex from '../../common/regex'
import request from '../../common/request'

const namespace = basename(fileURLToPath(import.meta.url))

describe('Test secrets get delete list calls', () => {
	beforeAll(helpers.deleteAllNamespaces)
	helpers.itShouldCreateNamespace(it, expect, namespace)

	it(`should create a new secret foo-one`, async () => {
		const res = await request(config.getDirektivBaseUrl())
			.post(`/api/v2/namespaces/${namespace}/secrets`)
			.send(makeDummySecret('foo-one'))
		expect(res.statusCode).toEqual(200)
	})

	it(`should create a new secret foo-two`, async () => {
		const res = await request(config.getDirektivBaseUrl())
			.post(`/api/v2/namespaces/${namespace}/secrets`)
			.send(makeDummySecret('foo-two'))
		expect(res.statusCode).toEqual(200)
	})

	it(`should get the new secret foo-one`, async () => {
		const res = await request(config.getDirektivBaseUrl()).get(
			`/api/v2/namespaces/${namespace}/secrets/foo-one`,
		)
		expect(res.statusCode).toEqual(200)

		expect(res.body.data).toEqual(expectDummySecret('foo-one'))
	})

	it(`should get the new secret foo-two`, async () => {
		const res = await request(config.getDirektivBaseUrl()).get(
			`/api/v2/namespaces/${namespace}/secrets/foo-two`,
		)
		expect(res.statusCode).toEqual(200)
		expect(res.body.data).toEqual(expectDummySecret('foo-two'))
	})

	it(`should list foo1 and foo2`, async () => {
		const res = await request(config.getDirektivBaseUrl()).get(
			`/api/v2/namespaces/${namespace}/secrets`,
		)
		expect(res.statusCode).toEqual(200)
		expect(res.body).toEqual({
			data: [expectDummySecret('foo-one'), expectDummySecret('foo-two')],
		})
	})

	it(`should delete foo-one`, async () => {
		const res = await request(config.getDirektivBaseUrl()).delete(
			`/api/v2/namespaces/${namespace}/secrets/foo-one`,
		)
		expect(res.statusCode).toEqual(200)
	})

	it(`should list foo1 and foo-two`, async () => {
		const res = await request(config.getDirektivBaseUrl()).get(
			`/api/v2/namespaces/${namespace}/secrets`,
		)
		expect(res.statusCode).toEqual(200)
		expect(res.body).toEqual({
			data: [expectDummySecret('foo-two')],
		})
	})
})

function makeDummySecret(name) {
	return {
		name,
		data: btoa('value of' + name),
	}
}

function expectDummySecret(name) {
	return {
		name,
		initialized: true,
		createdAt: expect.stringMatching(regex.timestampRegex),
		updatedAt: expect.stringMatching(regex.timestampRegex),
	}
}
