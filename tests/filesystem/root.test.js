import { beforeAll, describe, expect, it } from '@jest/globals'

import common from '../common'
import request from '../common/request'

const namespaceName = 'root'

describe('Test behaviour specific to the root node', () => {
	beforeAll(common.helpers.deleteAllNamespaces)

	common.helpers.itShouldCreateNamespace(it, expect, namespaceName)

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
				results: [],
			},
		})
	})

	it(`should fail to manually create a root directory`, async () => {
		const createRootDirResponse = await request(common.config.getDirektivHost()).put(`/api/namespaces/${ namespaceName }/tree/`)
		expect(createRootDirResponse.statusCode).toEqual(405)
		expect(createRootDirResponse.body).toEqual({})
	})

	it(`should fail to delete a root directory`, async () => {
		const deleteRootDirResponse = await request(common.config.getDirektivHost()).delete(`/api/namespaces/${ namespaceName }/tree/`)
		expect(deleteRootDirResponse.statusCode).toEqual(405)
		expect(deleteRootDirResponse.body).toEqual({})
	})

	// TODO: test fail to rename root node
})
