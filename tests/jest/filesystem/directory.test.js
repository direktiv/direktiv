import request from 'supertest'

import common from "../common"

const createDirResponse = {
    namespace: expect.anything(),
    node: common.structs.nodeObject,
}

const readDirResponse = {
    namespace: expect.anything(),
    node: common.structs.nodeObject,
    children: {
        pageInfo: common.structs.pageInfoObject,
         results: expect.anything(),
    },
}

// TODO: test fail to rename a node into itself
// TODO: test fail to rename a node out of itself
// TODO: test fail to rename a node into a non-existent place

describe('Test basic directory operations', () => {
    it(`should create a namespace and create a non-root directory`, async () => {
        const namespaceName = "a"
        const directoryName = "b"
        await request(common.config.getDirektivHost()).delete(`/api/namespaces/${namespaceName}?recursive=true`)
        var createNamespaceResponse = await request(common.config.getDirektivHost()).put(`/api/namespaces/${namespaceName}`)
        var createDirectoryResponse = await request(common.config.getDirektivHost()).put(`/api/namespaces/${namespaceName}/tree/${directoryName}?op=create-directory`)
        var readRootDirResponse = await request(common.config.getDirektivHost()).get(`/api/namespaces/${namespaceName}/tree/`)
        var readNonRootDirResponse = await request(common.config.getDirektivHost()).get(`/api/namespaces/${namespaceName}/tree/${directoryName}`)
        await request(common.config.getDirektivHost()).delete(`/api/namespaces/${namespaceName}?recursive=true`)
    
        expect(createNamespaceResponse.statusCode).toEqual(200)

        var expectedChildNodeObject = {
            createdAt: expect.stringMatching(common.regex.timestampRegex),
            updatedAt: expect.stringMatching(common.regex.timestampRegex),
            name: directoryName,
            path: `/${directoryName}`,
            parent: `/`,
            type: common.filesystem.nodeTypeDirectory,
            expandedType: common.filesystem.extendedNodeTypeDirectory,
            attributes: expect.anything(),
            oid: '', // TODO: revisit
            readOnly: false,
        }

        expect(createDirectoryResponse.statusCode).toEqual(200)
        expect(createDirectoryResponse.body).toEqual(createDirResponse)
        expect(createDirectoryResponse.body.namespace).toEqual(namespaceName)
        expect(createDirectoryResponse.body.node).toEqual(expectedChildNodeObject)

        expect(readRootDirResponse.statusCode).toEqual(200)
        expect(readRootDirResponse.body).toMatchObject(readDirResponse)
        expect(readRootDirResponse.body.namespace).toEqual(namespaceName)
        expect(readRootDirResponse.body.node.name).toEqual(``)
        expect(readRootDirResponse.body.node.path).toEqual(`/`)
        expect(readRootDirResponse.body.node.parent).toEqual(`/`)
        expect(readRootDirResponse.body.node.type).toEqual(common.filesystem.nodeTypeDirectory)
        expect(readRootDirResponse.body.node.attributes.length).toEqual(0)
        expect(readRootDirResponse.body.node.oid).toEqual(``) // TODO: revisit
        expect(readRootDirResponse.body.node.readOnly).toEqual(false)
        expect(readRootDirResponse.body.node.expandedType).toEqual(common.filesystem.extendedNodeTypeDirectory)
        expect(readRootDirResponse.body.children.results.length).toEqual(1)
        expect(readRootDirResponse.body.children.pageInfo.limit).toEqual(0)
        expect(readRootDirResponse.body.children.pageInfo.offset).toEqual(0)
        expect(readRootDirResponse.body.children.pageInfo.total).toEqual(1)
        expect(readRootDirResponse.body.children.pageInfo.order.length).toEqual(0)
        expect(readRootDirResponse.body.children.pageInfo.filter.length).toEqual(0)
        expect(readRootDirResponse.body.children.results[0]).toEqual(expectedChildNodeObject)

        expect(readNonRootDirResponse.statusCode).toEqual(200)
        expect(readNonRootDirResponse.body).toMatchObject(readDirResponse)
        expect(readNonRootDirResponse.body.namespace).toEqual(namespaceName)
        expect(readNonRootDirResponse.body.node).toEqual(expectedChildNodeObject)
        expect(readNonRootDirResponse.body.children.results.length).toEqual(0)
        expect(readNonRootDirResponse.body.children.pageInfo.limit).toEqual(0)
        expect(readNonRootDirResponse.body.children.pageInfo.offset).toEqual(0)
        expect(readNonRootDirResponse.body.children.pageInfo.total).toEqual(0)
        expect(readNonRootDirResponse.body.children.pageInfo.order.length).toEqual(0)
        expect(readNonRootDirResponse.body.children.pageInfo.filter.length).toEqual(0)
    })

    it(`should create a namespace and fail to create a non-root directory because of a missing/invalid 'op' param`, async () => {
        const namespaceName = "a"
        const directoryName = "b"
        await request(common.config.getDirektivHost()).delete(`/api/namespaces/${namespaceName}?recursive=true`)
        var createNamespaceResponse = await request(common.config.getDirektivHost()).put(`/api/namespaces/${namespaceName}`)
        var createDirectoryResponse1 = await request(common.config.getDirektivHost()).put(`/api/namespaces/${namespaceName}/tree/${directoryName}`)
        var createDirectoryResponse2 = await request(common.config.getDirektivHost()).put(`/api/namespaces/${namespaceName}/tree/${directoryName}?op=delete-directory`)
        var createDirectoryResponse3 = await request(common.config.getDirektivHost()).put(`/api/namespaces/${namespaceName}/tree/${directoryName}?opa=create-directory`)
        await request(common.config.getDirektivHost()).delete(`/api/namespaces/${namespaceName}?recursive=true`)
    
        expect(createNamespaceResponse.statusCode).toEqual(200)

        expect(createDirectoryResponse1.statusCode).toEqual(405)
        expect(createDirectoryResponse2.statusCode).toEqual(405)
        expect(createDirectoryResponse3.statusCode).toEqual(405)

        expect(createDirectoryResponse1.body).toEqual({}) // TODO: revisit
    })

    it(`should create a namespace and fail to create a non-root directory because of a bad method`, async () => {
        const namespaceName = "a"
        const directoryName = "b"
        await request(common.config.getDirektivHost()).delete(`/api/namespaces/${namespaceName}?recursive=true`)
        var createNamespaceResponse = await request(common.config.getDirektivHost()).put(`/api/namespaces/${namespaceName}`)
        var createDirectoryResponse1 = await request(common.config.getDirektivHost()).post(`/api/namespaces/${namespaceName}/tree/${directoryName}?op=create-directory`)
        var createDirectoryResponse2 = await request(common.config.getDirektivHost()).get(`/api/namespaces/${namespaceName}/tree/${directoryName}?op=create-directory`)
        var createDirectoryResponse3 = await request(common.config.getDirektivHost()).patch(`/api/namespaces/${namespaceName}/tree/${directoryName}?op=create-directory`)
        var createDirectoryResponse4 = await request(common.config.getDirektivHost()).delete(`/api/namespaces/${namespaceName}/tree/${directoryName}?op=create-directory`)
        await request(common.config.getDirektivHost()).delete(`/api/namespaces/${namespaceName}?recursive=true`)
    
        expect(createNamespaceResponse.statusCode).toEqual(200)

        expect(createDirectoryResponse1.statusCode).toEqual(405)
        expect(createDirectoryResponse2.statusCode).toEqual(404)
        expect(createDirectoryResponse3.statusCode).toEqual(405)
        expect(createDirectoryResponse4.statusCode).toEqual(405)

        expect(createDirectoryResponse1.body).toEqual({}) // TODO: revisit
        expect(createDirectoryResponse2.body.code).toEqual(404)
        expect(createDirectoryResponse2.body.message).toEqual(`file does not exist`)
    })

    it(`should fail to create a non-root directory because of a missing namespace`, async () => {
        const namespaceName = "a"
        const directoryName = "b"
        await request(common.config.getDirektivHost()).delete(`/api/namespaces/${namespaceName}?recursive=true`)
        var createDirectoryResponse = await request(common.config.getDirektivHost()).put(`/api/namespaces/${namespaceName}/tree/${directoryName}?op=create-directory`)
        await request(common.config.getDirektivHost()).delete(`/api/namespaces/${namespaceName}?recursive=true`)
    
        expect(createDirectoryResponse.statusCode).toEqual(404)
        expect(createDirectoryResponse.body.code).toEqual(404)
        expect(createDirectoryResponse.body.message).toEqual(`namespace not found`)
    })

    it(`should create a namespace and create a non-root directory with a trailing slash`, async () => {
        const namespaceName = "a"
        const directoryName = "b"
        await request(common.config.getDirektivHost()).delete(`/api/namespaces/${namespaceName}?recursive=true`)
        var createNamespaceResponse = await request(common.config.getDirektivHost()).put(`/api/namespaces/${namespaceName}`)
        var createDirectoryResponse = await request(common.config.getDirektivHost()).put(`/api/namespaces/${namespaceName}/tree/${directoryName}/?op=create-directory`)
        await request(common.config.getDirektivHost()).delete(`/api/namespaces/${namespaceName}?recursive=true`)
    
        expect(createNamespaceResponse.statusCode).toEqual(200)

        var expectedChildNodeObject = {
            createdAt: expect.stringMatching(common.regex.timestampRegex),
            updatedAt: expect.stringMatching(common.regex.timestampRegex),
            name: directoryName,
            path: `/${directoryName}`,
            parent: `/`,
            type: common.filesystem.nodeTypeDirectory,
            expandedType: common.filesystem.extendedNodeTypeDirectory,
            attributes: expect.anything(),
            oid: '', // TODO: revisit
            readOnly: false,
        }

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