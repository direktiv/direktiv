// http://192.168.0.145/api/namespaces/test/broadcast

import request from 'supertest'

import common from "../common"

const namespaceName = "sendevents"

const eventWithNonJSON = `{
    "specversion" : "1.0",
    "type" : "testerXML",
    "source" : "https://direktiv.io/test",
    "datacontenttype" : "text/xml",
    "data" : "<data>DATA</data>"
}`

const eventWithJSON = `{
    "specversion" : "1.0",
    "type" : "testerJSON",
    "source" : "https://direktiv.io/test",
    "datacontenttype" : "application/json",
    "data" : {
        "hello": "world",
        "123": 456
    }
}`

const eventDuplicate = `{
    "specversion" : "1.0",
    "type" : "testerDuplicate",
    "source" : "https://direktiv.io/test",
    "id": "123"
}`

describe('Test send events', () => {
    beforeAll(common.helpers.deleteAllNamespaces)
    afterAll(common.helpers.deleteAllNamespaces)

    it(`should create namespace`, async () => {
        var createNamespaceResponse = await request(common.config.getDirektivHost()).put(`/api/namespaces/${namespaceName}`)
        expect(createNamespaceResponse.statusCode).toEqual(200)
    })

    it(`should send event to namespace`, async () => {
        var workflowEventResponse = await request(common.config.getDirektivHost()).post(`/api/namespaces/${namespaceName}/broadcast`)
            .set('Content-Type', 'application/json')
            .send(eventDuplicate)
        expect(workflowEventResponse.statusCode).toEqual(200)
    })

    it(`fails with duplicate id`, async () => {
        var workflowEventResponse = await request(common.config.getDirektivHost()).post(`/api/namespaces/${namespaceName}/broadcast`)
            .set('Content-Type', 'application/json')
            .send(eventDuplicate)
        expect(workflowEventResponse.statusCode).toEqual(409)
    })

    it(`should send event to namespace with JSON`, async () => {
        var workflowEventResponse = await request(common.config.getDirektivHost()).post(`/api/namespaces/${namespaceName}/broadcast`)
            .set('Content-Type', 'application/json')
            .send(eventWithJSON)
        expect(workflowEventResponse.statusCode).toEqual(200)
    })

    it(`should send event to namespace with non-JSON`, async () => {
        var workflowEventResponse = await request(common.config.getDirektivHost()).post(`/api/namespaces/${namespaceName}/broadcast`)
            .set('Content-Type', 'application/json')
            .send(eventWithNonJSON)
        expect(workflowEventResponse.statusCode).toEqual(200)
    })

    it(`should send event as non-compliant`, async () => {
        var workflowEventResponse = await request(common.config.getDirektivHost()).post(`/api/namespaces/${namespaceName}/broadcast`)
            .set('Content-Type', 'application/json')
            .send("NON-COMPLIANT")
        expect(workflowEventResponse.statusCode).toEqual(200)
    })

    it(`should list events`, async () => {
        var workflowEventResponse = await request(common.config.getDirektivHost()).get(`/api/namespaces/${namespaceName}/events?limit=10&offset=0`)
            .send()


        expect(workflowEventResponse.statusCode).toEqual(200)

        // test there are the four created
        expect(workflowEventResponse.body.events.pageInfo.total).toEqual(4)

        // test that all types are in
        expect(workflowEventResponse.body.events.results).toEqual(
            expect.arrayContaining([
                expect.objectContaining({
                    type: 'noncompliant'
                })
            ])
        )


        expect(workflowEventResponse.body.events.results).toEqual(
            expect.arrayContaining([
                expect.objectContaining({
                    type: 'testerDuplicate'
                })
            ])
        )

        expect(workflowEventResponse.body.events.results).toEqual(
            expect.arrayContaining([
                expect.objectContaining({
                    type: 'testerXML'
                })
            ])
        )

        expect(workflowEventResponse.body.events.results).toEqual(
            expect.arrayContaining([
                expect.objectContaining({
                    type: 'testerJSON'
                })
            ])
        )

    })

})