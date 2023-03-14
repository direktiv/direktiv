import request from 'supertest'

import common from "../common"

const readDirResponse = {
    namespace: expect.anything(),
    node: common.structs.nodeObject,
    children: {
        pageInfo: common.structs.pageInfoObject,
         results: expect.anything(),
    },
}

describe('Test basic filesystem operations', () => {
    it(`should create a namespace and validate the automatically created root directory`, async () => {
        const namespaceName = "a"
        await request(common.config.getDirektivHost()).delete(`/api/namespaces/${namespaceName}?recursive=true`)
        var createNamespaceResponse = await request(common.config.getDirektivHost()).put(`/api/namespaces/${namespaceName}`)
        var readRootDirResponse = await request(common.config.getDirektivHost()).get(`/api/namespaces/${namespaceName}/tree/`)
        await request(common.config.getDirektivHost()).delete(`/api/namespaces/${namespaceName}?recursive=true`)
    
        expect(createNamespaceResponse.statusCode).toEqual(200)

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
        expect(readRootDirResponse.body.children.results.length).toEqual(0)
        expect(readRootDirResponse.body.children.pageInfo.limit).toEqual(0)
        expect(readRootDirResponse.body.children.pageInfo.offset).toEqual(0)
        expect(readRootDirResponse.body.children.pageInfo.total).toEqual(0)
        expect(readRootDirResponse.body.children.pageInfo.order.length).toEqual(0)
        expect(readRootDirResponse.body.children.pageInfo.filter.length).toEqual(0)
    })

    it(`should fail to manually create a root directory`, async () => {
        const namespaceName = "a"
        await request(common.config.getDirektivHost()).delete(`/api/namespaces/${namespaceName}?recursive=true`)
        var createNamespaceResponse = await request(common.config.getDirektivHost()).put(`/api/namespaces/${namespaceName}`)
        var createRootDirResponse = await request(common.config.getDirektivHost()).put(`/api/namespaces/${namespaceName}/tree/`)
        await request(common.config.getDirektivHost()).delete(`/api/namespaces/${namespaceName}?recursive=true`)
    
        expect(createNamespaceResponse.statusCode).toEqual(200)

        expect(createRootDirResponse.statusCode).toEqual(405)
        expect(createRootDirResponse.body).toEqual({}) // TODO: revisit
    })

    it(`should fail to delete a root directory`, async () => {
        const namespaceName = "a"
        await request(common.config.getDirektivHost()).delete(`/api/namespaces/${namespaceName}?recursive=true`)
        var createNamespaceResponse = await request(common.config.getDirektivHost()).put(`/api/namespaces/${namespaceName}`)
        var deleteRootDirResponse = await request(common.config.getDirektivHost()).delete(`/api/namespaces/${namespaceName}/tree/`)
        await request(common.config.getDirektivHost()).delete(`/api/namespaces/${namespaceName}?recursive=true`)
    
        expect(createNamespaceResponse.statusCode).toEqual(200)

        expect(deleteRootDirResponse.statusCode).toEqual(405)
        expect(deleteRootDirResponse.body).toEqual({}) // TODO: revisit
    })

    // TODO: test fail to rename root node 
})
