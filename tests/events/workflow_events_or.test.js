import { beforeAll, describe, expect, it } from '@jest/globals'

import common from '../common'
import helpers from '../common/helpers'
import request from '../common/request'
import events from './send_helper'

const namespaceName = 'sendeventsor'

const waitWorkflowName = 'waitor.yaml'
const waitEventWorkflow = `
states:
- id: event-and
  type: eventsXor
  timeout: PT1H
  events:
  - event:
      type: eventtype1
    transition: greet
  - event:
      type: eventtype2
    transition: greet
- id: greet
  type: noop
  log: jq(.)
`

const startWorkflowName = 'startor.yaml'
const startEventWorkflow = `
start:
  type: eventsXor
  state: greet
  events:
    - type: eventtype3
    - type: eventtype4
states:
- id: greet
  type: noop
  log: jq(.)
`

const waitWorkflowTimeoutName = 'waitandtimeout.yaml'
const waitEventWorkflowTimeout = `
states:
- id: event-and
  type: eventsXor
  timeout: PT1S
  events:
  - event:
      type: eventtype5
    transition: greet
  - event:
      type: eventtype6
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

	it(`should create namespace`, async () => {
		const createNamespaceResponse = await request(common.config.getDirektivHost()).put(`/api/namespaces/${ namespaceName }`)
		expect(createNamespaceResponse.statusCode).toEqual(200)
	})

	// workflow with start
	helpers.itShouldCreateYamlFileV2(it, expect, namespaceName,
		'', startWorkflowName, 'workflow',
		startEventWorkflow)

	it(`should have one event listeners`, async () => {
		await helpers.sleep(1000)

		const getEventListenerResponse = await request(common.config.getDirektivHost()).get(`/api/namespaces/${ namespaceName }/event-listeners?limit=8&offset=0`)
			.send()

		expect(getEventListenerResponse.body.results[0]).toMatchObject({
			workflow: '/startor.yaml',
			mode: 'or',
			instance: '',
			createdAt: expect.stringMatching(common.regex.timestampRegex),
			updatedAt: expect.stringMatching(common.regex.timestampRegex),
			events: [ { type: 'eventtype3',
				filters: {} }, { type: 'eventtype4',
				filters: {} } ],
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
			mode: 'or',
			instance: expect.stringMatching(common.regex.uuidRegex),
			createdAt: expect.stringMatching(common.regex.timestampRegex),
			updatedAt: expect.stringMatching(common.regex.timestampRegex),
			events: [ { type: 'eventtype1',
				filters: {} }, { type: 'eventtype2',
				filters: {} } ],
		})
	})

	it(`should kick off in flow workflow`, async () => {
		// should fire workflow
		await events.sendEventAndList(namespaceName, basevent('eventtype1', 'eventtype1', 'world1'))

		let instancesResponse = await events.listInstancesAndFilter(namespaceName, waitWorkflowName, 'complete')
		expect(instancesResponse).not.toBeFalsy()

		const instanceOutput = await request(common.config.getDirektivHost()).get(`/api/namespaces/${ namespaceName }/instances/${ instancesResponse.id }/output`)
			.send()

		const output = Buffer.from(instanceOutput.body.data, 'base64')
		const outputJSON = JSON.parse(output.toString())

		//  custom value set
		expect(outputJSON.eventtype1.hello).toEqual('world1')

		// restart workflow
		const runWorkflowResponse = await request(common.config.getDirektivHost()).post(`/api/namespaces/${ namespaceName }/tree/${ waitWorkflowName }?op=execute`)
			.send()
		expect(runWorkflowResponse.statusCode).toEqual(200)
		await new Promise(r => setTimeout(r, 250))

		await events.sendEventAndList(namespaceName, basevent('eventtype2', 'eventtype2a', 'world2'))

		// there are two workflows now
		instancesResponse = await request(common.config.getDirektivHost()).get(`/api/namespaces/${ namespaceName }/instances?limit=10&offset=0&filter.field=AS&filter.type=CONTAINS&filter.val=` + waitWorkflowName)
			.send()
		expect(instancesResponse.body.instances.pageInfo.total).toEqual(2)
	})

	it(`should kick off start event workflow`, async () => {
		await events.sendEventAndList(namespaceName, basevent('eventtype3', 'eventtype3', 'world1'))
		const instance = await events.listInstancesAndFilter(namespaceName, startWorkflowName, 'complete')

		expect(instance).not.toBeFalsy()

		const instanceOutput = await request(common.config.getDirektivHost()).get(`/api/namespaces/${ namespaceName }/instances/${ instance.id }/output`)
			.send()

		const output = Buffer.from(instanceOutput.body.data, 'base64')
		const outputJSON = JSON.parse(output.toString())

		// custom data set
		expect(outputJSON.eventtype3.data.hello).toEqual('world')

		await events.sendEventAndList(namespaceName, basevent('eventtype4', 'eventtype4', 'world2'))

		const instancesResponse = await request(common.config.getDirektivHost()).get(`/api/namespaces/${ namespaceName }/instances?limit=10&offset=0&filter.field=AS&filter.type=CONTAINS&filter.val=` + startWorkflowName)
			.send()
		expect(instancesResponse.body.instances.pageInfo.total).toEqual(2)
	})

	// timeout workflow
	helpers.itShouldCreateYamlFileV2(it, expect, namespaceName,
		'', waitWorkflowTimeoutName, 'workflow',
		waitEventWorkflowTimeout)

	it(`should timeout flow`, async () => {
		await helpers.sleep(1000)

		// start workflow
		const runWorkflowResponse = await request(common.config.getDirektivHost()).post(`/api/namespaces/${ namespaceName }/tree/${ waitWorkflowTimeoutName }?op=execute`)
			.send()
		expect(runWorkflowResponse.statusCode).toEqual(200)
		await new Promise(r => setTimeout(r, 7000))

		const instancesResponse = await events.listInstancesAndFilter(namespaceName, waitWorkflowTimeoutName, 'failed')
		expect(instancesResponse).not.toBeFalsy()
	})
})
