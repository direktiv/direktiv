import request from 'supertest'

import common from "../common"

const namespaceName = "a"
const workflowName = "b.yaml"
const simpleWorkflow = `
states:
- id: hello
  type: noop
  transform: 'jq({ msg: "Hello, world!" })'
`

const updatedSimpleWorkflow = `
states:
- id: hello_updated
  type: noop
  transform: 'jq({ msg: "Hello, world!" })'
`

var expectedChildNodeObject = {
    createdAt: expect.stringMatching(common.regex.timestampRegex),
    updatedAt: expect.stringMatching(common.regex.timestampRegex),
    name: workflowName,
    path: `/${workflowName}`,
    parent: `/`,
    type: common.filesystem.nodeTypeWorkflow,
    expandedType: common.filesystem.extendedNodeTypeWorkflow,
    attributes: expect.anything(),
    oid: '', // TODO: revisit
    readOnly: false,
}

describe('Test basic directory operations', () => {
    beforeAll(common.helpers.deleteAllNamespaces)
    afterAll(common.helpers.deleteAllNamespaces)

    it(`should fail to create a workflow because of a missing namespace`, async () => {
        var createWorkflowResponse = await request(common.config.getDirektivHost()).put(`/api/namespaces/${namespaceName}/tree/${workflowName}?op=create-workflow`)

        expect(createWorkflowResponse.statusCode).toEqual(404)
        expect(createWorkflowResponse.body).toMatchObject({
            code: 404,
            message: `namespace not found`,
        })
    })

    it(`should create a namespace`, async () => {
        var createNamespaceResponse = await request(common.config.getDirektivHost()).put(`/api/namespaces/${namespaceName}`)
        expect(createNamespaceResponse.statusCode).toEqual(200)
    })

    it(`should fail to create a workflow because of a missing/invalid 'op' param`, async () => {
        var createWorkflowResponse1 = await request(common.config.getDirektivHost()).put(`/api/namespaces/${namespaceName}/tree/${workflowName}`)
        var createWorkflowResponse2 = await request(common.config.getDirektivHost()).put(`/api/namespaces/${namespaceName}/tree/${workflowName}?op=delete-directory`)
        var createWorkflowResponse3 = await request(common.config.getDirektivHost()).put(`/api/namespaces/${namespaceName}/tree/${workflowName}?opa=create-workflow`)

        expect(createWorkflowResponse1.statusCode).toEqual(405)
        expect(createWorkflowResponse2.statusCode).toEqual(405)
        expect(createWorkflowResponse3.statusCode).toEqual(405)

        expect(createWorkflowResponse1.body).toEqual({}) // TODO: revisit
    })

    it(`should fail to create a workflow because of a bad method`, async () => {
        var createWorkflowResponse1 = await request(common.config.getDirektivHost()).post(`/api/namespaces/${namespaceName}/tree/${workflowName}?op=create-workflow`)
        var createWorkflowResponse2 = await request(common.config.getDirektivHost()).get(`/api/namespaces/${namespaceName}/tree/${workflowName}?op=create-workflow`)
        var createWorkflowResponse3 = await request(common.config.getDirektivHost()).patch(`/api/namespaces/${namespaceName}/tree/${workflowName}?op=create-workflow`)
        var createWorkflowResponse4 = await request(common.config.getDirektivHost()).delete(`/api/namespaces/${namespaceName}/tree/${workflowName}?op=create-workflow`)

        expect(createWorkflowResponse1.statusCode).toEqual(405)
        expect(createWorkflowResponse2.statusCode).toEqual(404)
        expect(createWorkflowResponse3.statusCode).toEqual(405)
        expect(createWorkflowResponse4.statusCode).toEqual(405)

        expect(createWorkflowResponse1.body).toEqual({}) // TODO: revisit
        expect(createWorkflowResponse2.body).toMatchObject({
            code: 404,
            message: `file '/b.yaml': not found`,
        })
    })

    it(`should fail to create a workflow at the root of the filesystem`, async () => {
        var createWorkflowResponse = await request(common.config.getDirektivHost()).put(`/api/namespaces/${namespaceName}/tree/?op=create-workflow`)

        expect(createWorkflowResponse.statusCode).toEqual(406)
    })

    it(`should fail to create an invalid workflow`, async () => {
        var createWorkflowResponse = await request(common.config.getDirektivHost()).put(`/api/namespaces/${namespaceName}/tree/${workflowName}?op=create-workflow`)

        expect(createWorkflowResponse.statusCode).toEqual(406)
        expect(createWorkflowResponse.body).toMatchObject({
            code: 406,
            message: `workflow has no defined states`,
        })
    })

    it(`should create a workflow`, async () => {
        var createWorkflowResponse = await request(common.config.getDirektivHost()).put(`/api/namespaces/${namespaceName}/tree/${workflowName}?op=create-workflow`).send(simpleWorkflow)

        expect(createWorkflowResponse.statusCode).toEqual(200)
        expect(createWorkflowResponse.body).toMatchObject({
            namespace: namespaceName,
            node: expectedChildNodeObject,
            revision: {
                createdAt: expect.stringMatching(common.regex.timestampRegex),
                hash: expect.stringMatching(common.regex.hashRegex),
                source: expect.stringMatching(common.regex.base64Regex),
                name: expect.stringMatching(common.regex.uuidRegex),
            },
        })

        var buf = Buffer.from(createWorkflowResponse.body.revision.source, 'base64')
        expect(buf.toString()).toEqual(simpleWorkflow)
    })

    it(`should update a workflow`, async () => {
        var createWorkflowResponse = await request(common.config.getDirektivHost()).post(`/api/namespaces/${namespaceName}/tree/${workflowName}?op=update-workflow`).send(updatedSimpleWorkflow)
        expect(createWorkflowResponse.statusCode).toEqual(200)
    })

    it(`should update a workflow for the second time`, async () => {
        var createWorkflowResponse = await request(common.config.getDirektivHost()).post(`/api/namespaces/${namespaceName}/tree/${workflowName}?op=update-workflow`).send(updatedSimpleWorkflow)
        expect(createWorkflowResponse.statusCode).toEqual(200)
    })

    it(`should read the root directory`, async () => {
        var readRootDirResponse = await request(common.config.getDirektivHost()).get(`/api/namespaces/${namespaceName}/tree/`)
        expect(readRootDirResponse.statusCode).toEqual(200)
        expect(readRootDirResponse.body).toMatchObject({
            namespace: namespaceName,
            node: {
                name: "",
                path: "/",
                parent: "/",
                type: common.filesystem.nodeTypeDirectory,
                attributes: [],
                oid: "",
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
                results: expect.arrayContaining([expectedChildNodeObject]),
            }
        })
    })

    it(`should read the workflow node`, async () => {
        var readWorkflowResponse = await request(common.config.getDirektivHost()).get(`/api/namespaces/${namespaceName}/tree/${workflowName}`)
        expect(readWorkflowResponse.statusCode).toEqual(200)
        expect(readWorkflowResponse.body).toMatchObject({
            namespace: namespaceName,
            node: expectedChildNodeObject,
        })
    })

    // TODO: post identical
    // TODO: post non identical

    it(`should read the workflow tags`, async () => {
        var readTagsResponse = await request(common.config.getDirektivHost()).get(`/api/namespaces/${namespaceName}/tree/${workflowName}?op=tags`)
        expect(readTagsResponse.statusCode).toEqual(200)
        expect(readTagsResponse.body).toMatchObject({
            namespace: namespaceName,
            node: expectedChildNodeObject,
            pageInfo: {
                limit: 0,
                offset: 0,
                total: 1,
                order: [],
                filter: [],
            },
            results: [{name: 'latest'}],
        })
    })

    it(`should read the workflow refs`, async () => {
        var readRefsResponse = await request(common.config.getDirektivHost()).get(`/api/namespaces/${namespaceName}/tree/${workflowName}?op=refs`)
        expect(readRefsResponse.statusCode).toEqual(200)
        expect(readRefsResponse.body).toMatchObject({
            namespace: namespaceName,
            node: expectedChildNodeObject,
            pageInfo: {
                limit: 0,
                offset: 0,
                total: 1,
                order: [],
                filter: [],
            },
            results: [{name: 'latest'}],
        })
    })

    it(`should read the workflow revisions`, async () => {
        var readRevsResponse = await request(common.config.getDirektivHost()).get(`/api/namespaces/${namespaceName}/tree/${workflowName}?op=revisions`)
        expect(readRevsResponse.statusCode).toEqual(200)
        expect(readRevsResponse.body).toMatchObject({
            namespace: namespaceName,
            node: expectedChildNodeObject,
            pageInfo: {
                limit: 0,
                offset: 0,
                total: 0,
                order: [],
                filter: [],
            },
            results: [],
        })
    })

    it(`should read the workflow router`, async () => {
        var readRouterResponse = await request(common.config.getDirektivHost()).get(`/api/namespaces/${namespaceName}/tree/${workflowName}?op=router`)
        expect(readRouterResponse.statusCode).toEqual(200)
        expect(readRouterResponse.body).toMatchObject({
            namespace: namespaceName,
            node: expectedChildNodeObject,
            live: true,
            routes: [],
        })
    })

    // TODO: tags pagination / filtering / ordering
    // TODO: refs pagination / filtering / ordering
    // TODO: revisions stuff pagination / filtering / ordering
    // TODO: router stuff
    // TODO: update / save / discard
    // TODO: validator paths

    it(`should fail to delete the workflow due to a missing op param`, async () => {
        var deleteWorkflowResponse = await request(common.config.getDirektivHost()).delete(`/api/namespaces/${namespaceName}/tree/${workflowName}`)
        expect(deleteWorkflowResponse.statusCode).toEqual(405)
        expect(deleteWorkflowResponse.body).toMatchObject({})
    })

    it(`should fail to delete the workflow due to a bad method`, async () => {
        var deleteWorkflowResponse = await request(common.config.getDirektivHost()).post(`/api/namespaces/${namespaceName}/tree/${workflowName}?op=delete-node`)
        expect(deleteWorkflowResponse.statusCode).toEqual(405)
        expect(deleteWorkflowResponse.body).toMatchObject({})

        deleteWorkflowResponse = await request(common.config.getDirektivHost()).put(`/api/namespaces/${namespaceName}/tree/${workflowName}?op=delete-node`)
        expect(deleteWorkflowResponse.statusCode).toEqual(405)
        expect(deleteWorkflowResponse.body).toMatchObject({})
    })

    it(`should delete the workflow`, async () => {
        var deleteWorkflowResponse = await request(common.config.getDirektivHost()).delete(`/api/namespaces/${namespaceName}/tree/${workflowName}?op=delete-node`)
        expect(deleteWorkflowResponse.statusCode).toEqual(200)
        expect(deleteWorkflowResponse.body).toMatchObject({})
    })

    it(`should fail to delete a non-existant workflow`, async () => {
        var deleteWorkflowResponse = await request(common.config.getDirektivHost()).delete(`/api/namespaces/${namespaceName}/tree/${workflowName}?op=delete-node`)
        expect(deleteWorkflowResponse.statusCode).toEqual(404)
        expect(deleteWorkflowResponse.body).toMatchObject({})
    })

    // TODO: test node name regex compliance
    // TODO: test everything with/without trailing slash
    // TODO: test delete
    // TODO: test pagination
    // TODO: test filtering
    // TODO: test ordering
    // TODO: test logs

    // TODO: test all sorts of workflow linting & validation
})