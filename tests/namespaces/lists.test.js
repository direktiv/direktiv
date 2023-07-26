import request from 'supertest'

import common from "../common"

const namespaceNames = ["the", "be", "to", "of", "and", "a", "in", "that", "have", "at"]

describe('Test namespace listing functionality', () => {
    beforeAll(common.helpers.deleteAllNamespaces)
    afterAll(common.helpers.deleteAllNamespaces)

    it(`should create a number of different namespaces`, async () => {
        for (let i = 0; i < namespaceNames.length; i++) {
            var name = namespaceNames[i]
            var createResponse = await request(common.config.getDirektivHost()).put(`/api/namespaces/${name}`)
            expect(createResponse.statusCode).toEqual(200)
            expect(createResponse.body).toMatchObject({
                namespace: {
                    name: name,
                    oid: expect.stringMatching(common.regex.uuidRegex),
                    createdAt: expect.stringMatching(common.regex.timestampRegex),
                    updatedAt: expect.stringMatching(common.regex.timestampRegex),
                }
            })
        }
    })

    it(`test get all list`, async () => {
        var listResponse = await request(common.config.getDirektivHost()).get(`/api/namespaces`)
        expect(listResponse.statusCode).toEqual(200)
        expect(listResponse.body).toMatchObject({
            pageInfo: null,
            results: expect.anything(),
        })

        var expected = [...namespaceNames]

        for (let i = 0; i < listResponse.body.results.length; i++) {
            expect(listResponse.body.results[i]).toMatchObject({
                name: expected[i],
                oid: expect.stringMatching(common.regex.uuidRegex),
                createdAt: expect.stringMatching(common.regex.timestampRegex),
                updatedAt: expect.stringMatching(common.regex.timestampRegex),
            })
        }
    })
})
