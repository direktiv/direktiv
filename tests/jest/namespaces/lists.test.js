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
                    oid: "",
                    createdAt: expect.stringMatching(common.regex.timestampRegex),
                    updatedAt: expect.stringMatching(common.regex.timestampRegex),
                }
            })
        }
    })

    it(`should ensure default ordering is alphabetical`, async () => {
        var listResponse = await request(common.config.getDirektivHost()).get(`/api/namespaces`)
        expect(listResponse.statusCode).toEqual(200)
        expect(listResponse.body).toMatchObject({
            pageInfo: {
                order: [],
                filter: [],
                limit: 0,
                offset: 0,
                total: 10,
            },
            results: expect.anything(),
        })

        var expected = [...namespaceNames]
        expected.sort()

        for (let i = 0; i < listResponse.body.results.length; i++) {
            expect(listResponse.body.results[i]).toMatchObject({
                name: expected[i],
                oid: "",
                createdAt: expect.stringMatching(common.regex.timestampRegex),
                updatedAt: expect.stringMatching(common.regex.timestampRegex),
            })
        }
    })

    it(`should test NAME ASC ordering`, async () => {
        var listResponse = await request(common.config.getDirektivHost()).get(`/api/namespaces?order.field=NAME&order.direction=ASC`)
        expect(listResponse.statusCode).toEqual(200)
        expect(listResponse.body).toMatchObject({
            pageInfo: {
                order: [{
                    direction: "ASC",
                    field: "NAME",
                }],
                filter: [],
                limit: 0,
                offset: 0,
                total: 10,
            },
            results: expect.anything(),
        })

        var expected = [...namespaceNames]
        expected.sort()

        for (let i = 0; i < listResponse.body.results.length; i++) {
            expect(listResponse.body.results[i]).toMatchObject({
                name: expected[i],
                oid: "",
                createdAt: expect.stringMatching(common.regex.timestampRegex),
                updatedAt: expect.stringMatching(common.regex.timestampRegex),
            })
        }
    })

    it(`should test NAME DESC ordering`, async () => {
        var listResponse = await request(common.config.getDirektivHost()).get(`/api/namespaces?order.field=NAME&order.direction=DESC`)
        expect(listResponse.statusCode).toEqual(200)
        expect(listResponse.body).toMatchObject({
            pageInfo: {
                order: [{
                    direction: "DESC",
                    field: "NAME",
                }],
                filter: [],
                limit: 0,
                offset: 0,
                total: 10,
            },
            results: expect.anything(),
        })

        var expected = [...namespaceNames]
        expected.sort()
        expected.reverse()

        for (let i = 0; i < listResponse.body.results.length; i++) {
            expect(listResponse.body.results[i]).toMatchObject({
                name: expected[i],
                oid: "",
                createdAt: expect.stringMatching(common.regex.timestampRegex),
                updatedAt: expect.stringMatching(common.regex.timestampRegex),
            })
        }
    })

    it(`should test default does not perform pagination`, async () => {
        var listResponse = await request(common.config.getDirektivHost()).get(`/api/namespaces`)
        expect(listResponse.statusCode).toEqual(200)
        expect(listResponse.body).toMatchObject({
            pageInfo: {
                order: [],
                filter: [],
                limit: 0,
                offset: 0,
                total: 10,
            },
            results: expect.anything(),
        })

        var expected = [...namespaceNames]
        expected.sort()

        for (let i = 0; i < listResponse.body.results.length; i++) {
            expect(listResponse.body.results[i]).toMatchObject({
                name: expected[i],
                oid: "",
                createdAt: expect.stringMatching(common.regex.timestampRegex),
                updatedAt: expect.stringMatching(common.regex.timestampRegex),
            })
        }
    })

    it(`should test getting the first page`, async () => {
        var listResponse = await request(common.config.getDirektivHost()).get(`/api/namespaces?limit=4`)
        expect(listResponse.statusCode).toEqual(200)
        expect(listResponse.body).toMatchObject({
            pageInfo: {
                order: [],
                filter: [],
                limit: 4,
                offset: 0,
                total: 10,
            },
            results: expect.anything(),
        })

        var alphabetical = [...namespaceNames]
        alphabetical.sort()
        var expected = alphabetical.slice(0, 4)

        for (let i = 0; i < listResponse.body.results.length; i++) {
            expect(listResponse.body.results[i]).toMatchObject({
                name: expected[i],
                oid: "",
                createdAt: expect.stringMatching(common.regex.timestampRegex),
                updatedAt: expect.stringMatching(common.regex.timestampRegex),
            })
        }
    })

    it(`should test getting the second page`, async () => {
        var listResponse = await request(common.config.getDirektivHost()).get(`/api/namespaces?limit=4&offset=4`)
        expect(listResponse.statusCode).toEqual(200)
        expect(listResponse.body).toMatchObject({
            pageInfo: {
                order: [],
                filter: [],
                limit: 4,
                offset: 4,
                total: 10,
            },
            results: expect.anything(),
        })

        var alphabetical = [...namespaceNames]
        alphabetical.sort()
        var expected = alphabetical.slice(4, 8)

        for (let i = 0; i < listResponse.body.results.length; i++) {
            expect(listResponse.body.results[i]).toMatchObject({
                name: expected[i],
                oid: "",
                createdAt: expect.stringMatching(common.regex.timestampRegex),
                updatedAt: expect.stringMatching(common.regex.timestampRegex),
            })
        }
    })

    it(`should test getting the final page`, async () => {
        var listResponse = await request(common.config.getDirektivHost()).get(`/api/namespaces?limit=4&offset=8`)
        expect(listResponse.statusCode).toEqual(200)
        expect(listResponse.body).toMatchObject({
            pageInfo: {
                order: [],
                filter: [],
                limit: 4,
                offset: 8,
                total: 10,
            },
            results: expect.anything(),
        })

        var alphabetical = [...namespaceNames]
        alphabetical.sort()
        var expected = alphabetical.slice(8, 10)

        for (let i = 0; i < listResponse.body.results.length; i++) {
            expect(listResponse.body.results[i]).toMatchObject({
                name: expected[i],
                oid: "",
                createdAt: expect.stringMatching(common.regex.timestampRegex),
                updatedAt: expect.stringMatching(common.regex.timestampRegex),
            })
        }
    })

    it(`should test paginating out of bounds`, async () => {
        var listResponse = await request(common.config.getDirektivHost()).get(`/api/namespaces?limit=4&offset=12`)
        expect(listResponse.statusCode).toEqual(200)
        expect(listResponse.body).toMatchObject({
            pageInfo: {
                order: [],
                filter: [],
                limit: 4,
                offset: 12,
                total: 10,
            },
            results: [],
        })
    })

    it(`should test NAME CONTAINS filter`, async () => {
        var listResponse = await request(common.config.getDirektivHost()).get(`/api/namespaces?filter.field=NAME&filter.type=CONTAINS&filter.val=at`)
        expect(listResponse.statusCode).toEqual(200)
        expect(listResponse.body).toMatchObject({
            pageInfo: {
                order: [],
                filter: expect.arrayContaining([
                    {
                        type: "CONTAINS",
                        field: "NAME",
                        val: "at",
                    },
                ]),
                limit: 0,
                offset: 0,
                total: 2,
            },
            results: expect.arrayContaining([
                {
                    name: "at",
                    oid: "",
                    createdAt: expect.stringMatching(common.regex.timestampRegex),
                    updatedAt: expect.stringMatching(common.regex.timestampRegex),
                },
                {
                    name: "that",
                    oid: "",
                    createdAt: expect.stringMatching(common.regex.timestampRegex),
                    updatedAt: expect.stringMatching(common.regex.timestampRegex),
                },
            ]),
        })
    })
})
