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
    afterAll(common.helpers.deleteAllNamespaces)

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
                oid: '', // TODO: revisit
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
                    pageInfo: {
                        limit: 0,
                        offset: 0,
                        total: 1,
                        order: [],
                        filter: [],
                    },
                    results: [
                        {
                            id: expect.stringMatching(common.regex.uuidRegex),
                            type: "init",
                            status: expect.stringMatching("^complete|pending$"), // TODO: polling
                            createdAt: expect.stringMatching(common.regex.timestampRegex),
                            updatedAt: expect.stringMatching(common.regex.timestampRegex),
                        },
                    ],
                }
            })

            activityId = req.body.activities.results[0].id
            status = req.body.activities.results[0].status

        } while(status == "pending")

    })

    it(`should read the root directory`, async () => {
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
                readOnly: true,
                expandedType: common.filesystem.extendedNodeTypeMirror,
            },
            children: {
                pageInfo: {
                    limit: 0,
                    offset: 0,
                    total: 4,
                    order: [],
                    filter: [],
                },
                results: expect.arrayContaining([
                    {
                        name: "a",
                        path: "/a",
                        parent: "/",
                        type: common.filesystem.nodeTypeWorkflow,
                        attributes: [],
                        oid: "",
                        readOnly: true,
                        expandedType: common.filesystem.extendedNodeTypeWorkflow,
                        createdAt: expect.stringMatching(common.regex.timestampRegex),
                        updatedAt: expect.stringMatching(common.regex.timestampRegex),
                    },
                    {
                        name: "broken",
                        path: "/broken",
                        parent: "/",
                        type: common.filesystem.nodeTypeWorkflow,
                        attributes: [],
                        oid: "",
                        readOnly: true,
                        expandedType: common.filesystem.extendedNodeTypeWorkflow,
                        createdAt: expect.stringMatching(common.regex.timestampRegex),
                        updatedAt: expect.stringMatching(common.regex.timestampRegex),
                    },
                    {
                        name: "listener",
                        path: "/listener",
                        parent: "/",
                        type: common.filesystem.nodeTypeWorkflow,
                        attributes: [],
                        oid: "",
                        readOnly: true,
                        expandedType: common.filesystem.extendedNodeTypeWorkflow,
                        createdAt: expect.stringMatching(common.regex.timestampRegex),
                        updatedAt: expect.stringMatching(common.regex.timestampRegex),
                    },
                    {
                        name: "apple",
                        path: "/apple",
                        parent: "/",
                        type: common.filesystem.nodeTypeDirectory,
                        attributes: [],
                        oid: "",
                        readOnly: true,
                        expandedType: common.filesystem.extendedNodeTypeDirectory,
                        createdAt: expect.stringMatching(common.regex.timestampRegex),
                        updatedAt: expect.stringMatching(common.regex.timestampRegex),
                    },
                ]),
            },
        })
    })

    it(`should read the '/a' workflow node`, async () => {
        const req = await request(common.config.getDirektivHost()).get(`/api/namespaces/${namespaceName}/tree/a`)
        expect(req.statusCode).toEqual(200)
        expect(req.body).toMatchObject({
            namespace: namespaceName,
            node: {
                createdAt: expect.stringMatching(common.regex.timestampRegex),
                updatedAt: expect.stringMatching(common.regex.timestampRegex),
                name: `a`,
                path: `/a`,
                parent: `/`,
                type: common.filesystem.nodeTypeWorkflow,
                expandedType: common.filesystem.extendedNodeTypeWorkflow,
                attributes: expect.anything(),
                oid: '',
                readOnly: true,
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

    it(`should read the '/broken' workflow node`, async () => {
        const req = await request(common.config.getDirektivHost()).get(`/api/namespaces/${namespaceName}/tree/broken`)
        expect(req.statusCode).toEqual(200)
        expect(req.body).toMatchObject({
            namespace: namespaceName,
            node: {
                createdAt: expect.stringMatching(common.regex.timestampRegex),
                updatedAt: expect.stringMatching(common.regex.timestampRegex),
                name: `broken`,
                path: `/broken`,
                parent: `/`,
                type: common.filesystem.nodeTypeWorkflow,
                expandedType: common.filesystem.extendedNodeTypeWorkflow,
                attributes: expect.anything(),
                oid: '',
                readOnly: true,
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

    it(`should read the '/listener' workflow node`, async () => {
        const req = await request(common.config.getDirektivHost()).get(`/api/namespaces/${namespaceName}/tree/listener`)
        expect(req.statusCode).toEqual(200)
        expect(req.body).toMatchObject({
            namespace: namespaceName,
            node: {
                createdAt: expect.stringMatching(common.regex.timestampRegex),
                updatedAt: expect.stringMatching(common.regex.timestampRegex),
                name: `listener`,
                path: `/listener`,
                parent: `/`,
                type: common.filesystem.nodeTypeWorkflow,
                expandedType: common.filesystem.extendedNodeTypeWorkflow,
                attributes: expect.anything(),
                oid: '',
                readOnly: true,
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

    it(`should check for the expected list of namespace variables`, async () => {
        const req = await request(common.config.getDirektivHost()).get(`/api/namespaces/${namespaceName}/vars`)
        expect(req.statusCode).toEqual(200)
        expect(req.body).toMatchObject({
            namespace: namespaceName,
            variables: {
                pageInfo: {
                    limit: 0,
                    offset: 0,
                    total: 9,
                    order: [],
                    filter: [],
                },
                results: expect.arrayContaining([
                    {
                        checksum: "a386b3c9b4b4786df5bb6474bab0a62b8476e2f3f9c8a6433aca40152840f6b7",
                        mimeType: "",
                        name: "alpha.csv",
                        size: "7", // TODO: this is a string, which is probably a bug
                        createdAt: expect.stringMatching(common.regex.timestampRegex),
                        updatedAt: expect.stringMatching(common.regex.timestampRegex),
                    },
                    {
                        checksum: "38b2babda7c6c19238b4546403c5db1373c05204fdfcc403ee2104176d5eccbf",
                        mimeType: "",
                        name: "alp-ha.json",
                        size: "9",
                        createdAt: expect.stringMatching(common.regex.timestampRegex),
                        updatedAt: expect.stringMatching(common.regex.timestampRegex),
                    },
                    {
                        checksum: "38b2babda7c6c19238b4546403c5db1373c05204fdfcc403ee2104176d5eccbf",
                        mimeType: "",
                        name: "alp_ha.json",
                        size: "9",
                        createdAt: expect.stringMatching(common.regex.timestampRegex),
                        updatedAt: expect.stringMatching(common.regex.timestampRegex),
                    },
                    {
                        checksum: "a36b1f2c3f84522dd1005145646617d7054c0851e97c72a039c0bdfac9fa07f3",
                        mimeType: "",
                        name: "alpha.json",
                        size: "9",
                        createdAt: expect.stringMatching(common.regex.timestampRegex),
                        updatedAt: expect.stringMatching(common.regex.timestampRegex),
                    },
                    {
                        checksum: "38b2babda7c6c19238b4546403c5db1373c05204fdfcc403ee2104176d5eccbf",
                        mimeType: "",
                        name: "alpha_.json",
                        size: "9",
                        createdAt: expect.stringMatching(common.regex.timestampRegex),
                        updatedAt: expect.stringMatching(common.regex.timestampRegex),
                    },
                    {
                        checksum: "38b2babda7c6c19238b4546403c5db1373c05204fdfcc403ee2104176d5eccbf",
                        mimeType: "",
                        name: "ALPHA.json",
                        size: "9",
                        createdAt: expect.stringMatching(common.regex.timestampRegex),
                        updatedAt: expect.stringMatching(common.regex.timestampRegex),
                    },
                    {
                        checksum: "a36b1f2c3f84522dd1005145646617d7054c0851e97c72a039c0bdfac9fa07f3",
                        mimeType: "",
                        name: "beta.json",
                        size: "9",
                        createdAt: expect.stringMatching(common.regex.timestampRegex),
                        updatedAt: expect.stringMatching(common.regex.timestampRegex),
                    },
                    {
                        checksum: expect.stringMatching(common.regex.hashRegex),
                        mimeType: "",
                        name: "delta",
                        size: expect.stringMatching(/^[0-9]*$/), // This archive changes every time. Presumably because of timestamps in the tar archive.
                        createdAt: expect.stringMatching(common.regex.timestampRegex),
                        updatedAt: expect.stringMatching(common.regex.timestampRegex),
                    },
                    {
                        checksum: "457de4239fb1beaad00cbecb6815a9d873a090bf4b1e2cea79c6c9ae48fdedd5",
                        mimeType: "",
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
        const req = await request(common.config.getDirektivHost()).get(`/api/namespaces/${namespaceName}/activities/${activityId}/logs`)
        expect(req.statusCode).toEqual(200)
        // console.log(req.body)
        // TODO: the logic doesn't currently log many errors
    })

    it(`should check for event filters`, async () => {
        const req = await request(common.config.getDirektivHost()).get(`/api/namespaces/${namespaceName}/eventfilter`)
        expect(req.statusCode).toEqual(200)
        // console.log(req.body)
        // TODO: I think was an idea for a feature that was never implemented.
    })

    it(`should invoke the '/a' workflow`, async () => {
        const req = await request(common.config.getDirektivHost()).get(`/api/namespaces/${namespaceName}/tree/a?op=wait`)
        expect(req.statusCode).toEqual(200)
        expect(req.body).toMatchObject({
            msg: 'Hello, world!',
        })
    })

    it(`should fail to invoke the '/broken' workflow`, async () => {
        const req = await request(common.config.getDirektivHost()).get(`/api/namespaces/${namespaceName}/tree/broken?op=wait`)
        expect(req.statusCode).toEqual(500)
        expect(req.body).toMatchObject({
            code: 500,
            message: "cannot parse workflow 'broken': workflow has no defined states",
        })
    })

    it(`should fail to invoke the '/listener' workflow`, async () => {
        const req = await request(common.config.getDirektivHost()).get(`/api/namespaces/${namespaceName}/tree/listener?op=wait`)
        expect(req.statusCode).toEqual(500)
        expect(req.body).toMatchObject({
            code: 500,
            message: 'cannot manually invoke event-based workflow',
        })
    })

    it(`should invoke the '/listener' workflow with an event`, async () => {
        var req = await request(common.config.getDirektivHost()).post(`/api/namespaces/${namespaceName}/broadcast`).send({
            "specversion" : "1.0",
            "type" : "greeting",
            "source" : "https://github.com/cloudevents/spec/pull",
            "subject" : "123",
            "time" : "2018-04-05T17:31:00Z",
            "comexampleextension1" : "value",
            "comexampleothervalue" : 5,
            "datacontenttype" : "text/xml",
            "data" : "<much wow=\"xml\"/>",
        })
        expect(req.statusCode).toEqual(200)
        expect(req.body).toMatchObject({})

        var invoked = false
        var counter = -1
        do {

            counter++
            if (counter > 100) {
                fail('invoke workflow took too long')
            }

            await sleep(100)

            req = await request(common.config.getDirektivHost()).get(`/api/namespaces/${namespaceName}/instances`)
            expect(req.statusCode).toEqual(200)
            
            if (req.body.instances.pageInfo.total > 1) {
                invoked = true
            }

        } while (!invoked)

        expect(req.body).toMatchObject({
            namespace: namespaceName,
            instances: {
                pageInfo: {
                    limit: 0,
                    offset: 0,
                    total: 2,
                    order: [],
                    filter: [],
                },
                results: expect.arrayContaining([
                    {
                        as: "listener",
                        createdAt: expect.stringMatching(common.regex.timestampRegex),
                        updatedAt: expect.stringMatching(common.regex.timestampRegex),
                        errorCode: "",
                        errorMessage: "",
                        id: expect.stringMatching(common.regex.uuidRegex),
                        invoker: "cloudevent",
                        status: "complete", // TODO: polling
                    },
                ]),
            },
        })

    })

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
