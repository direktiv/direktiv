import request from 'supertest'

import common from "../common"

const namespaceName = "mirtest"
const url = "https://github.com/direktiv/direktiv-test-project.git"
const branch = "main"

var activityId = ""

function sleep(ms) {
    return new Promise(resolve => setTimeout(resolve, ms))
}

describe('Test behaviour specific to the root node', () => {
    beforeAll(common.helpers.deleteAllNamespaces)

    it(`should create a namespace`, async () => {
        var req = await request(common.config.getDirektivHost()).put(`/api/namespaces/${namespaceName}`)
            .send({
                url: url,
                ref: branch,
                cron: "",
                passphrase: "",
                publicKey: "",
                privateKey: ""
            })
        expect(req.statusCode).toEqual(200)
        expect(req.body).toMatchObject({
            namespace: {
                createdAt: expect.stringMatching(common.regex.timestampRegex),
                updatedAt: expect.stringMatching(common.regex.timestampRegex),
                name: namespaceName,
                oid: expect.stringMatching(common.regex.uuidRegex),
            },
        })
    })

    it(`should get mirror info`, async () => {
        var status = "pending"
        var counter = -1
        do {

            counter++
            if (counter > 100) {
                fail('init activity took too long')
            }

            await sleep(100)

            var req = await request(common.config.getDirektivHost()).get(`/api/namespaces/${namespaceName}/tree/?op=mirror-info`)
            // expect(req.statusCode).toEqual(200)
            expect(req.body).toMatchObject({
                namespace: namespaceName,
                info: {
                    url: url,
                    ref: branch,
                    cron: "",
                    publicKey: "",
                    commitId: "",
                    lastSync: null,
                    privateKey: "",
                    passphrase: "",
                },
                activities: {
                    pageInfo: null,
                    results: [
                        {
                            id: expect.stringMatching(common.regex.uuidRegex),
                            type: "init",
                            status: expect.stringMatching("^complete|pending|executing$"), // TODO: polling
                            createdAt: expect.stringMatching(common.regex.timestampRegex),
                            updatedAt: expect.stringMatching(common.regex.timestampRegex),
                        },
                    ],
                }
            })

            activityId = req.body.activities.results[0].id
            status = req.body.activities.results[0].status

        } while (status == "pending")

    })

    it(`should read the root directory`, async () => {
        await sleep(5000)
        var req = await request(common.config.getDirektivHost()).get(`/api/namespaces/${namespaceName}/tree/`)
        expect(req.statusCode).toEqual(200)
        expect(req.body).toMatchObject({
            namespace: namespaceName,
            node: {
                name: "",
                path: "/",
                parent: "/",
                type: common.filesystem.nodeTypeDirectory,
                attributes: [],
                oid: "",
                readOnly: false,
                expandedType: common.filesystem.extendedNodeTypeMirror,
                mimeType: "",
            },
            children: {
                pageInfo: {
                    limit: 0,
                    offset: 0,
                    total: 16,
                    order: [],
                    filter: [],
                },
                results: expect.arrayContaining([
                    {
                        name: "apple",
                        path: "/apple",
                        parent: "/",
                        type: common.filesystem.nodeTypeDirectory,
                        attributes: [],
                        oid: "",
                        readOnly: false,
                        mimeType: "",
                        expandedType: common.filesystem.extendedNodeTypeDirectory,
                        createdAt: expect.stringMatching(common.regex.timestampRegex),
                        updatedAt: expect.stringMatching(common.regex.timestampRegex),
                    },
                    {
                        name: "a.yaml",
                        path: "/a.yaml",
                        parent: "/",
                        type: common.filesystem.nodeTypeWorkflow,
                        attributes: [],
                        oid: "",
                        readOnly: false,
                        mimeType: "application/direktiv",
                        expandedType: common.filesystem.extendedNodeTypeWorkflow,
                        createdAt: expect.stringMatching(common.regex.timestampRegex),
                        updatedAt: expect.stringMatching(common.regex.timestampRegex),
                    },
                    {
                        name: "a.yml",
                        path: "/a.yml",
                        parent: "/",
                        type: common.filesystem.nodeTypeWorkflow,
                        attributes: [],
                        oid: "",
                        readOnly: false,
                        mimeType: "application/direktiv",
                        expandedType: common.filesystem.extendedNodeTypeWorkflow,
                        createdAt: expect.stringMatching(common.regex.timestampRegex),
                        updatedAt: expect.stringMatching(common.regex.timestampRegex),
                    },
                    {
                        name: "banana",
                        path: "/banana",
                        parent: "/",
                        type: common.filesystem.nodeTypeDirectory,
                        attributes: [],
                        oid: "",
                        readOnly: false,
                        mimeType: "",
                        expandedType: common.filesystem.extendedNodeTypeDirectory,
                        createdAt: expect.stringMatching(common.regex.timestampRegex),
                        updatedAt: expect.stringMatching(common.regex.timestampRegex),
                    },
                    {
                        name: "broken.yaml",
                        path: "/broken.yaml",
                        parent: "/",
                        type: common.filesystem.nodeTypeFile,
                        attributes: [],
                        oid: "",
                        readOnly: false,
                        mimeType: "text/plain",
                        expandedType: common.filesystem.nodeTypeFile,
                        createdAt: expect.stringMatching(common.regex.timestampRegex),
                        updatedAt: expect.stringMatching(common.regex.timestampRegex),
                    },
                    {
                        name: "broke._yaml",
                        path: "/broke._yaml",
                        parent: "/",
                        type: common.filesystem.nodeTypeFile,
                        attributes: [],
                        oid: "",
                        readOnly: false,
                        mimeType: "text/plain",
                        expandedType: common.filesystem.nodeTypeFile,
                        createdAt: expect.stringMatching(common.regex.timestampRegex),
                        updatedAt: expect.stringMatching(common.regex.timestampRegex),
                    },
                    {
                        name: "listener.yml",
                        path: "/listener.yml",
                        parent: "/",
                        type: common.filesystem.nodeTypeWorkflow,
                        attributes: [],
                        oid: "",
                        readOnly: false,
                        mimeType: "application/direktiv",
                        expandedType: common.filesystem.extendedNodeTypeWorkflow,
                        createdAt: expect.stringMatching(common.regex.timestampRegex),
                        updatedAt: expect.stringMatching(common.regex.timestampRegex),
                    },
                    {
                        name: "var.alpha.csv",
                        path: "/var.alpha.csv",
                        parent: "/",
                        type: common.filesystem.nodeTypeFile,
                        attributes: [],
                        oid: "",
                        readOnly: false,
                        mimeType: "text/plain",
                        expandedType: common.filesystem.nodeTypeFile,
                        createdAt: expect.stringMatching(common.regex.timestampRegex),
                        updatedAt: expect.stringMatching(common.regex.timestampRegex),
                    },
                    {
                        name: "var._alpha.json",
                        path: "/var._alpha.json",
                        parent: "/",
                        type: common.filesystem.nodeTypeFile,
                        attributes: [],
                        oid: "",
                        readOnly: false,
                        mimeType: "application/json",
                        expandedType: common.filesystem.nodeTypeFile,
                        createdAt: expect.stringMatching(common.regex.timestampRegex),
                        updatedAt: expect.stringMatching(common.regex.timestampRegex),
                    },
                    {
                        name: "var.alp-ha.json",
                        path: "/var.alp-ha.json",
                        parent: "/",
                        type: common.filesystem.nodeTypeFile,
                        attributes: [],
                        oid: "",
                        readOnly: false,
                        mimeType: "application/json",
                        expandedType: common.filesystem.nodeTypeFile,
                        createdAt: expect.stringMatching(common.regex.timestampRegex),
                        updatedAt: expect.stringMatching(common.regex.timestampRegex),
                    },
                    {
                        name: "var.alp_ha.json",
                        path: "/var.alp_ha.json",
                        parent: "/",
                        type: common.filesystem.nodeTypeFile,
                        attributes: [],
                        oid: "",
                        readOnly: false,
                        mimeType: "application/json",
                        expandedType: common.filesystem.nodeTypeFile,
                        createdAt: expect.stringMatching(common.regex.timestampRegex),
                        updatedAt: expect.stringMatching(common.regex.timestampRegex),
                    },
                    {
                        name: "var.alpha_.json",
                        path: "/var.alpha_.json",
                        parent: "/",
                        type: common.filesystem.nodeTypeFile,
                        attributes: [],
                        oid: "",
                        readOnly: false,
                        mimeType: "application/json",
                        expandedType: common.filesystem.nodeTypeFile,
                        createdAt: expect.stringMatching(common.regex.timestampRegex),
                        updatedAt: expect.stringMatching(common.regex.timestampRegex),
                    },
                    {
                        name: "var.ALPHA.json",
                        path: "/var.ALPHA.json",
                        parent: "/",
                        type: common.filesystem.nodeTypeFile,
                        attributes: [],
                        oid: "",
                        readOnly: false,
                        mimeType: "application/json",
                        expandedType: common.filesystem.nodeTypeFile,
                        createdAt: expect.stringMatching(common.regex.timestampRegex),
                        updatedAt: expect.stringMatching(common.regex.timestampRegex),
                    },
                    {
                        name: "var.beta.json",
                        path: "/var.beta.json",
                        parent: "/",
                        type: common.filesystem.nodeTypeFile,
                        attributes: [],
                        oid: "",
                        readOnly: false,
                        mimeType: "application/json",
                        expandedType: common.filesystem.nodeTypeFile,
                        createdAt: expect.stringMatching(common.regex.timestampRegex),
                        updatedAt: expect.stringMatching(common.regex.timestampRegex),
                    },
                    {
                        name: "var.delta",
                        path: "/var.delta",
                        parent: "/",
                        type: common.filesystem.nodeTypeDirectory,
                        attributes: [],
                        oid: "",
                        readOnly: false,
                        mimeType: "",
                        expandedType: common.filesystem.nodeTypeDirectory,
                        createdAt: expect.stringMatching(common.regex.timestampRegex),
                        updatedAt: expect.stringMatching(common.regex.timestampRegex),
                    },
                    {
                        name: "var.gamma.css",
                        path: "/var.gamma.css",
                        parent: "/",
                        type: common.filesystem.nodeTypeFile,
                        attributes: [],
                        oid: "",
                        readOnly: false,
                        mimeType: "text/plain",
                        expandedType: common.filesystem.nodeTypeFile,
                        createdAt: expect.stringMatching(common.regex.timestampRegex),
                        updatedAt: expect.stringMatching(common.regex.timestampRegex),
                    },
                ]),
            },
        })
    })

    it(`should read the '/a.yaml' workflow node`, async () => {
        const req = await request(common.config.getDirektivHost()).get(`/api/namespaces/${namespaceName}/tree/a.yaml`)
        expect(req.statusCode).toEqual(200)
        expect(req.body).toMatchObject({
            namespace: namespaceName,
            node: {
                createdAt: expect.stringMatching(common.regex.timestampRegex),
                updatedAt: expect.stringMatching(common.regex.timestampRegex),
                name: `a.yaml`,
                path: `/a.yaml`,
                parent: `/`,
                type: common.filesystem.nodeTypeWorkflow,
                expandedType: common.filesystem.extendedNodeTypeWorkflow,
                attributes: expect.anything(),
                oid: '',
                readOnly: false,
                mimeType: "application/direktiv",
            },
            revision: {
                createdAt: expect.stringMatching(common.regex.timestampRegex),
                hash: "d4ac523a7b82b805eb0bec604ce16cfb0a4e54c9280bb98fe4e1b58e8722c1d9",
                source: expect.stringMatching(common.regex.base64Regex),
                name: expect.stringMatching(common.regex.uuidRegex),
            },
            eventLogging: ``,
            oid: expect.stringMatching(common.regex.uuidRegex),
        })
    })

    it(`should read the '/broken.yaml' workflow node`, async () => {
        const req = await request(common.config.getDirektivHost()).get(`/api/namespaces/${namespaceName}/tree/broken.yaml`)
        expect(req.statusCode).toEqual(200)
        expect(req.body).toMatchObject({
            namespace: namespaceName,
            node: {
                createdAt: expect.stringMatching(common.regex.timestampRegex),
                updatedAt: expect.stringMatching(common.regex.timestampRegex),
                name: `broken.yaml`,
                path: `/broken.yaml`,
                parent: `/`,
                type: common.filesystem.nodeTypeFile,
                expandedType: common.filesystem.nodeTypeFile,
                attributes: expect.anything(),
                oid: '',
                readOnly: false,
                mimeType: "text/plain",
            },
            revision: {
                createdAt: expect.stringMatching(common.regex.timestampRegex),
                hash: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
                source: expect.stringMatching(common.regex.base64Regex),
                name: expect.stringMatching(common.regex.uuidRegex),
            },
            eventLogging: ``,
            oid: expect.stringMatching(common.regex.uuidRegex),
        })
    })

    it(`should read the '/listener.yml' workflow node`, async () => {
        const req = await request(common.config.getDirektivHost()).get(`/api/namespaces/${namespaceName}/tree/listener.yml`)
        expect(req.statusCode).toEqual(200)
        expect(req.body).toMatchObject({
            namespace: namespaceName,
            node: {
                createdAt: expect.stringMatching(common.regex.timestampRegex),
                updatedAt: expect.stringMatching(common.regex.timestampRegex),
                name: `listener.yml`,
                path: `/listener.yml`,
                parent: `/`,
                type: common.filesystem.nodeTypeWorkflow,
                expandedType: common.filesystem.extendedNodeTypeWorkflow,
                attributes: expect.anything(),
                oid: '',
                readOnly: false,
                mimeType: "application/direktiv",
            },
            revision: {
                createdAt: expect.stringMatching(common.regex.timestampRegex),
                hash: "2a4f39df7002abc30c919d47a62b06c7a4b978a384a4ac2f93c18fb0f56adab6",
                source: expect.stringMatching(common.regex.base64Regex),
                name: expect.stringMatching(common.regex.uuidRegex),
            },
            eventLogging: ``,
            oid: expect.stringMatching(common.regex.uuidRegex),
        })
    })

    it(`should read the /apple node`, async () => {
        const req = await request(common.config.getDirektivHost()).get(`/api/namespaces/${namespaceName}/tree/apple`)
        expect(req.statusCode).toEqual(200)
        expect(req.body).toMatchObject({
            namespace: namespaceName,
            node: {
                createdAt: expect.stringMatching(common.regex.timestampRegex),
                updatedAt: expect.stringMatching(common.regex.timestampRegex),
                name: `apple`,
                path: `/apple`,
                parent: `/`,
                type: common.filesystem.nodeTypeDirectory,
                expandedType: common.filesystem.extendedNodeTypeDirectory,
                attributes: expect.anything(),
                oid: '',
                readOnly: false,
                mimeType: "",
            },
            children: {
                pageInfo: {
                    order: [],
                    filter: [],
                    limit: 0,
                    offset: 0,
                    total: 2,
                },
            },
        })
    })

    it(`should read the /banana node`, async () => {
        const req = await request(common.config.getDirektivHost()).get(`/api/namespaces/${namespaceName}/tree/banana`)
        expect(req.statusCode).toEqual(200)
        expect(req.body).toMatchObject({
            namespace: namespaceName,
            node: {
                createdAt: expect.stringMatching(common.regex.timestampRegex),
                updatedAt: expect.stringMatching(common.regex.timestampRegex),
                name: `banana`,
                path: `/banana`,
                parent: `/`,
                type: common.filesystem.nodeTypeDirectory,
                expandedType: common.filesystem.extendedNodeTypeDirectory,
                attributes: expect.anything(),
                oid: '',
                readOnly: false,
                mimeType: "",
            },
            children: {
                pageInfo: {
                    order: [],
                    filter: [],
                    limit: 0,
                    offset: 0,
                    total: 6,
                },
                results: expect.arrayContaining([
                    {
                        attributes: [],
                        createdAt: expect.stringMatching(common.regex.timestampRegex),
                        updatedAt: expect.stringMatching(common.regex.timestampRegex),
                        type: common.filesystem.nodeTypeWorkflow,
                        expandedType: common.filesystem.extendedNodeTypeWorkflow,
                        name: "css.yaml",
                        oid: "",
                        parent: "/banana",
                        path: "/banana/css.yaml",
                        readOnly: false,
                        mimeType: "application/direktiv",
                    },
                    {
                        attributes: [],
                        createdAt: expect.stringMatching(common.regex.timestampRegex),
                        updatedAt: expect.stringMatching(common.regex.timestampRegex),
                        type: "file",
                        expandedType: "file",
                        name: "page-1.yaml.page.html",
                        oid: "",
                        parent: "/banana",
                        path: "/banana/page-1.yaml.page.html",
                        readOnly: false,
                        mimeType: expect.anything(),
                    },
                    {
                        attributes: [],
                        createdAt: expect.stringMatching(common.regex.timestampRegex),
                        updatedAt: expect.stringMatching(common.regex.timestampRegex),
                        type: "file",
                        expandedType: "file",
                        name: "page-2.yaml.Page.HTML",
                        oid: "",
                        parent: "/banana",
                        path: "/banana/page-2.yaml.Page.HTML",
                        readOnly: false,
                        mimeType: expect.anything(),
                    },
                    {
                        attributes: [],
                        createdAt: expect.stringMatching(common.regex.timestampRegex),
                        updatedAt: expect.stringMatching(common.regex.timestampRegex),
                        type: common.filesystem.nodeTypeDirectory,
                        expandedType: common.filesystem.extendedNodeTypeDirectory,
                        name: "util",
                        oid: "",
                        parent: "/banana",
                        path: "/banana/util",
                        readOnly: false,
                        mimeType: "",
                    },
                ]),
            },
        })
    })

    it(`should read the '/banana/css.yaml' workflow node`, async () => {
        const req = await request(common.config.getDirektivHost()).get(`/api/namespaces/${namespaceName}/tree/banana/css.yaml`)
        expect(req.statusCode).toEqual(200)
        expect(req.body).toMatchObject({
            namespace: namespaceName,
            node: {
                createdAt: expect.stringMatching(common.regex.timestampRegex),
                updatedAt: expect.stringMatching(common.regex.timestampRegex),
                name: `css.yaml`,
                path: `/banana/css.yaml`,
                parent: `/banana`,
                type: common.filesystem.nodeTypeWorkflow,
                expandedType: common.filesystem.extendedNodeTypeWorkflow,
                attributes: expect.anything(),
                oid: '',
                readOnly: false,
                mimeType: "application/direktiv",
            },
            revision: {
                createdAt: expect.stringMatching(common.regex.timestampRegex),
                hash: "a600b303d59570902466822693a92a410bc0b5894f19e85af9b6cbf0d9f2a53b",
                source: expect.stringMatching(common.regex.base64Regex),
                name: expect.stringMatching(common.regex.uuidRegex),
            },
            eventLogging: ``,
            oid: expect.stringMatching(common.regex.uuidRegex),
        })
    })

    it(`should read the '/banana/page-1.yaml.page.html' workflow node`, async () => {
        const req = await request(common.config.getDirektivHost()).get(`/api/namespaces/${namespaceName}/tree/banana/page-1.yaml.page.html`)
        expect(req.statusCode).toEqual(200)
        expect(req.body).toMatchObject({
            namespace: namespaceName,
            node: {
                createdAt: expect.stringMatching(common.regex.timestampRegex),
                updatedAt: expect.stringMatching(common.regex.timestampRegex),
                name: `page-1.yaml.page.html`,
                path: `/banana/page-1.yaml.page.html`,
                parent: `/banana`,
                type: "file",
                expandedType: "file",
                attributes: expect.anything(),
                oid: '',
                readOnly: false,
                mimeType: expect.anything(),
            },
            revision: {
                createdAt: expect.stringMatching(common.regex.timestampRegex),
                //hash: "5595048ad23cdef4a0c10a36e7d9a335264e55182046ed213d5aacda0803812e",
                source: expect.stringMatching(common.regex.base64Regex),
                name: expect.stringMatching(common.regex.uuidRegex),
            },
            eventLogging: ``,
            oid: expect.stringMatching(common.regex.uuidRegex),
        })
    })

    it(`should read the workflow variables of '/banana/page-1.yaml'`, async () => {
        const req = await request(common.config.getDirektivHost()).get(`/api/namespaces/${namespaceName}/tree/banana/page-1.yaml?op=vars`)
        expect(req.statusCode).toEqual(200)
        expect(req.body).toMatchObject({
            namespace: namespaceName,
            path: `/banana/page-1.yaml`,
            variables: {
                pageInfo: null,
                results: [
                    {
                        mimeType: "text/html",
                        name: "page.html",
                        size: "221",
                        checksum: "",
                        createdAt: expect.stringMatching(common.regex.timestampRegex),
                        updatedAt: expect.stringMatching(common.regex.timestampRegex),
                    },
                ],
            }
        })
    })

    it(`should read the '/banana/page-2.yaml' workflow node`, async () => {
        const req = await request(common.config.getDirektivHost()).get(`/api/namespaces/${namespaceName}/tree/banana/page-2.yaml`)
        expect(req.statusCode).toEqual(200)
        expect(req.body).toMatchObject({
            namespace: namespaceName,
            node: {
                createdAt: expect.stringMatching(common.regex.timestampRegex),
                updatedAt: expect.stringMatching(common.regex.timestampRegex),
                name: `page-2.yaml`,
                path: `/banana/page-2.yaml`,
                parent: `/banana`,
                type: common.filesystem.nodeTypeWorkflow,
                expandedType: common.filesystem.extendedNodeTypeWorkflow,
                attributes: expect.anything(),
                oid: '',
                readOnly: false,
                mimeType: "application/direktiv",
            },
            revision: {
                createdAt: expect.stringMatching(common.regex.timestampRegex),
                hash: "5595048ad23cdef4a0c10a36e7d9a335264e55182046ed213d5aacda0803812e",
                source: expect.stringMatching(common.regex.base64Regex),
                name: expect.stringMatching(common.regex.uuidRegex),
            },
            eventLogging: ``,
            oid: expect.stringMatching(common.regex.uuidRegex),
        })
    })

    it(`should read the workflow variables of '/banana/page-2.yaml'`, async () => {
        const req = await request(common.config.getDirektivHost()).get(`/api/namespaces/${namespaceName}/tree/banana/page-2.yaml?op=vars`)
        expect(req.statusCode).toEqual(200)
        expect(req.body).toMatchObject({
            namespace: namespaceName,
            path: `/banana/page-2.yaml`,
            variables: {
                pageInfo: null,
                results: [
                    {
                        mimeType: "text/html",
                        name: "Page.HTML",
                        size: "233",
                        checksum: "",
                        createdAt: expect.stringMatching(common.regex.timestampRegex),
                        updatedAt: expect.stringMatching(common.regex.timestampRegex),
                    },
                ],
            }
        })
    })

    it(`should read the /banana/util node`, async () => {
        const req = await request(common.config.getDirektivHost()).get(`/api/namespaces/${namespaceName}/tree/banana/util`)
        expect(req.statusCode).toEqual(200)
        expect(req.body).toMatchObject({
            namespace: namespaceName,
            node: {
                createdAt: expect.stringMatching(common.regex.timestampRegex),
                updatedAt: expect.stringMatching(common.regex.timestampRegex),
                name: `util`,
                path: `/banana/util`,
                parent: `/banana`,
                type: common.filesystem.nodeTypeDirectory,
                expandedType: common.filesystem.extendedNodeTypeDirectory,
                attributes: expect.anything(),
                oid: '',
                readOnly: false,
                mimeType: "",
            },
            children: {
                pageInfo: {
                    order: [],
                    filter: [],
                    limit: 0,
                    offset: 0,
                    total: 2,
                },
                results: expect.arrayContaining([
                    {
                        attributes: [],
                        createdAt: expect.stringMatching(common.regex.timestampRegex),
                        updatedAt: expect.stringMatching(common.regex.timestampRegex),
                        type: common.filesystem.nodeTypeWorkflow,
                        expandedType: common.filesystem.extendedNodeTypeWorkflow,
                        name: "caller.yaml",
                        oid: "",
                        parent: "/banana/util",
                        path: "/banana/util/caller.yaml",
                        readOnly: false,
                        mimeType: "application/direktiv",
                    },
                    {
                        attributes: [],
                        createdAt: expect.stringMatching(common.regex.timestampRegex),
                        updatedAt: expect.stringMatching(common.regex.timestampRegex),
                        type: common.filesystem.nodeTypeWorkflow,
                        expandedType: common.filesystem.extendedNodeTypeWorkflow,
                        name: "curler.yaml",
                        oid: "",
                        parent: "/banana/util",
                        path: "/banana/util/curler.yaml",
                        readOnly: false,
                        mimeType: "application/direktiv",
                    },
                ]),
            },
        })
    })

    it(`should read the '/banana/util/caller.yaml' workflow node`, async () => {
        const req = await request(common.config.getDirektivHost()).get(`/api/namespaces/${namespaceName}/tree/banana/util/caller.yaml`)
        expect(req.statusCode).toEqual(200)
        expect(req.body).toMatchObject({
            namespace: namespaceName,
            node: {
                createdAt: expect.stringMatching(common.regex.timestampRegex),
                updatedAt: expect.stringMatching(common.regex.timestampRegex),
                name: `caller.yaml`,
                path: `/banana/util/caller.yaml`,
                parent: `/banana/util`,
                type: common.filesystem.nodeTypeWorkflow,
                expandedType: common.filesystem.extendedNodeTypeWorkflow,
                attributes: expect.anything(),
                oid: '',
                readOnly: false,
                mimeType: "application/direktiv",
            },
            revision: {
                createdAt: expect.stringMatching(common.regex.timestampRegex),
                hash: "05729d2916b0cfff71291ca877600173520734f13da273859a9701b8efd10975",
                source: expect.stringMatching(common.regex.base64Regex),
                name: expect.stringMatching(common.regex.uuidRegex),
            },
            eventLogging: ``,
            oid: expect.stringMatching(common.regex.uuidRegex),
        })
    })

    it(`should read the '/banana/util/curler.yaml' workflow node`, async () => {
        const req = await request(common.config.getDirektivHost()).get(`/api/namespaces/${namespaceName}/tree/banana/util/curler.yaml`)
        expect(req.statusCode).toEqual(200)
        expect(req.body).toMatchObject({
            namespace: namespaceName,
            node: {
                createdAt: expect.stringMatching(common.regex.timestampRegex),
                updatedAt: expect.stringMatching(common.regex.timestampRegex),
                name: `curler.yaml`,
                path: `/banana/util/curler.yaml`,
                parent: `/banana/util`,
                type: common.filesystem.nodeTypeWorkflow,
                expandedType: common.filesystem.extendedNodeTypeWorkflow,
                attributes: expect.anything(),
                oid: '',
                readOnly: false,
                mimeType: "application/direktiv",
            },
            revision: {
                createdAt: expect.stringMatching(common.regex.timestampRegex),
                hash: "ac0fea085b3889f7411ef777ed4d89af6d7f7a1ef787cbea37431ae086be1318",
                source: expect.stringMatching(common.regex.base64Regex),
                name: expect.stringMatching(common.regex.uuidRegex),
            },
            eventLogging: ``,
            oid: expect.stringMatching(common.regex.uuidRegex),
        })
    })

    it(`should check for the expected list of namespace variables`, async () => {
        const req = await request(common.config.getDirektivHost()).get(`/api/namespaces/${namespaceName}/vars`)
        expect(req.statusCode).toEqual(200)
        expect(req.body).toMatchObject({
            namespace: namespaceName,
            variables: {
                pageInfo: null,
                results: expect.arrayContaining([
                    {
                        checksum: "",
                        mimeType: "text/plain",
                        name: "alpha.csv",
                        size: "7", // TODO: this is a string, which is probably a bug
                        createdAt: expect.stringMatching(common.regex.timestampRegex),
                        updatedAt: expect.stringMatching(common.regex.timestampRegex),
                    },
                    {
                        checksum: "",
                        mimeType: "application/json",
                        name: "alp-ha.json",
                        size: "9",
                        createdAt: expect.stringMatching(common.regex.timestampRegex),
                        updatedAt: expect.stringMatching(common.regex.timestampRegex),
                    },
                    {
                        checksum: "",
                        mimeType: "application/json",
                        name: "alp_ha.json",
                        size: "9",
                        createdAt: expect.stringMatching(common.regex.timestampRegex),
                        updatedAt: expect.stringMatching(common.regex.timestampRegex),
                    },
                    {
                        checksum: "",
                        mimeType: "application/json",
                        name: "alpha.json",
                        size: "9",
                        createdAt: expect.stringMatching(common.regex.timestampRegex),
                        updatedAt: expect.stringMatching(common.regex.timestampRegex),
                    },
                    {
                        checksum: "",
                        mimeType: "application/json",
                        name: "alpha_.json",
                        size: "9",
                        createdAt: expect.stringMatching(common.regex.timestampRegex),
                        updatedAt: expect.stringMatching(common.regex.timestampRegex),
                    },
                    {
                        checksum: "",
                        mimeType: "application/json",
                        name: "ALPHA.json",
                        size: "9",
                        createdAt: expect.stringMatching(common.regex.timestampRegex),
                        updatedAt: expect.stringMatching(common.regex.timestampRegex),
                    },
                    {
                        checksum: "",
                        mimeType: "application/json",
                        name: "beta.json",
                        size: "9",
                        createdAt: expect.stringMatching(common.regex.timestampRegex),
                        updatedAt: expect.stringMatching(common.regex.timestampRegex),
                    },
                    {
                        checksum: "",
                        mimeType: "application/json",
                        name: "data.json",
                        size: "9",
                        createdAt: expect.stringMatching(common.regex.timestampRegex),
                        updatedAt: expect.stringMatching(common.regex.timestampRegex),
                    },
                    {
                        checksum: "",
                        mimeType: "text/plain",
                        name: "gamma.css",
                        size: "103",
                        createdAt: expect.stringMatching(common.regex.timestampRegex),
                        updatedAt: expect.stringMatching(common.regex.timestampRegex),
                    },
                ]),
            }
        })
    })

    it(`should check the activity logs for errors`, async () => {
        // TODO: this test need to expect stream response.
        return
        const req = await request(common.config.getDirektivHost()).get(`/api/namespaces/${namespaceName}/activities/${activityId}/logs`)
        expect(req.statusCode).toEqual(200)
        // console.log(req.body)
        // TODO: the logic doesn't currently log many errors
    })

    it(`should invoke the '/a.yaml' workflow`, async () => {
        const req = await request(common.config.getDirektivHost()).get(`/api/namespaces/${namespaceName}/tree/a.yaml?op=wait`)
        expect(req.statusCode).toEqual(200)
        expect(req.body).toMatchObject({
            msg: 'Hello, world!',
        })
    })

    it(`should fail to invoke the '/broken.yaml' workflow`, async () => {
        const req = await request(common.config.getDirektivHost()).get(`/api/namespaces/${namespaceName}/tree/broken.yaml?op=wait`)
        expect(req.statusCode).toEqual(500)
        expect(req.body).toMatchObject({
            code: 500,
            message: "cannot parse workflow 'broken.yaml': workflow has no defined states",
        })
    })

    it(`should invoke the '/banana/css.yaml' workflow`, async () => {
        const req = await request(common.config.getDirektivHost()).get(`/api/namespaces/${namespaceName}/tree/banana/css.yaml?op=wait`)
        expect(req.statusCode).toEqual(200)
        expect(req.body).toMatchObject({
            "gamma.css": 'Ym9keSB7CiAgICBiYWNrZ3JvdW5kLWNvbG9yOiBwb3dkZXJibHVlOwogIH0KICBoMSB7CiAgICBjb2xvcjogYmx1ZTsKICB9CiAgcCB7CiAgICBjb2xvcjogcmVkOwogIH0KICAgIA==',
        })
    })

    it(`should invoke the '/banana/page-1.yaml' workflow`, async () => {
        const req = await request(common.config.getDirektivHost()).get(`/api/namespaces/${namespaceName}/tree/banana/page-1.yaml?op=wait`)
        expect(req.statusCode).toEqual(200)
        expect(req.body).toMatchObject({
            "page.html": 'PCFET0NUWVBFIGh0bWw+CjxodG1sPgo8aGVhZD4KICA8bGluayByZWw9InN0eWxlc2hlZXQiIGhyZWY9Ii4vY3NzP29wPXdhaXQmcmVmPWxhdGVzdCZyYXctb3V0cHV0PXRydWUmZmllbGQ9dmFyMy5jc3MmY3R5cGU9dGV4dC9jc3MiPgo8L2hlYWQ+Cjxib2R5PgoKPGgxPlRoaXMgaXMgYSBoZWFkaW5nPC9oMT4KPHA+VGhpcyBpcyBhIHBhcmFncmFwaC48L3A+Cgo8L2JvZHk+CjwvaHRtbD4=',
        })
    })

    // TODO: find a way to enable this as an optional test, because it takes too long to run in most cases.
    it(`should invoke the '/banana/util/caller.yaml' workflow`, async () => {
        // TODO: yassir enable this test before release.
        // disabled for WIP.
        return;
        const req = await request(common.config.getDirektivHost()).get(`/api/namespaces/${namespaceName}/tree/banana/util/caller.yaml?op=wait`)
        expect(req.statusCode).toEqual(200)
        expect(req.body.return.return.status).toEqual('200 OK')
    }, 90000)

    it(`should fail to delete a namespace because of a lack of a recursive param`, async () => {
        const req = await request(common.config.getDirektivHost()).delete(`/api/namespaces/${namespaceName}`)
        expect(req.statusCode).toEqual(500)
        expect(req.body).toMatchObject({})
    })

    it(`should delete a namespace`, async () => {
        const req = await request(common.config.getDirektivHost()).delete(`/api/namespaces/${namespaceName}?recursive=true`)
        expect(req.statusCode).toEqual(200)
        expect(req.body).toMatchObject({})
    })
})
