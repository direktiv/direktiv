import request from 'supertest'

import common from "../common"

const namespaceName = "mirtest"
const url = "https://github.com/direktiv/direktiv-test-project.git"
const branch = "main"

var activityId = ""

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
                results: expect.arrayContaining([
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

    it(`should get mirror info`, async () => {
        var req = await request(common.config.getDirektivHost()).get(`/api/namespaces/${namespaceName}/tree/?op=mirror-info`)
        expect(req.statusCode).toEqual(200)
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
                        status: "complete", // TODO: polling
                        createdAt: expect.stringMatching(common.regex.timestampRegex),
                        updatedAt: expect.stringMatching(common.regex.timestampRegex),
                    },
                ],
            }
        })

        activityId = req.body.activities.results[0].id
    })

    it(`should check for the expected list of namespace variables`, async () => {
        const req = await request(common.config.getDirektivHost()).get(`/api/namespaces/${namespaceName}/vars`)
        expect(req.statusCode).toEqual(200)
        // console.log(req.body.variables.results)
        expect(req.body).toMatchObject({
            namespace: namespaceName,
            variables: {
                pageInfo: {
                    limit: 0,
                    offset: 0,
                    total: 4,
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
                        checksum: "a36b1f2c3f84522dd1005145646617d7054c0851e97c72a039c0bdfac9fa07f3",
                        mimeType: "",
                        name: "alpha.json",
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
