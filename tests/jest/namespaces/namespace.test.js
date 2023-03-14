import request from 'supertest'

import common from "../common"

const createNamespaceResponse = {
    namespace: common.structs.namespaceObject,
}

const deleteNamespaceResponse = {}

const listNamespacesResponse = {
    pageInfo: common.structs.pageInfoObject,
    results: expect.anything(),
}

// NOTE: no need to test get namespace. It's not yet called by the API.
// NOTE: no need to test rename. It's not yet called by the API.
// TODO: test 404 from a missing namespace indirectly (tree, logs, etc)
// TODO: test recursive argument 
// TODO: test SSE
// TODO: test bad method
// TODO: test namespace logs 
// TODO: test namespace config
// TODO: test namespace annotations

describe('Test basic namespace operations', () => {
    it(`should create, get, and delete a namespace`, async () => {
        const name = "a"
        await request(common.config.getDirektivHost()).delete(`/api/namespaces/${name}?recursive=true`)
        const createResponse = await request(common.config.getDirektivHost()).put(`/api/namespaces/${name}`)
        const deleteResponse = await request(common.config.getDirektivHost()).delete(`/api/namespaces/${name}`)
        request(common.config.getDirektivHost()).delete(`/api/namespaces/${name}?recursive=true`)

        expect(createResponse.statusCode).toEqual(200)
        expect(createResponse.body).toMatchObject(createNamespaceResponse)
        expect(createResponse.body.namespace.name).toBe(name)

        expect(deleteResponse.statusCode).toEqual(200)
        expect(deleteResponse.body).toMatchObject(deleteNamespaceResponse)
    })

    it("should attempt to delete a namespace that doesn't exist", async () => {
        const name = "a"
        request(common.config.getDirektivHost()).delete(`/api/namespaces/${name}?recursive=true`)
        const deleteResponse = await request(common.config.getDirektivHost()).delete(`/api/namespaces/${name}`)
        expect(deleteResponse.statusCode).toEqual(404)
        expect(deleteResponse.body).toMatchObject(deleteNamespaceResponse)
    })

    it(`should attempt to create a namespace that already exists`, async () => {
        const name = "a"
        await request(common.config.getDirektivHost()).delete(`/api/namespaces/${name}?recursive=true`)
        const res1 = await request(common.config.getDirektivHost()).put(`/api/namespaces/${name}`)
        const res2 = await request(common.config.getDirektivHost()).put(`/api/namespaces/${name}`)
        request(common.config.getDirektivHost()).delete(`/api/namespaces/${name}?recursive=true`)

        expect(res1.statusCode).toEqual(200)
        expect(res1.body).toMatchObject(createNamespaceResponse)
        expect(res1.body.namespace.name).toBe(name)

        expect(res2.statusCode).toEqual(409)
        expect(res2.body).toMatchObject(common.structs.errorResponse)
        expect(res2.body.code).toBe(409)
        expect(res2.body.message).toBe("resource already exists")
    })

    it(`should list namespaces`, async () => {
        const initListResponse = await request(common.config.getDirektivHost()).get(`/api/namespaces`)
        expect(initListResponse.statusCode).toEqual(200)
        expect(initListResponse.body).toMatchObject(listNamespacesResponse)

        for (let i = 0; i < initListResponse.body.results.length; i++) {
            var name = initListResponse.body.results[i].name
            await request(common.config.getDirektivHost()).delete(`/api/namespaces/${name}?recursive=true`)
        }

        const names = ["a", "b", "c"]
        for (let i = 0; i < names.length; i++) {
            var l = i + 1
            var name = names[i]
            var createResponse = await request(common.config.getDirektivHost()).put(`/api/namespaces/${name}`)
            var listResponse = await request(common.config.getDirektivHost()).get(`/api/namespaces`)

            expect(createResponse.statusCode).toEqual(200)
            expect(createResponse.body).toMatchObject(createNamespaceResponse)
            expect(createResponse.body.namespace.name).toBe(name)

            expect(listResponse.statusCode).toEqual(200)
            expect(listResponse.body).toMatchObject(listNamespacesResponse)
            expect(listResponse.body.pageInfo.order).toHaveLength(0)
            expect(listResponse.body.pageInfo.filter).toHaveLength(0)
            expect(listResponse.body.pageInfo.limit).toEqual(0)
            expect(listResponse.body.pageInfo.offset).toEqual(0)
            expect(listResponse.body.pageInfo.total).toEqual(l)

            for (let j = 0; j < listResponse.body.results.length; j++) {
                expect(listResponse.body.results[j].name).toEqual(names[j])
            }
        }

        for (let i = 0; i < names.length; i++) {
            var name = names[i]
            await request(common.config.getDirektivHost()).delete(`/api/namespaces/${name}?recursive=true`)
        }
    })

    it(`should test pagination`, async () => {
        const initListResponse = await request(common.config.getDirektivHost()).get(`/api/namespaces`)
        expect(initListResponse.statusCode).toEqual(200)
        expect(initListResponse.body).toMatchObject(listNamespacesResponse)

        for (let i = 0; i < initListResponse.body.results.length; i++) {
            var name = initListResponse.body.results[i].name
            await request(common.config.getDirektivHost()).delete(`/api/namespaces/${name}?recursive=true`)
        }

        const names = ["a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", 
            "m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z"]
        for (let i = 0; i < names.length; i++) {
            var name = names[i]
            var createResponse = await request(common.config.getDirektivHost()).put(`/api/namespaces/${name}`)  
            expect(createResponse.statusCode).toEqual(200)
            expect(createResponse.body).toMatchObject(createNamespaceResponse)
            expect(createResponse.body.namespace.name).toBe(name)
        }

        var listResponse = await request(common.config.getDirektivHost()).get(`/api/namespaces`)
        expect(listResponse.statusCode).toEqual(200)
        expect(listResponse.body).toMatchObject(listNamespacesResponse)
        expect(listResponse.body.pageInfo.order).toHaveLength(0)
        expect(listResponse.body.pageInfo.filter).toHaveLength(0)
        expect(listResponse.body.pageInfo.limit).toEqual(0)
        expect(listResponse.body.pageInfo.offset).toEqual(0)
        expect(listResponse.body.pageInfo.total).toEqual(26)
        for (let i = 0; i < listResponse.body.results.length; i++) {
            expect(listResponse.body.results[i].name).toEqual(names[i])
        }

        var listResponse = await request(common.config.getDirektivHost()).get(`/api/namespaces?limit=10`)
        expect(listResponse.statusCode).toEqual(200)
        expect(listResponse.body).toMatchObject(listNamespacesResponse)
        expect(listResponse.body.pageInfo.order).toHaveLength(0)
        expect(listResponse.body.pageInfo.filter).toHaveLength(0)
        expect(listResponse.body.pageInfo.limit).toEqual(10)
        expect(listResponse.body.pageInfo.offset).toEqual(0)
        expect(listResponse.body.pageInfo.total).toEqual(26)
        expect(listResponse.body.results.length).toEqual(10)
        for (let i = 0; i < listResponse.body.results.length; i++) {
            expect(listResponse.body.results[i].name).toEqual(names[i])
        }

        var listResponse = await request(common.config.getDirektivHost()).get(`/api/namespaces?limit=10&offset=10`)
        expect(listResponse.statusCode).toEqual(200)
        expect(listResponse.body).toMatchObject(listNamespacesResponse)
        expect(listResponse.body.pageInfo.order).toHaveLength(0)
        expect(listResponse.body.pageInfo.filter).toHaveLength(0)
        expect(listResponse.body.pageInfo.limit).toEqual(10)
        expect(listResponse.body.pageInfo.offset).toEqual(10)
        expect(listResponse.body.pageInfo.total).toEqual(26)
        expect(listResponse.body.results.length).toEqual(10)
        for (let i = 0; i < listResponse.body.results.length; i++) {
            expect(listResponse.body.results[i].name).toEqual(names[i+10])
        }

        var listResponse = await request(common.config.getDirektivHost()).get(`/api/namespaces?limit=10&offset=20`)
        expect(listResponse.statusCode).toEqual(200)
        expect(listResponse.body).toMatchObject(listNamespacesResponse)
        expect(listResponse.body.pageInfo.order).toHaveLength(0)
        expect(listResponse.body.pageInfo.filter).toHaveLength(0)
        expect(listResponse.body.pageInfo.limit).toEqual(10)
        expect(listResponse.body.pageInfo.offset).toEqual(20)
        expect(listResponse.body.pageInfo.total).toEqual(26)
        expect(listResponse.body.results.length).toEqual(6)
        for (let i = 0; i < listResponse.body.results.length; i++) {
            expect(listResponse.body.results[i].name).toEqual(names[i+20])
        }

        var listResponse = await request(common.config.getDirektivHost()).get(`/api/namespaces?limit=10&offset=30`)
        expect(listResponse.statusCode).toEqual(200)
        expect(listResponse.body).toMatchObject(listNamespacesResponse)
        expect(listResponse.body.pageInfo.order).toHaveLength(0)
        expect(listResponse.body.pageInfo.filter).toHaveLength(0)
        expect(listResponse.body.pageInfo.limit).toEqual(10)
        expect(listResponse.body.pageInfo.offset).toEqual(30)
        expect(listResponse.body.pageInfo.total).toEqual(26)
        expect(listResponse.body.results.length).toEqual(0)

        for (let i = 0; i < names.length; i++) {
            var name = names[i]
            await request(common.config.getDirektivHost()).delete(`/api/namespaces/${name}?recursive=true`)
        }
    })

    it(`should test valid orderings`, async () => {
        const initListResponse = await request(common.config.getDirektivHost()).get(`/api/namespaces`)
        expect(initListResponse.statusCode).toEqual(200)
        expect(initListResponse.body).toMatchObject(listNamespacesResponse)

        for (let i = 0; i < initListResponse.body.results.length; i++) {
            var name = initListResponse.body.results[i].name
            await request(common.config.getDirektivHost()).delete(`/api/namespaces/${name}?recursive=true`)
        }

        const names = ["s", "t", "u", "v", "w", "x", "y", "z", "a", "b", "c", 
            "d", "e", "f", "m", "n", "o", "p", "q", "r", "g", "h", "i", "j", "k", "l"]
        const alphabetical_names = ["a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", 
            "l", "m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z"]
        for (let i = 0; i < names.length; i++) {
            var name = names[i]
            var createResponse = await request(common.config.getDirektivHost()).put(`/api/namespaces/${name}`)  
            expect(createResponse.statusCode).toEqual(200)
            expect(createResponse.body).toMatchObject(createNamespaceResponse)
            expect(createResponse.body.namespace.name).toBe(name)
        }

        var listResponse = await request(common.config.getDirektivHost()).get(`/api/namespaces`)
        expect(listResponse.statusCode).toEqual(200)
        expect(listResponse.body).toMatchObject(listNamespacesResponse)
        expect(listResponse.body.pageInfo.order).toHaveLength(0)
        expect(listResponse.body.pageInfo.filter).toHaveLength(0)
        expect(listResponse.body.pageInfo.limit).toEqual(0)
        expect(listResponse.body.pageInfo.offset).toEqual(0)
        expect(listResponse.body.pageInfo.total).toEqual(26)
        for (let i = 0; i < listResponse.body.results.length; i++) {
            expect(listResponse.body.results[i].name).toEqual(alphabetical_names[i])
        }

        var listResponse = await request(common.config.getDirektivHost()).get(`/api/namespaces?order.field=NAME&order.direction=ASC`)
        expect(listResponse.statusCode).toEqual(200)
        expect(listResponse.body).toMatchObject(listNamespacesResponse)
        expect(listResponse.body.pageInfo.order).toHaveLength(1)
        expect(listResponse.body.pageInfo.order[0].direction).toEqual("ASC")
        expect(listResponse.body.pageInfo.order[0].field).toEqual("NAME")
        expect(listResponse.body.pageInfo.filter).toHaveLength(0)
        expect(listResponse.body.pageInfo.limit).toEqual(0)
        expect(listResponse.body.pageInfo.offset).toEqual(0)
        expect(listResponse.body.pageInfo.total).toEqual(26)
        for (let i = 0; i < listResponse.body.results.length; i++) {
            expect(listResponse.body.results[i].name).toEqual(alphabetical_names[i])
        }

        var listResponse = await request(common.config.getDirektivHost()).get(`/api/namespaces?order.field=NAME&order.direction=DESC`)
        expect(listResponse.statusCode).toEqual(200)
        expect(listResponse.body).toMatchObject(listNamespacesResponse)
        expect(listResponse.body.pageInfo.order).toHaveLength(1)
        expect(listResponse.body.pageInfo.order[0].direction).toEqual("DESC")
        expect(listResponse.body.pageInfo.order[0].field).toEqual("NAME")
        expect(listResponse.body.pageInfo.filter).toHaveLength(0)
        expect(listResponse.body.pageInfo.limit).toEqual(0)
        expect(listResponse.body.pageInfo.offset).toEqual(0)
        expect(listResponse.body.pageInfo.total).toEqual(26)
        for (let i = 0; i < listResponse.body.results.length; i++) {
            expect(listResponse.body.results[i].name).toEqual(alphabetical_names[names.length-1-i])
        }

        for (let i = 0; i < names.length; i++) {
            var name = names[i]
            await request(common.config.getDirektivHost()).delete(`/api/namespaces/${name}?recursive=true`)
        }
    })

    it(`should test valid orderings`, async () => {
        const initListResponse = await request(common.config.getDirektivHost()).get(`/api/namespaces`)
        expect(initListResponse.statusCode).toEqual(200)
        expect(initListResponse.body).toMatchObject(listNamespacesResponse)

        for (let i = 0; i < initListResponse.body.results.length; i++) {
            var name = initListResponse.body.results[i].name
            await request(common.config.getDirektivHost()).delete(`/api/namespaces/${name}?recursive=true`)
        }

        const names = ["the", "be", "to", "of", "and", "a", "in", "that", "have", "at"]
        for (let i = 0; i < names.length; i++) {
            var name = names[i]
            var createResponse = await request(common.config.getDirektivHost()).put(`/api/namespaces/${name}`)  
            expect(createResponse.statusCode).toEqual(200)
            expect(createResponse.body).toMatchObject(createNamespaceResponse)
            expect(createResponse.body.namespace.name).toBe(name)
        }

        var listResponse = await request(common.config.getDirektivHost()).get(`/api/namespaces?filter.field=NAME&filter.type=CONTAINS&filter.val=at`)
        expect(listResponse.statusCode).toEqual(200)
        expect(listResponse.body).toMatchObject(listNamespacesResponse)
        expect(listResponse.body.pageInfo.order).toHaveLength(0)
        expect(listResponse.body.pageInfo.filter).toHaveLength(1)
        expect(listResponse.body.pageInfo.filter[0].type).toEqual("CONTAINS")
        expect(listResponse.body.pageInfo.filter[0].field).toEqual("NAME")
        expect(listResponse.body.pageInfo.filter[0].val).toEqual("at")
        expect(listResponse.body.pageInfo.limit).toEqual(0)
        expect(listResponse.body.pageInfo.offset).toEqual(0)
        expect(listResponse.body.pageInfo.total).toEqual(2)
        var expectedResults = ["at", "that"]
        for (let i = 0; i < listResponse.body.results.length; i++) {
            expect(listResponse.body.results[i].name).toEqual(expectedResults[i])
        }

        for (let i = 0; i < names.length; i++) {
            var name = names[i]
            await request(common.config.getDirektivHost()).delete(`/api/namespaces/${name}?recursive=true`)
        }
    })

    it(`should check for server logs on basic namespace operations`, async () => {
        const name = "nslogstest"
        await request(common.config.getDirektivHost()).delete(`/api/namespaces/${name}?recursive=true`)
        const createResponse = await request(common.config.getDirektivHost()).put(`/api/namespaces/${name}`)
        const deleteResponse = await request(common.config.getDirektivHost()).delete(`/api/namespaces/${name}`)
        request(common.config.getDirektivHost()).delete(`/api/namespaces/${name}?recursive=true`)
        expect(createResponse.statusCode).toEqual(200)
        expect(deleteResponse.statusCode).toEqual(200)

        var logsResponse = await request(common.config.getDirektivHost()).get(`/api/logs?order.field=TIMESTAMP&order.direction=DESC&limit=2`)
        expect(logsResponse.statusCode).toEqual(200)
        expect(logsResponse.body.results).toEqual(expect.arrayContaining([{t: expect.anything(), msg: `Created namespace '${name}'.`}]))
        expect(logsResponse.body.results).toEqual(expect.arrayContaining([{t: expect.anything(), msg: `Deleted namespace '${name}'.`}]))
    })
})
