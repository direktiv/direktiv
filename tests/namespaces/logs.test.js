import request from "../common/request"

import common from "../common"

const testNamespace = "nslogstest"

describe('Test that basic namespace operations generate expected logs.', () => {
    beforeAll(common.helpers.deleteAllNamespaces)


    it(`should create a namespace`, async () => {
        const createResponse = await request(common.config.getDirektivHost()).put(`/api/namespaces/${testNamespace}`)
        expect(createResponse.statusCode).toEqual(200)
        expect(createResponse.body).toMatchObject({
            namespace: {
                name: testNamespace,
                createdAt: expect.stringMatching(common.regex.timestampRegex),
                updatedAt: expect.stringMatching(common.regex.timestampRegex),
            }
        })
    })

    it(`should attempt to create a namespace that already exists`, async () => {
        const createResponse = await request(common.config.getDirektivHost()).put(`/api/namespaces/${testNamespace}`)
        expect(createResponse.statusCode).toEqual(409)
        expect(createResponse.body).toEqual({
            code: 409,
            message: "resource already exists",
        })
    })

    it(`should delete a namespace`, async () => {
        const deleteResponse = await request(common.config.getDirektivHost()).delete(`/api/namespaces/${testNamespace}`)
        expect(deleteResponse.statusCode).toEqual(200)
        expect(deleteResponse.body).toMatchObject({})
    })

    it("should attempt to delete a namespace that doesn't exist", async () => {
        const deleteResponse = await request(common.config.getDirektivHost()).delete(`/api/namespaces/${testNamespace}`)
        expect(deleteResponse.statusCode).toEqual(404)
        expect(deleteResponse.body).toMatchObject({})
    })

    it(`should check for server logs on the recent namespace operations`, async () => {
        var logsResponse = await request(common.config.getDirektivHost()).get(`/api/logs?order.field=TIMESTAMP&order.direction=DESC&limit=2`)
        expect(logsResponse.statusCode).toEqual(200)
        expect(logsResponse.body.results).toEqual(expect.arrayContaining([
            {
                level: "info",
                t: expect.anything(),
                msg: `Deleted namespace '${testNamespace}'.`,
                tags: expect.anything()
            },
            {
                level: "info",
                t: expect.anything(),
                msg: `Created namespace '${testNamespace}'.`,
                tags: expect.anything()
            }]))
    })
})
