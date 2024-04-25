import { beforeAll, describe, expect, it } from '@jest/globals'

import common from '../common'
import helpers from '../common/helpers'
import request from '../common/request'
import events from './send_helper'

const namespaceName = 'sendeventsand'

const waitWorkflowName = 'waitand.yaml'
const waitEventWorkflow = `
states:
- id: event-and
  type: eventsAnd
  timeout: PT1H
  events:
  - type: eventtype1
  - type: eventtype2
  transition: greet
- id: greet
  type: noop
  log: jq(.)
`

const startWorkflowName = 'startand.yaml'
const startEventWorkflow = `
start:
  type: eventsAnd
  state: greet
  events:
    - type: eventtype3
    - type: eventtype4
states:
- id: greet
  type: noop
  log: jq(.)
`

const startWorkflowContextName = 'startandcontext.yaml'
const startEventContextWorkflow = `
start:
  type: eventsAnd
  state: greet
  events:
    - type: eventtype9
      context:
        hello: world1
    - type: eventtype10
      context: 
        hello: world2
states:
- id: greet
  type: noop
  log: jq(.)
`

const waitWorkflowContextName = 'waitandcontext.yaml'
const waitEventContextimeout = `
states:
- id: event-and
  type: eventsAnd
  events:
  - type: eventtype11
    context:
      hello: world1
  - type: eventtype12
    context:
      hello: world1
  transition: greet
- id: greet
  type: noop
  log: jq(.)
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

describe('Test workflow events and', () => {
	beforeAll(common.helpers.deleteAllNamespaces)

	helpers.itShouldCreateNamespace(it, expect, namespaceName)

	// workflow with start
	helpers.itShouldCreateYamlFileV2(it, expect, namespaceName,
		'', startWorkflowName, 'workflow',
		startEventWorkflow)

	it(`should have one event listeners`, async () => {
		await helpers.sleep(1000)

		const getEventListenerResponse = await request(common.config.getDirektivHost()).get(`/api/namespaces/${ namespaceName }/event-listeners?limit=8&offset=0`)
			.send()

		expect(getEventListenerResponse.body.results[0]).toMatchObject({
			workflow: '/startand.yaml',
			mode: 'and',
			instance: '',
			createdAt: expect.stringMatching(common.regex.timestampRegex),
			updatedAt: expect.stringMatching(common.regex.timestampRegex),
			events: [ {
				type: 'eventtype3',
				filters: {},
			}, {
				type: 'eventtype4',
				filters: {},
			} ],
		})

		expect(getEventListenerResponse.body.pageInfo.total).toEqual(1)
	})

	// workflow with start
	helpers.itShouldCreateYamlFileV2(it, expect, namespaceName,
		'', waitWorkflowName, 'workflow',
		waitEventWorkflow)

	it(`should have two event listeners`, async () => {
		await helpers.sleep(1000)

		// start workflow
		const runWorkflowResponse = await request(common.config.getDirektivHost()).post(`/api/namespaces/${ namespaceName }/tree/${ waitWorkflowName }?op=execute`)
			.send()
		expect(runWorkflowResponse.statusCode).toEqual(200)

		await new Promise(r => setTimeout(r, 250))

		const getEventListenerResponse = await request(common.config.getDirektivHost()).get(`/api/namespaces/${ namespaceName }/event-listeners?limit=8&offset=0`)
			.send()

		expect(getEventListenerResponse.body.pageInfo.total).toEqual(2)

		const result = getEventListenerResponse.body.results.find(item => item.workflow === '')

		expect(result).toMatchObject({
			workflow: '',
			mode: 'and',
			instance: expect.stringMatching(common.regex.uuidRegex),
			createdAt: expect.stringMatching(common.regex.timestampRegex),
			updatedAt: expect.stringMatching(common.regex.timestampRegex),
			events: [ {
				type: 'eventtype1',
				filters: {},
			}, {
				type: 'eventtype2',
				filters: {},
			} ],
		})
	})

	it(`should kick off in flow workflow`, async () => {
		// should not continue workflow
		await events.sendEventAndList(namespaceName, basevent('eventtype1', 'eventtype1', 'world1'))

		// the waiting workflow is running but nothing triggered by event, state pending
		let instancesResponse = await events.listInstancesAndFilter(namespaceName, waitWorkflowName, 'pending')
		expect(instancesResponse).not.toBeFalsy()

		//  await events.sendEventAndList(namespaceName, basevent("eventtype1", "eventtype1", "world"))
		await events.sendEventAndList(namespaceName, basevent('eventtype2', 'eventtype2', 'world2'))
		instancesResponse = await events.listInstancesAndFilter(namespaceName, waitWorkflowName, 'complete')
		expect(instancesResponse).not.toBeFalsy()

		const instanceOutput = await request(common.config.getDirektivHost()).get(`/api/namespaces/${ namespaceName }/instances/${ instancesResponse.id }/output`)
			.send()

		const output = Buffer.from(instanceOutput.body.data, 'base64')
		const outputJSON = JSON.parse(output.toString())

		//  custom value set
		expect(outputJSON.eventtype1.hello).toEqual('world1')
		expect(outputJSON.eventtype2.hello).toEqual('world2')
	})

	it(`should kick off start event workflow`, async () => {
		await events.sendEventAndList(namespaceName, basevent('eventtype3', 'eventtype3', 'world1'))
		let instance = await events.listInstancesAndFilter(namespaceName, startWorkflowName)

		expect(instance).toBeFalsy()

		await events.sendEventAndList(namespaceName, basevent('eventtype4', 'eventtype4', 'world2'))
		instance = await events.listInstancesAndFilter(namespaceName, startWorkflowName, 'complete')

		const instanceOutput = await request(common.config.getDirektivHost()).get(`/api/namespaces/${ namespaceName }/instances/${ instance.id }/output`)
			.send()

		const output = Buffer.from(instanceOutput.body.data, 'base64')
		const outputJSON = JSON.parse(output.toString())

		// custom data set
		expect(outputJSON.eventtype3.data.hello).toEqual('world')
		expect(outputJSON.eventtype4.data.hello).toEqual('world')
	})

	// timeout workflow
	helpers.itShouldCreateYamlFileV2(it, expect, namespaceName,
		'', startWorkflowContextName, 'workflow',
		startEventContextWorkflow)

	it(`start context`, async () => {
		await helpers.sleep(1000)

		let workflowEventResponse = await request(common.config.getDirektivHost()).post(`/api/namespaces/${ namespaceName }/broadcast`)
			.set('Content-Type', 'application/json')
			.send(basevent('eventtype9', 'eventtype9zz', 'world1'))
		expect(workflowEventResponse.statusCode).toEqual(200)

		workflowEventResponse = await request(common.config.getDirektivHost()).post(`/api/namespaces/${ namespaceName }/broadcast`)
			.set('Content-Type', 'application/json')
			.send(basevent('eventtype10', 'eventtype1324320', 'world3'))
		expect(workflowEventResponse.statusCode).toEqual(200)

		let instance = await events.listInstancesAndFilter(namespaceName, startWorkflowContextName)
		expect(instance).toBeFalsy()

		workflowEventResponse = await request(common.config.getDirektivHost()).post(`/api/namespaces/${ namespaceName }/broadcast`)
			.set('Content-Type', 'application/json')
			.send(basevent('eventtype10', 'eventtype10afg', 'world2'))
		expect(workflowEventResponse.statusCode).toEqual(200)

		instance = await events.listInstancesAndFilter(namespaceName, startWorkflowContextName)
		expect(instance).not.toBeFalsy()
	})

	// timeout workflow
	helpers.itShouldCreateYamlFileV2(it, expect, namespaceName,
		'', waitWorkflowContextName, 'workflow',
		waitEventContextimeout)

	it(`flow context`, async () => {
		await helpers.sleep(1000)

		const runWorkflowResponse = await request(common.config.getDirektivHost()).post(`/api/namespaces/${ namespaceName }/tree/${ waitWorkflowContextName }?op=execute`)
			.send()
		expect(runWorkflowResponse.statusCode).toEqual(200)

		await helpers.sleep(1000)

		await events.sendEventAndList(namespaceName, basevent('eventtype11', 'eventtype11dfds', 'world1'))
		await events.sendEventAndList(namespaceName, basevent('eventtype12', 'eventtype12dsfds', 'world3'))

		let instance = await events.listInstancesAndFilter(namespaceName, waitWorkflowContextName, 'pending')
		expect(instance).not.toBeFalsy()

		await events.sendEventAndList(namespaceName, basevent('eventtype11', 'eventtype12a', 'world2'))
		instance = await events.listInstancesAndFilter(namespaceName, waitWorkflowContextName, 'pending')
		expect(instance).not.toBeFalsy()

		await events.sendEventAndList(namespaceName, basevent('eventtype11', 'eventtype12abc', 'world4'))
		instance = await events.listInstancesAndFilter(namespaceName, waitWorkflowContextName, 'pending')
		expect(instance).not.toBeFalsy()

		await events.sendEventAndList(namespaceName, basevent('eventtype11', 'eventtype12ab', 'world2'))
		instance = await events.listInstancesAndFilter(namespaceName, startWorkflowContextName, 'complete')
		expect(instance).not.toBeFalsy()
	})
})
