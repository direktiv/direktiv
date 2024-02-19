import request from "./request"

import config from './config'
import common from './index'
import regex from './regex'

async function deleteAllNamespaces () {
	const listResponse = await request(config.getDirektivHost()).get(`/api/namespaces`)
	if (listResponse.statusCode !== 200)
		throw Error(`none ok namespaces list statusCode(${ listResponse.statusCode })`)


	for (const namespace of listResponse.body.results) {
		const response = await request(config.getDirektivHost()).delete(`/api/namespaces/${ namespace.name }?recursive=true`)

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

async function itShouldCreateFile (it, expect, ns, path, data) {
	it(`should create a new file ${ path }`, async () => {
		const res = await request(common.config.getDirektivHost())
			.put(`/api/namespaces/${ ns }/tree${ path }?op=create-workflow`)
			.set({
				'Content-Type': 'text/plain',
			})

			.send(data)

		expect(res.statusCode).toEqual(200)
		expect(res.body).toMatchObject({
			namespace: ns,
		})
	})
}

async function itShouldCreateFileV2 (it, expect, ns, path, name, type, mimeType, data) {
	it(`should create a new file ${ path }`, async () => {
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
		if (path === '/')
			path = ''

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

async function itShouldUpdatePathV2 (it, expect, ns, path, newPath) {
	it(`should update file path ${ path } to ${ newPath }`, async () => {
		const res = await request(common.config.getDirektivHost())
			.patch(`/api/v2/namespaces/${ ns }/files${ path }`)
			.set('Content-Type', 'application/json')
			.send({
				path: newPath,
			})
		expect(res.statusCode).toEqual(200)
		expect(res.body.data).toMatchObject({
			path: newPath,
			type:expect.stringMatching("directory|file|workflow|service|endpoint|consumer"),
			createdAt: expect.stringMatching(regex.timestampRegex),
			updatedAt: expect.stringMatching(regex.timestampRegex),
		})
	})
}

async function itShouldUpdateFileV2 (it, expect, ns, path, newPatch) {
	it(`should update file path ${ path }`, async () => {
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


async function itShouldCreateDirectory (it, expect, ns, path) {
	it(`should create a directory ${ path }`, async () => {
		const res = await request(common.config.getDirektivHost())
			.put(`/api/namespaces/${ ns }/tree${ path }?op=create-directory`)

		expect(res.statusCode).toEqual(200)
		expect(res.body).toMatchObject({
			namespace: ns,
		})
	})
}

async function itShouldUpdateFile (it, expect, ns, path, data) {
	it(`should update existing file ${ path }`, async () => {
		const res = await request(common.config.getDirektivHost())
			.post(`/api/namespaces/${ ns }/tree${ path }?op=update-workflow`)
			.set({
				'Content-Type': 'text/plain',
			})

			.send(data)

		expect(res.statusCode).toEqual(200)
		expect(res.body).toMatchObject({
			namespace: ns,
		})
	})
}

async function itShouldDeleteFile (it, expect, ns, path) {
	it(`should delete a file ${ path }`, async () => {
		const res = await request(common.config.getDirektivHost())
			.delete(`/api/namespaces/${ ns }/tree${ path }?op=delete-node`)

		expect(res.statusCode).toEqual(200)
		expect(res.body).toMatchObject({})
	})
}

async function itShouldRenameFile (it, expect, ns, path, newPath) {
	it(`should delete a file ${ path }`, async () => {
		const res = await request(common.config.getDirektivHost())
			.post(`/api/namespaces/${ ns }/tree${ path }?op=rename-node`)
			.send({ new: newPath })

		expect(res.statusCode).toEqual(200)
		expect(res.body).toMatchObject({})
	})
}




export default {
	deleteAllNamespaces,
	itShouldCreateNamespace,
	itShouldCreateFile,
	itShouldDeleteFile,
	itShouldRenameFile,
	itShouldUpdateFile,
	itShouldCreateDirectory,
	dummyWorkflow,
	itShouldCreateDirV2,
	itShouldCreateFileV2,
	itShouldCheckPathExistsV2,
	itShouldUpdatePathV2,
	itShouldUpdateFileV2,
}