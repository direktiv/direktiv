// http://192.168.0.145/api/namespaces/test/broadcast

import { beforeAll, describe, expect, it } from '@jest/globals'

import common from '../common'
import helpers from '../common/helpers'
import request from '../common/request'

const namespaceName = 'sendevents'

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

	helpers.itShouldCreateNamespace(it, expect, namespaceName)

	it(`should send event to namespace`, async () => {
		const workflowEventResponse = await request(common.config.getDirektivHost()).post(`/api/v2/namespaces/${namespaceName}/events/broadcast`)
			.set('Content-Type', 'application/json')
			.send(eventDuplicate)
		expect(workflowEventResponse.statusCode).toEqual(200)
	})

	it(`fails with duplicate id`, async () => {
		const broadcastResponse = await request(common.config.getDirektivHost()).post(`/api/v2/namespaces/${namespaceName}/events/broadcast`)
			.set('Content-Type', 'application/json')
			.send(eventDuplicate)
		const historyResponse = await request(common.config.getDirektivHost()).get(`/api/v2/namespaces/${namespaceName}/events/history?limit=10&offset=0`)
			.send()
		expect(broadcastResponse.statusCode).toEqual(400)
		expect(historyResponse.statusCode).toEqual(200)
		expect(historyResponse.body.data.length).toEqual(1)
	})

	it(`should send event to namespace with JSON`, async () => {
		const broadcastResponse = await request(common.config.getDirektivHost()).post(`/api/v2/namespaces/${namespaceName}/events/broadcast`)
			.set('Content-Type', 'application/json')
			.send(eventWithJSON)
		const workflowEventListResponse = await request(common.config.getDirektivHost()).get(`/api/v2/namespaces/${namespaceName}/events/history?limit=10&offset=0`)
			.send()
		expect(broadcastResponse.statusCode).toEqual(200)
		expect(workflowEventListResponse.statusCode).toEqual(200)
		expect(workflowEventListResponse.body.data.length).toEqual(2)
	})

	it(`should send event to namespace with non-JSON`, async () => {
		const broadcastResponse = await request(common.config.getDirektivHost()).post(`/api/v2/namespaces/${namespaceName}/events/broadcast`)
			.set('Content-Type', 'application/json')
			.send(eventWithNonJSON)
		const historyResponse = await request(common.config.getDirektivHost()).get(`/api/v2/namespaces/${namespaceName}/events/history?limit=10&offset=0`)
			.send()
		expect(broadcastResponse.statusCode).toEqual(200)
		expect(historyResponse.statusCode).toEqual(200)
		expect(historyResponse.body.data.length).toEqual(3)
	})

	it(`should send event as non-compliant`, async () => {
		const broadcastResponse = await request(common.config.getDirektivHost()).post(`/api/v2/namespaces/${namespaceName}/events/broadcast`)
			.set('Content-Type', 'application/json')
			.send('NON-COMPLIANT')
		const workflowEventListResponse = await request(common.config.getDirektivHost()).get(`/api/v2/namespaces/${namespaceName}/events/history?limit=10&offset=0`)
			.send()
		expect(broadcastResponse.statusCode).toEqual(400)
		expect(workflowEventListResponse.statusCode).toEqual(200)
		expect(workflowEventListResponse.body.data.length).toEqual(3)
	})

	it(`should list events`, async () => {
		const historyResponse = await request(common.config.getDirektivHost()).get(`/api/v2/namespaces/${namespaceName}/events/history?limit=10&offset=0`)
			.send()
		expect(historyResponse.statusCode).toEqual(200)
		expect(historyResponse.body.data.length).toEqual(3)
		expect(historyResponse.body.data.find(item => item.event.type === 'testerDuplicate')).not.toBeFalsy()
		expect(historyResponse.body.data.find(item => item.event.type === 'testerXML')).not.toBeFalsy()
		expect(historyResponse.body.data.find(item => item.event.type === 'testerJSON')).not.toBeFalsy()
	})

	it(`bad filter value applied on the eventlog`, async () => {
		// &filter.field=TEXT&filter.type=CONTAINS&filter.val=dfda&filter.field=CREATED&filter.type=AFTER&filter.val=2023-07-11T22%3A00%3A00.000Z
		const historyResponse = await request(common.config.getDirektivHost()).get(`/api/v2/namespaces/${namespaceName}/events/history?limit=10&offset=0&eventContains=dfda`)
			.send()

		expect(historyResponse.statusCode).toEqual(200)

		expect(historyResponse.body.data.length).toEqual(0)

		expect(historyResponse.body.data).toEqual(
			expect.arrayContaining([]),
		)
	})

	it(`should filter the eventlog by TEXT`, async () => {
		const historyResponse = await request(common.config.getDirektivHost()).get(`/api/v2/namespaces/${namespaceName}/events/history?limit=10&offset=0&filter.field=TEXT&eventContains=world`)
			.send()

		expect(historyResponse.statusCode).toEqual(200)

		// test there are the four created
		expect(historyResponse.body.data.length).toEqual(1)
		expect(historyResponse.body.data.find(item => item.event.type === 'testerJSON')).not.toBeFalsy()
	})
	it(`should filter the eventlog by TYPE`, async () => {
		const historyResponse = await request(common.config.getDirektivHost()).get(`/api/v2/namespaces/${namespaceName}/events/history?limit=10&offset=0&filter.field=TYPE&typeContains=testerJSON`)
			.send()

		expect(historyResponse.statusCode).toEqual(200)

		// test there are the four created
		expect(historyResponse.body.data.length).toEqual(1)
		expect(historyResponse.body.data.find(item => item.event.type === 'testerJSON')).not.toBeFalsy()
		expect(historyResponse.body.data.find(item => item.event.type === 'testerJSON')).not.toBeFalsy()
	})
})
