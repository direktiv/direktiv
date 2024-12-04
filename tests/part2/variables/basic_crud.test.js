import { beforeAll, describe, expect, it } from '@jest/globals'
import { basename } from 'path'

import config from '../../common/config'
import helpers from '../../common/helpers'
import request from '../../common/request'

const namespace = basename(__filename)

describe('Test variable get delete list calls', () => {
	beforeAll(helpers.deleteAllNamespaces)
	helpers.itShouldCreateNamespace(it, expect, namespace)

	let foo1Res
	it(`should create a new variable foo1`, async () => {
		foo1Res = await request(config.getDirektivHost())
			.post(`/api/v2/namespaces/${ namespace }/variables`)
			.send(makeDummyVar('foo1'))
		expect(foo1Res.statusCode).toEqual(200)
	})

	let foo2Res
	it(`should create a new variable foo2`, async () => {
		foo2Res = await request(config.getDirektivHost())
			.post(`/api/v2/namespaces/${ namespace }/variables`)
			.send(makeDummyVar('foo2'))
		expect(foo2Res.statusCode).toEqual(200)
	})

	it(`should get the new variable foo1`, async () => {
		const res = await request(config.getDirektivHost())
			.get(`/api/v2/namespaces/${ namespace }/variables/${ foo1Res.body.data.id }`)
		expect(res.statusCode).toEqual(200)

		expect(res.body.data).toMatchObject(expectDummyVar('foo1'))
	})

	it(`should raw get the new variable foo1`, async () => {
		const res = await request(config.getDirektivHost())
			.get(`/api/v2/namespaces/${ namespace }/variables/${ foo1Res.body.data.id }?raw=true`)
		expect(res.statusCode).toEqual(200)
		expect(res.headers['content-type']).toEqual('mime_foo1')
		expect(res.headers['content-length']).toEqual('13')
		expect(res.text).toEqual('value of foo1')
	})

	it(`should raw list the new variable foo1`, async () => {
		const res = await request(config.getDirektivHost())
			.get(`/api/v2/namespaces/${ namespace }/variables/?name=foo1&raw=true`)
		expect(res.statusCode).toEqual(200)
		expect(res.headers['content-type']).toEqual('mime_foo1')
		expect(res.headers['content-length']).toEqual('13')
		expect(res.text).toEqual('value of foo1')
	})

	it(`should raw get 404 for none existed var`, async () => {
		const res = await request(config.getDirektivHost())
			.get(`/api/v2/namespaces/${ namespace }/variables/daa9e62b-3e20-4853-b415-d97851f200aa?raw=true`)
		expect(res.statusCode).toEqual(404)
	})

	it(`should raw list 404 for none existed var`, async () => {
		const res = await request(config.getDirektivHost())
			.get(`/api/v2/namespaces/${ namespace }/variables/?name=something&raw=true`)
		expect(res.statusCode).toEqual(404)
	})

	it(`should get the new variable foo2`, async () => {
		const res = await request(config.getDirektivHost())
			.get(`/api/v2/namespaces/${ namespace }/variables/${ foo2Res.body.data.id }`)
		expect(res.statusCode).toEqual(200)
		expect(res.body.data).toMatchObject(expectDummyVar('foo2'))
	})

	it(`should list foo1 and foo2`, async () => {
		const res = await request(config.getDirektivHost())
			.get(`/api/v2/namespaces/${ namespace }/variables`)
		expect(res.statusCode).toEqual(200)
		expect(res.body).toMatchObject({
			data: [ expectDummyVar('foo1'), expectDummyVar('foo2') ],
		})
	})

	it(`should delete foo1`, async () => {
		const res = await request(config.getDirektivHost())
			.delete(`/api/v2/namespaces/${ namespace }/variables/${ foo1Res.body.data.id }`)
		expect(res.statusCode).toEqual(200)
	})

	it(`should list foo1 and foo2`, async () => {
		const res = await request(config.getDirektivHost())
			.get(`/api/v2/namespaces/${ namespace }/variables`)
		expect(res.statusCode).toEqual(200)
		expect(res.body).toMatchObject({
			data: [ expectDummyVar('foo2') ],
		})
	})
})

function makeDummyVar (name) {
	return {
		name,
		data: btoa('value of ' + name),
		mimeType: 'mime_' + name,
	}
}

function expectDummyVar (name) {
	return {
		name,
	}
}
