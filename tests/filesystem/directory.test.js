import { beforeAll, describe, expect, it } from '@jest/globals'
import request from 'supertest'

import common from '../common'

const createDirResponse = {
	namespace: expect.anything(),
	node: common.structs.nodeObject,
}

// const readDirResponse = {
// 	namespace: expect.anything(),
// 	node: common.structs.nodeObject,
// 	children: {
// 		pageInfo: common.structs.pageInfoObject,
// 		results: expect.anything(),
// 	},
// }

// TODO: test fail to rename a node into itself
// TODO: test fail to rename a node out of itself
// TODO: test fail to rename a node into a non-existent place

const namespaceName = 'a'
const subdirName = 'b'

const expectedChildNodeObject = {
	createdAt: expect.stringMatching(common.regex.timestampRegex),
	updatedAt: expect.stringMatching(common.regex.timestampRegex),
	name: subdirName,
	path: `/${ subdirName }`,
	parent: `/`,
	type: common.filesystem.nodeTypeDirectory,
	expandedType: common.filesystem.extendedNodeTypeDirectory,
	attributes: expect.anything(),
	readOnly: false,
	mimeType: expect.anything(),
}

describe('Test basic directory operations', () => {
	beforeAll(common.helpers.deleteAllNamespaces)

	it(`should fail to create a non-root directory because of a missing namespace`, async () => {
		const createDirectoryResponse = await request(common.config.getDirektivHost()).put(`/api/namespaces/${ namespaceName }/tree/${ subdirName }?op=create-directory`)

		expect(createDirectoryResponse.statusCode).toEqual(404)
		expect(createDirectoryResponse.body).toMatchObject({
			code: 404,
			message: `ErrNotFound`,
		})
	})

	common.helpers.itShouldCreateNamespace(it, expect, namespaceName)

	it(`should fail to create a non-root directory because of a missing/invalid 'op' param`, async () => {
		const createDirectoryResponse1 = await request(common.config.getDirektivHost()).put(`/api/namespaces/${ namespaceName }/tree/${ subdirName }`)
		const createDirectoryResponse2 = await request(common.config.getDirektivHost()).put(`/api/namespaces/${ namespaceName }/tree/${ subdirName }?op=delete-directory`)
		const createDirectoryResponse3 = await request(common.config.getDirektivHost()).put(`/api/namespaces/${ namespaceName }/tree/${ subdirName }?opa=create-directory`)

		expect(createDirectoryResponse1.statusCode).toEqual(405)
		expect(createDirectoryResponse2.statusCode).toEqual(405)
		expect(createDirectoryResponse3.statusCode).toEqual(405)

		expect(createDirectoryResponse1.body).toEqual({}) // TODO: revisit
	})

	it(`should fail to create a non-root directory because of a bad method`, async () => {
		const createDirectoryResponse1 = await request(common.config.getDirektivHost()).post(`/api/namespaces/${ namespaceName }/tree/${ subdirName }?op=create-directory`)
		const createDirectoryResponse2 = await request(common.config.getDirektivHost()).get(`/api/namespaces/${ namespaceName }/tree/${ subdirName }?op=create-directory`)
		const createDirectoryResponse3 = await request(common.config.getDirektivHost()).patch(`/api/namespaces/${ namespaceName }/tree/${ subdirName }?op=create-directory`)
		const createDirectoryResponse4 = await request(common.config.getDirektivHost()).delete(`/api/namespaces/${ namespaceName }/tree/${ subdirName }?op=create-directory`)

		expect(createDirectoryResponse1.statusCode).toEqual(405)
		expect(createDirectoryResponse2.statusCode).toEqual(404)
		expect(createDirectoryResponse3.statusCode).toEqual(405)
		expect(createDirectoryResponse4.statusCode).toEqual(405)

		expect(createDirectoryResponse1.body).toEqual({}) // TODO: revisit
		expect(createDirectoryResponse2.body).toMatchObject({
			code: 404,
			message: `file '/b': not found`,
		})
	})

	it(`should create a sub-directory`, async () => {
		const createDirectoryResponse = await request(common.config.getDirektivHost()).put(`/api/namespaces/${ namespaceName }/tree/${ subdirName }?op=create-directory`)
		expect(createDirectoryResponse.statusCode).toEqual(200)
		expect(createDirectoryResponse.body).toEqual(createDirResponse)
		expect(createDirectoryResponse.body.namespace).toEqual(namespaceName)
		expect(createDirectoryResponse.body.node).toEqual(expectedChildNodeObject)
	})

	it(`should read the root directory`, async () => {
		const readRootDirResponse = await request(common.config.getDirektivHost()).get(`/api/namespaces/${ namespaceName }/tree/`)
		expect(readRootDirResponse.statusCode).toEqual(200)
		expect(readRootDirResponse.body).toMatchObject({
			namespace: namespaceName,
			node: {
				name: '',
				path: '/',
				parent: '/',
				type: common.filesystem.nodeTypeDirectory,
				attributes: [],
				readOnly: false,
				expandedType: common.filesystem.extendedNodeTypeDirectory,
			},
			children: {
				pageInfo: {
					limit: 0,
					offset: 0,
					total: 1,
					order: [],
					filter: [],
				},
				results: expect.arrayContaining([ expectedChildNodeObject ]),
			},
		})
	})

	it(`should read the sub-directory`, async () => {
		const readNonRootDirResponse = await request(common.config.getDirektivHost()).get(`/api/namespaces/${ namespaceName }/tree/${ subdirName }`)
		expect(readNonRootDirResponse.statusCode).toEqual(200)
		expect(readNonRootDirResponse.body).toMatchObject({
			namespace: namespaceName,
			node: expectedChildNodeObject,
			children: {
				pageInfo: {
					limit: 0,
					offset: 0,
					total: 0,
					order: [],
					filter: [],
				},
				results: [],
			},
		})
	})

	it(`should fail to delete the empty sub-directory due to a missing op param`, async () => {
		const deleteDirectoryResponse = await request(common.config.getDirektivHost()).delete(`/api/namespaces/${ namespaceName }/tree/${ subdirName }`)
		expect(deleteDirectoryResponse.statusCode).toEqual(405)
		expect(deleteDirectoryResponse.body).toMatchObject({})
	})

	it(`should fail to delete the empty sub-directory due to a bad method`, async () => {
		let deleteDirectoryResponse = await request(common.config.getDirektivHost()).post(`/api/namespaces/${ namespaceName }/tree/${ subdirName }?op=delete-node`)
		expect(deleteDirectoryResponse.statusCode).toEqual(405)
		expect(deleteDirectoryResponse.body).toMatchObject({})

		deleteDirectoryResponse = await request(common.config.getDirektivHost()).put(`/api/namespaces/${ namespaceName }/tree/${ subdirName }?op=delete-node`)
		expect(deleteDirectoryResponse.statusCode).toEqual(405)
		expect(deleteDirectoryResponse.body).toMatchObject({})
	})

	it(`should delete the empty sub-directory`, async () => {
		const deleteDirectoryResponse = await request(common.config.getDirektivHost()).delete(`/api/namespaces/${ namespaceName }/tree/${ subdirName }?op=delete-node`)
		expect(deleteDirectoryResponse.statusCode).toEqual(200)
		expect(deleteDirectoryResponse.body).toMatchObject({})
	})

	it(`should fail to delete a non-existant sub-directory`, async () => {
		const deleteDirectoryResponse = await request(common.config.getDirektivHost()).delete(`/api/namespaces/${ namespaceName }/tree/${ subdirName }?op=delete-node`)
		expect(deleteDirectoryResponse.statusCode).toEqual(404)
		expect(deleteDirectoryResponse.body).toMatchObject({})
	})

	it(`should create a sub-directory with a trailing slash in its path`, async () => {
		const createDirectoryResponse = await request(common.config.getDirektivHost()).put(`/api/namespaces/${ namespaceName }/tree/${ subdirName }/?op=create-directory`)
		expect(createDirectoryResponse.statusCode).toEqual(200)
		expect(createDirectoryResponse.body).toEqual(createDirResponse)
		expect(createDirectoryResponse.body.namespace).toEqual(namespaceName)
		expect(createDirectoryResponse.body.node).toEqual(expectedChildNodeObject)
	})

	// TODO: test node name regex compliance
	// TODO: test everything with/without trailing slash
	// TODO: test delete
	// TODO: test pagination
	// TODO: test filtering
	// TODO: test ordering
	// TODO: test logs
})