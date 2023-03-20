import request from 'supertest'

import common from "../common"

// NOTE: no need to test get namespace. It's not yet called by the API.
// NOTE: no need to test rename. It's not yet called by the API.
// TODO: test 404 from a missing namespace indirectly (tree, logs, etc)
// TODO: test recursive argument 
// TODO: test SSE
// TODO: test bad method
// TODO: test namespace logs 
// TODO: test namespace config
// TODO: test namespace annotations

const testNamespace = "a"

describe('Test basic namespace operation.', () => {
    beforeAll(common.helpers.deleteAllNamespaces)
    afterAll(common.helpers.deleteAllNamespaces)

    it(`should create a namespace`, async () => {
        const createResponse = await request(common.config.getDirektivHost()).put(`/api/namespaces/${testNamespace}`)
        expect(createResponse.statusCode).toEqual(200)
        expect(createResponse.body).toMatchObject({
            namespace: {
                name: testNamespace,
                oid: "",
                createdAt: expect.stringMatching(common.regex.timestampRegex),
                updatedAt: expect.stringMatching(common.regex.timestampRegex),
            }
        })
    })
})
