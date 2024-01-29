import request from 'supertest'
import common from "../common";

const testNamespace = "test-file-namespace"

describe('Test namespaces crud operations', () => {
    beforeAll(common.helpers.deleteAllNamespaces)

    common.helpers.itShouldCreateNamespace(it, expect, testNamespace)

    common.helpers.itShouldCreateFile(it, expect, testNamespace,
        "/f1.yaml", generateTestFile("foo"))

    common.helpers.itShouldCreateFile(it, expect, testNamespace,
        "/f2.yaml", generateTestFile("bar"))

    it(`should read '/f1.yaml' file`, async () => {
        const req = await request(common.config.getDirektivHost()).get(`/api/namespaces/${testNamespace}/tree/f1.yaml`)
        expect(req.statusCode).toEqual(200)
        expect(req.body).toMatchObject({
            namespace: testNamespace,
            node: {
                name: `f1.yaml`,
                path: `/f1.yaml`,
                parent: `/`,
                type: common.filesystem.nodeTypeWorkflow,
                expandedType: common.filesystem.extendedNodeTypeWorkflow,
                mimeType: "application/direktiv",
            },
            source: btoa(generateTestFile("foo")),
        })
    })

    it(`should read '/f2.yaml' file`, async () => {
        const req = await request(common.config.getDirektivHost()).get(`/api/namespaces/${testNamespace}/tree/f2.yaml`)
        expect(req.statusCode).toEqual(200)
        expect(req.body).toMatchObject({
            namespace: testNamespace,
            node: {
                name: `f2.yaml`,
                path: `/f2.yaml`,
                parent: `/`,
                type: common.filesystem.nodeTypeWorkflow,
                expandedType: common.filesystem.extendedNodeTypeWorkflow,
                mimeType: "application/direktiv",
            },
            source: btoa(generateTestFile("bar")),
        })
    })

    common.helpers.itShouldUpdateFile(it, expect, testNamespace,
        "/f2.yaml", generateTestFile("bar2"))

    it(`should read updated '/f2.yaml' file`, async () => {
        const req = await request(common.config.getDirektivHost()).get(`/api/namespaces/${testNamespace}/tree/f2.yaml`)
        expect(req.statusCode).toEqual(200)
        expect(req.body).toMatchObject({
            namespace: testNamespace,
            node: {
                name: `f2.yaml`,
                path: `/f2.yaml`,
                parent: `/`,
                type: common.filesystem.nodeTypeWorkflow,
                expandedType: common.filesystem.extendedNodeTypeWorkflow,
                mimeType: "application/direktiv",
            },
            source: btoa(generateTestFile("bar2")),
        })
    })

    it(`should read unchanged '/f1.yaml' file`, async () => {
        const req = await request(common.config.getDirektivHost()).get(`/api/namespaces/${testNamespace}/tree/f1.yaml`)
        expect(req.statusCode).toEqual(200)
        expect(req.body).toMatchObject({
            namespace: testNamespace,
            node: {
                name: `f1.yaml`,
                path: `/f1.yaml`,
                parent: `/`,
                type: common.filesystem.nodeTypeWorkflow,
                expandedType: common.filesystem.extendedNodeTypeWorkflow,
                mimeType: "application/direktiv",
            },
            source: btoa(generateTestFile("foo")),
        })
    })
})

function generateTestFile(text) {
    return `
description: A simple file with text ${text}'
states:
- id: hello_world
  type: noop
  transform:
    result: Hello world!`
}