import config from './config'
import common from './index'
import regex from './regex'
import request from './request'

async function deleteAllNamespaces () {
	const listResponse = await request(config.getDirektivHost()).get(`/api/v2/namespaces`)
	if (listResponse.statusCode !== 200)
		throw Error(`none ok namespaces list statusCode(${ listResponse.statusCode })`)

	for (const namespace of listResponse.body.data) {
		const response = await request(config.getDirektivHost()).delete(`/api/namespaces/${ namespace.name }`)

		if (response.statusCode !== 200)
			throw Error(`none ok namespace(${ namespace.name }) delete statusCode(${ response.statusCode })`)
	}
}

async function itShouldCreateNamespace (it, expect, ns) {
	it(`should create a new namespace ${ ns }`, async () => {
		const res = await request(common.config.getDirektivHost()).put(`/api/namespaces/${ ns }`)
		expect(res.statusCode).toEqual(200)
		expect(res.body).toMatchObject({
			namespace: {
				name: ns,
				// regex /^2.*Z$/ matches timestamps like 2023-03-01T14:19:52.383871512Z
				createdAt: expect.stringMatching(/^2.*Z$/),
				updatedAt: expect.stringMatching(/^2.*Z$/),
			},
		})
	})
}

async function itShouldCreateFileV2 (it, expect, ns, path, name, type, mimeType, data) {
	it(`should create a new file ${ path }`, async () => {
		if (path === '/')
			path = ''

		const res = await request(common.config.getDirektivHost())
			.post(`/api/v2/namespaces/${ ns }/files${ path }`)
			.set('Content-Type', 'application/json')
			.send({
				name,
				type,
				mimeType,
				data,
			})
		expect(res.statusCode).toEqual(200)

		expect(res.body.data).toEqual({
			path: `${ path }/${ name }`,
			type,
			data,
			mimeType,
			size: Buffer.from(data, 'base64').length,
			createdAt: expect.stringMatching(regex.timestampRegex),
			updatedAt: expect.stringMatching(regex.timestampRegex),
		})
	})
}

function itShouldCreateYamlFileV2 (it, expect, ns, path, name, type, data) {
	return itShouldCreateFileV2(it, expect, ns, path, name, type, 'application/yaml', btoa(data))
}

async function itShouldCreateDirV2 (it, expect, ns, path, name) {
	it(`should create a new dir ${ path }`, async () => {
		const res = await request(common.config.getDirektivHost())
			.post(`/api/v2/namespaces/${ ns }/files${ path }`)
			.set('Content-Type', 'application/json')
			.send({
				name,
				type: 'directory',
			})
		expect(res.statusCode).toEqual(200)
		if (path === '/')
			path = ''

		expect(res.body.data).toEqual({
			path: `${ path }/${ name }`,
			type: 'directory',
			createdAt: expect.stringMatching(regex.timestampRegex),
			updatedAt: expect.stringMatching(regex.timestampRegex),
		})
	})
}

function itShouldUpdateFilePathV2 (it, expect, ns, path, newPath) {
	return itShouldUpdateFileV2(it, expect, ns, path, { path: newPath })
}

async function itShouldUpdateFileV2 (it, expect, ns, path, newPatch) {
	let title = `should update file path ${ path }`
	if (newPatch.path !== undefined)
		title = `should update file path ${ path } to ${ newPatch.path }`

	it(title, async () => {
		const res = await request(common.config.getDirektivHost())
			.patch(`/api/v2/namespaces/${ ns }/files${ path }`)
			.set('Content-Type', 'application/json')
			.send(newPatch)
		expect(res.statusCode).toEqual(200)

		const want = {
			createdAt: expect.stringMatching(regex.timestampRegex),
			updatedAt: expect.stringMatching(regex.timestampRegex),
		}
		if (newPatch.path !== undefined)
			want.path = newPatch.path

		if (newPatch.data !== undefined)
			want.data = newPatch.data

		expect(res.body.data).toMatchObject(want)
	})
}

async function itShouldCheckPathExistsV2 (it, expect, ns, path, assertExits) {
	it(`should check if path(${ path }) exists(${ assertExits })`, async () => {
		const res = await request(common.config.getDirektivHost())
			.get(`/api/v2/namespaces/${ ns }/files${ path }`)

		if (assertExits)
			expect(res.statusCode).toEqual(200)
		else
			expect(res.statusCode).toEqual(404)
	})
}

function dummyWorkflow (someText) {
	return `
direktiv_api: workflow/v1
description: A simple 'no-op' state that returns ${ someText } 'Hello world!'
states:
- id: helloworld
  type: noop
`
}

function itShouldUpdateYamlFileV2 (it, expect, ns, path, data) {
	return itShouldUpdateFileV2(it, expect, ns, path, { data: btoa(data) })
}

async function itShouldDeleteFileV2 (it, expect, ns, path) {
	it(`should delete a file ${ path }`, async () => {
		const res = await request(common.config.getDirektivHost())
			.delete(`/api/v2/namespaces/${ ns }/files${ path }`)

		expect(res.statusCode).toEqual(200)
		expect(res.body).toEqual('')
	})
}

async function itShouldCreateVariableV2 (it, expect, ns, variable) {
	it(`should create a variable ${ variable.name }`, async () => {
		const createRes = await request(config.getDirektivHost())
			.post(`/api/v2/namespaces/${ ns }/variables`)
			.send(variable)
		expect(createRes.statusCode).toEqual(200)
	})
}

function sleep (ms) {
	return new Promise(resolve => setTimeout(resolve, ms))
}

export default {
	deleteAllNamespaces,
	itShouldCreateNamespace,

	itShouldUpdateYamlFileV2,
	itShouldDeleteFileV2,
	dummyWorkflow,
	itShouldCreateYamlFileV2,
	itShouldCreateDirV2,
	itShouldCreateFileV2,
	itShouldCheckPathExistsV2,
	itShouldUpdateFilePathV2,
	itShouldUpdateFileV2,
	itShouldCreateVariableV2,
	sleep,
}
