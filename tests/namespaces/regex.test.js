import request from "../common/request"

import common from "../common"

const createNamespaceResponse = {
    namespace: common.structs.namespaceObject,
}

describe('Test namespace regex constraints', () => {
    beforeEach(common.helpers.deleteAllNamespaces)


    it(`should create namespaces with various valid names`, async () => {
        const names = [
            "a",
            "test-flow-namespace-regex-a",
            "test-flow-namespace-regex-1",
            "test-flow-namespace-regex-a_b.c",
        ]
        for (let i = 0; i < names.length; i++) {
            var name = names[i]
            const res = await request(common.config.getDirektivHost()).put(`/api/namespaces/${name}`)
            expect(res.statusCode).toEqual(200)
            expect(res.body).toMatchObject({
                namespace: {
                    name: name,
                    createdAt: expect.stringMatching(common.regex.timestampRegex),
                    updatedAt: expect.stringMatching(common.regex.timestampRegex),
                }
            })
        }
    })

    it(`should fail to create namespaces with various invalid names`, async () => {
        const names = [
            "test-flow-namespace-regex-A",
            "Test-flow-namespace-regex-a",
            "test-flow-namespace-reGex-a",
            "1test-flow-namespace-regex-a",
            ".test-flow-namespace-regex-a",
            "_test-flow-namespace-regex-a",
            "test-flow-namespace-regex-a_",
            "test-flow-namespace-regex-a.",
            // "test-flow-namespace/regex-a",
            "test-flow-namespace@regex-a",
            "test-flow-namespace+regex-a",
            // "test-flow-namespace%25regex-a",
            // "test-flow-namespace?regex-a",
            // "test-flow-namespace%3Fregex-a",
            "test-flow-namespace regex-a",
            "test-flow-namespace%20regex-a",
        ]
        for (let i = 0; i < names.length; i++) {
            var name = names[i]
            const res = await request(common.config.getDirektivHost()).put(`/api/namespaces/${name}`)
            expect(res.statusCode).toEqual(406)
        }
    })
})
