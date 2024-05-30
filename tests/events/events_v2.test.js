import { beforeAll, describe, expect, it } from '@jest/globals'

import common from '../common'
import helpers from '../common/helpers'
import request from '../common/request'

const namespaceName = 'wfeventsv2'

const basicEvent = `{
    "specversion" : "1.0",
    "type" : "testerDuplicate",
    "source" : "https://direktiv.io/test",
    "id": "123"
}`

let tmpid

describe('Test send events v2 api', () => {
	beforeAll(common.helpers.deleteAllNamespaces)
	helpers.itShouldCreateNamespace(it, expect, namespaceName)
	it(`should send event to namespace`, async () => {
		const sendEventResponse = await request(common.config.getDirektivHost()).post(`/api/v2/namespaces/${namespaceName}/events/broadcast`)
			.set('Content-Type', 'application/json')
			.send(basicEvent)
		expect(sendEventResponse.statusCode).toEqual(200)
	})
	it(`should not accept dup event`, async () => {
		const sendEventResponse = await request(common.config.getDirektivHost()).post(`/api/v2/namespaces/${namespaceName}/events/broadcast`)
			.set('Content-Type', 'application/cloudevents+json')
			.send(basicEvent)
		expect(sendEventResponse.statusCode).toEqual(400)
	})
	it(`should not break the server`, async () => {
		const sendEventResponse = await request(common.config.getDirektivHost()).post(`/api/v2/namespaces/${namespaceName}/events/broadcast`)
			.set('Content-Type', 'application/bad-header')
			.send(basicEvent)
		expect(sendEventResponse.statusCode).toEqual(415)
	})
	it(`should be regitered`, async () => {
		const eventHistoryResponse = await request(common.config.getDirektivHost()).get(`/api/v2/namespaces/${namespaceName}/events/history`)
			.send()
		expect(eventHistoryResponse.statusCode).toEqual(200)
		expect(eventHistoryResponse.body.data.length).toBeGreaterThan(0)
		expect(eventHistoryResponse.body.data[0].namespace).toBe(namespaceName)
		expect(eventHistoryResponse.body.data[0].event.id).toBe('123')
	})
	it(`event by id`, async () => {
		const eventHistoryResponse = await request(common.config.getDirektivHost()).get(`/api/v2/namespaces/${namespaceName}/events/history/123`)
			.send()
		expect(eventHistoryResponse.statusCode).toEqual(200)
	})
})

describe('Test basic workflow events v2', () => {
	beforeAll(common.helpers.deleteAllNamespaces)

	helpers.itShouldCreateNamespace(it, expect, namespaceName)

	common.helpers.itShouldCreateYamlFile(it, expect, namespaceName,
		'/', 'listener.yml', 'workflow', `
start:
  type: event
  event:
    type: greeting
  state: helloworld
states:
- id: helloworld
  type: noop
  transform:
    result: Hello world!
`)

	it(`should wait a second for the events logic to sync`, async () => {
		await helpers.sleep(1000)
	})

	it(`listener should be regitered`, async () => {
		const eventListenerResponse = await request(common.config.getDirektivHost()).get(`/api/v2/namespaces/${namespaceName}/events/listeners?limit=100&offset=0`)
			.send()
		expect(eventListenerResponse.statusCode).toEqual(200)
		expect(eventListenerResponse.body.data.length).toBeGreaterThan(0)
		expect(eventListenerResponse.body.data[0].triggerWorkflow).toBe('/listener.yml')
		expect(eventListenerResponse.body.data[0].namespace).toBe(namespaceName)
		tmpid = eventListenerResponse.body.data[0].id
	})
	it(`listener by id`, async () => {
		const eventListenerResponse = await request(common.config.getDirektivHost()).get(`/api/v2/namespaces/${namespaceName}/events/listeners/${tmpid}`)
			.send()
		expect(eventListenerResponse.statusCode).toEqual(200)
		expect(eventListenerResponse.body.data.id).toBe(tmpid)
	})
})
