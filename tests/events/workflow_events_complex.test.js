import { beforeAll, describe, expect, it } from '@jest/globals'

import common from '../common'
import helpers from '../common/helpers'
import request from '../common/request'
import events from './send_helper'

const namespaceName = 'wfevents-complex'

const startThenWaitWorkflowNameContext = 'startandwaitdyn.yaml'
const starthenWaitWorkflowContext = `
start:
  type: event
  state: ce
  event: 
    type: hello
states:
- id: ce
  type: consumeEvent
  log: jq(."hello".hello)
  event:
    type: hellowait
    context: 
      hello: jq(."hello".hello)
`

const basevent = (type, id, value) => `{
    "specversion" : "1.0",
    "type" : "${ type }",
    "id": "${ id }",
    "source" : "https://direktiv.io/test",
    "datacontenttype" : "application/json",
    "hello": "${ value }",
    "data" : {
        "hello": "world",
        "123": 456
    }
}`

describe('Test complex workflow events orchistration', () => {
	beforeAll(common.helpers.deleteAllNamespaces)

	it(`should create namespace`, async () => {
		const createNamespaceResponse = await request(common.config.getDirektivHost()).put(`/api/namespaces/${ namespaceName }`)
		expect(createNamespaceResponse.statusCode).toEqual(200)
	})

	helpers.itShouldCreateYamlFileV2(it, expect, namespaceName,
		'', startThenWaitWorkflowNameContext, 'workflow',
		starthenWaitWorkflowContext)

	it(`multiple event-streams`, async () => {
		await helpers.sleep(1000)
		const eventStream1 = basevent('hello', 'wait-ctx5', 'condition1')
		const eventStream1Stage2 = basevent('hellowait', 'wait-ctx-run5', 'condition1')
		const eventStream2 = basevent('hello', 'wait-ctx6', 'condition2')
		const eventStream2Stage2 = basevent('hellowait', 'wait-ctx-run53', 'condition2')
		const eventStream3 = basevent('hello', 'wait-ctx566', 'condition3')
		const eventStream3Stage2 = basevent('hellowait', 'wait-ctx-run43', 'condition3')

		await events.sendEventAndList(namespaceName, eventStream1)
		let instancesResponse = await request(common.config.getDirektivHost()).get(`/api/namespaces/${ namespaceName }/instances?limit=10&offset=0`)
			.send()
		expect(instancesResponse.body.instances.results.length).toBe(1)
		const stream1InstanceId = instancesResponse.body.instances.results[0].id

		await events.sendEventAndList(namespaceName, eventStream2)
		instancesResponse = await request(common.config.getDirektivHost()).get(`/api/namespaces/${ namespaceName }/instances?limit=10&offset=0`)
			.send()
		const stream2InstanceId = instancesResponse.body.instances.results[0].id // assuming they are sorted
		expect(instancesResponse.body.instances.results.length).toBe(2)
		expect(stream1InstanceId).not.toBe(stream2InstanceId)

		await events.sendEventAndList(namespaceName, eventStream3)
		instancesResponse = await request(common.config.getDirektivHost()).get(`/api/namespaces/${ namespaceName }/instances?limit=10&offset=0`)
			.send()
		const stream3InstanceId = instancesResponse.body.instances.results[0].id // assuming they are sorted
		expect(instancesResponse.body.instances.results.length).toBe(3)
		expect(stream3InstanceId).not.toBe(stream2InstanceId)

		await events.sendEventAndList(namespaceName, eventStream1Stage2)
		await helpers.sleep(300)

		const statusStream1 = await request(common.config.getDirektivHost()).get(`/api/namespaces/${ namespaceName }/instances/${ stream1InstanceId }`)
			.send()
		expect(statusStream1.body.instance.status).toBe('complete')

		let statusStream2 = await request(common.config.getDirektivHost()).get(`/api/namespaces/${ namespaceName }/instances/${ stream2InstanceId }`)
			.send()
		expect(statusStream2.body.instance.status).toBe('pending')

		let statusStream3 = await request(common.config.getDirektivHost()).get(`/api/namespaces/${ namespaceName }/instances/${ stream3InstanceId }`)
			.send()
		expect(statusStream3.body.instance.status).toBe('pending')

		const resultsStream1 = await request(common.config.getDirektivHost()).get(`/api/namespaces/${ namespaceName }/instances/${ stream1InstanceId }/output`)
			.send()

		const outputData1 = JSON.parse(atob(resultsStream1.body.data))
		expect(outputData1.hello.hello).toBe('condition1')

		await events.sendEventAndList(namespaceName, eventStream2Stage2)
		await helpers.sleep(300)

		const resultsStream2 = await request(common.config.getDirektivHost()).get(`/api/namespaces/${ namespaceName }/instances/${ stream2InstanceId }/output`)
			.send()
		statusStream2 = await request(common.config.getDirektivHost()).get(`/api/namespaces/${ namespaceName }/instances/${ stream2InstanceId }`)
			.send()
		expect(statusStream2.body.instance.status).toBe('complete')

		statusStream3 = await request(common.config.getDirektivHost()).get(`/api/namespaces/${ namespaceName }/instances/${ stream3InstanceId }`)
			.send()
		expect(statusStream3.body.instance.status).toBe('pending')

		const outputData2 = JSON.parse(atob(resultsStream2.body.data)); +expect(outputData2.hello.hello).toBe('condition2')

		await events.sendEventAndList(namespaceName, eventStream3Stage2)
		await helpers.sleep(300)

		const resultsStream3 = await request(common.config.getDirektivHost()).get(`/api/namespaces/${ namespaceName }/instances/${ stream3InstanceId }/output`)
			.send()

		statusStream3 = await request(common.config.getDirektivHost()).get(`/api/namespaces/${ namespaceName }/instances/${ stream3InstanceId }`)
			.send()
		expect(statusStream3.body.instance.status).toBe('complete')
		const outputData3 = JSON.parse(atob(resultsStream3.body.data))
		expect(outputData3.hello.hello).toBe('condition3')

		instancesResponse = await request(common.config.getDirektivHost()).get(`/api/namespaces/${ namespaceName }/instances?limit=10&offset=0`)
			.send()
		expect(instancesResponse.body.instances.results.length).toBe(3)
	})
})
