import { beforeAll, describe, expect, it } from '@jest/globals'

import common from '../common'
import helpers from '../common/helpers'
import request from '../common/request'
import events from './send_helper'

const namespaceName = 'wfevents'

const baseEventWithContext = (type, id, ck, cv) => `{
    "specversion" : "1.0",
    "type" : "${ type }",
    "id": "${ id }",
    "source" : "https://direktiv.io/test",
    "datacontenttype" : "application/json",
    "${ ck }": "${ cv }",
    "data" : {
        "hello": "world",
        "123": 456
    }
}`

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

describe('Test basic workflow events', () => {
	beforeAll(common.helpers.deleteAllNamespaces)

	it(`should create namespace`, async () => {
		const createNamespaceResponse = await request(common.config.getDirektivHost()).put(`/api/namespaces/${ namespaceName }`)
		expect(createNamespaceResponse.statusCode).toEqual(200)
	})

	common.helpers.itShouldCreateYamlFileV2(it, expect, namespaceName,
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

	it(`should fail to invoke the '/listener.yml' workflow`, async () => {
		const req = await request(common.config.getDirektivHost()).get(`/api/namespaces/${ namespaceName }/tree/listener.yml?op=wait`)
		expect(req.statusCode).toEqual(500)
		expect(req.body).toMatchObject({
			code: 500,
			message: 'cannot manually invoke event-based workflow',
		})
	})

	it(`should invoke the '/listener.yml' workflow with an event`, async () => {
		await events.sendEventAndList(namespaceName, basevent('greeting', 'greeting', 'world1'))

		const instance = await events.listInstancesAndFilter(namespaceName, 'listener.yml')
		expect(instance).not.toBeFalsy()
	})
})

describe('Test workflow events with filter/context', () => {
	beforeAll(common.helpers.deleteAllNamespaces)

	it(`should create namespace`, async () => {
		const createNamespaceResponse = await request(common.config.getDirektivHost()).put(`/api/namespaces/${ namespaceName }`)
		expect(createNamespaceResponse.statusCode).toEqual(200)
	})

	common.helpers.itShouldCreateYamlFileV2(it, expect, namespaceName,
		'/', 'startlistener.yml', 'workflow', `
start:
  type: event
  event:
    type: greeting
    context:
        state: "started"
  state: helloworld
states:
- id: helloworld
  type: noop
  transform:
    result: Hello world!
`)

	common.helpers.itShouldCreateYamlFileV2(it, expect, namespaceName,
		'/', 'stoplistener.yml', 'workflow', `
    start:
      type: event
      event:
        type: greeting
        context:
            state: "stopped"
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

	it(`should invoke the '/stoplistener.yml' workflow with an event`, async () => {
		await events.sendEventAndList(namespaceName, baseEventWithContext('greeting', 'greeting', 'state', 'stopped'))

		let instance = await events.listInstancesAndFilter(namespaceName, 'startlistener.yml')
		expect(instance).toBeFalsy()

		instance = await events.listInstancesAndFilter(namespaceName, 'stoplistener.yml')
		expect(instance).not.toBeFalsy()
	})
})

const startWorkflowName = 'start.yaml'
const startWorkflow = `
start:
  type: event
  state: helloworld
  event: 
    type: hello
states:
- id: helloworld
  type: noop
  transform: jq(.)
`

const waitWorkflowName = 'wait.yaml'
const waitWorkflow = `
states:
- id: ce
  type: consumeEvent
  event:
    type: hellowait
  timeout: PT1H
  transition: print
- id: print
  type: noop
  log: jq(.)
`

const startWorkflowNameContext = 'startcontext.yaml'
const startEventWorkflowContext = `
start:
  type: event
  state: helloworld
  event: 
    type: helloctx
    context:
      hello: world
states:
- id: helloworld
  type: noop
  transform: jq(.)
`

const waitWorkflowNameContext = 'waitcontext.yaml'
const waitWorkflowContext = `
states:
- id: ce
  type: consumeEvent
  event:
    type: hellowait
    context:
      hello: world
  timeout: PT1H
  transition: print
- id: print
  type: noop
  log: jq(.)
`

describe('Test workflow events', () => {
	beforeAll(common.helpers.deleteAllNamespaces)

	it(`should create namespace`, async () => {
		const createNamespaceResponse = await request(common.config.getDirektivHost()).put(`/api/namespaces/${ namespaceName }`)
		expect(createNamespaceResponse.statusCode).toEqual(200)
	})

	// workflow with start
	helpers.itShouldCreateYamlFileV2(it, expect, namespaceName,
		'', startWorkflowName, 'workflow',
		startWorkflow)

	it(`should have one event listeners`, async () => {
		await helpers.sleep(1000)

		const getEventListenerResponse = await request(common.config.getDirektivHost()).get(`/api/namespaces/${ namespaceName }/event-listeners?limit=8&offset=0`)
			.send()

		expect(getEventListenerResponse.body.results[0]).toMatchObject({
			workflow: '/start.yaml',
			mode: 'simple',
			instance: '',
			createdAt: expect.stringMatching(common.regex.timestampRegex),
			updatedAt: expect.stringMatching(common.regex.timestampRegex),
			events: [ { type: 'hello',
				filters: {} } ],
		})

		expect(getEventListenerResponse.body.pageInfo.total).toEqual(1)
	})

	// workflow with start
	helpers.itShouldCreateYamlFileV2(it, expect, namespaceName,
		'', waitWorkflowName, 'workflow',
		waitWorkflow)

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
			mode: 'simple',
			instance: expect.stringMatching(common.regex.uuidRegex),
			createdAt: expect.stringMatching(common.regex.timestampRegex),
			updatedAt: expect.stringMatching(common.regex.timestampRegex),
			events: [ { type: 'hellowait',
				filters: {} } ],
		})
	})

	it(`should kick off in flow workflow with custom attributes`, async () => {
		// should not continue workflow
		await events.sendEventAndList(namespaceName, basevent('no-kick', 'json-event'))

		// the waiting workflow is running but nothing triggered by event, state pending
		let instancesResponse = await events.listInstancesAndFilter(namespaceName, waitWorkflowName, 'pending')
		expect(instancesResponse).not.toBeFalsy()

		await events.sendEventAndList(namespaceName, basevent('hellowait', 'testinflow', 'world'))

		instancesResponse = await events.listInstancesAndFilter(namespaceName, waitWorkflowName, 'complete')
		expect(instancesResponse).not.toBeFalsy()

		const instanceOutput = await request(common.config.getDirektivHost()).get(`/api/namespaces/${ namespaceName }/instances/${ instancesResponse.id }/output`)
			.send()

		const output = Buffer.from(instanceOutput.body.data, 'base64')
		const outputJSON = JSON.parse(output.toString())

		// custom value set
		expect(outputJSON.hellowait.hello).toEqual('world')
	})

	it(`should kick off start event workflow`, async () => {
		await events.sendEventAndList(namespaceName, basevent('hello', 'start-event'))
		const instance = await events.listInstancesAndFilter(namespaceName, startWorkflowName, 'complete')
		expect(instance).not.toBeFalsy()

		const instanceOutput = await request(common.config.getDirektivHost()).get(`/api/namespaces/${ namespaceName }/instances/${ instance.id }/output`)
			.send()

		const output = Buffer.from(instanceOutput.body.data, 'base64')
		const outputJSON = JSON.parse(output.toString())

		// custom data set
		expect(outputJSON.hello.data.hello).toEqual('world')
	})

	helpers.itShouldCreateYamlFileV2(it, expect, namespaceName,
		'', startWorkflowNameContext, 'workflow',
		startEventWorkflowContext)

	it(`should kick off start event workflow with context filter`, async () => {
		await helpers.sleep(1000)

		// send event with same type but without context
		await events.sendEventAndList(namespaceName, basevent('helloctx', 'ctx-test'))
		let instancesResponse = await events.listInstancesAndFilter(namespaceName, startWorkflowNameContext)

		// no instance fired
		expect(instancesResponse).toBeFalsy()

		await events.sendEventAndList(namespaceName, basevent('helloctx', 'ctx-test-fire', 'world'))
		instancesResponse = await events.listInstancesAndFilter(namespaceName, startWorkflowNameContext)

		// instance fired
		expect(instancesResponse).not.toBeFalsy()
	})

	helpers.itShouldCreateYamlFileV2(it, expect, namespaceName,
		'', waitWorkflowNameContext, 'workflow',
		waitWorkflowContext)

	it(`should kick off running workflow with context filter`, async () => {
		await helpers.sleep(1000)

		// start workflow
		const runWorkflowResponse = await request(common.config.getDirektivHost()).post(`/api/namespaces/${ namespaceName }/tree/${ waitWorkflowNameContext }?op=execute`)
			.send()
		expect(runWorkflowResponse.statusCode).toEqual(200)

		// send event with same type but without context
		await events.sendEventAndList(namespaceName, basevent('hellowait', 'wait-ctx', 'dummy'))
		let instancesResponse = await events.listInstancesAndFilter(namespaceName, waitWorkflowNameContext, 'pending')

		// no instance fired, still pending
		expect(instancesResponse).not.toBeFalsy()

		await events.sendEventAndList(namespaceName, basevent('hellowait', 'wait-ctx-run', 'world'))

		instancesResponse = await events.listInstancesAndFilter(namespaceName, waitWorkflowNameContext, 'complete')

		// instance fired
		expect(instancesResponse).not.toBeFalsy()
	})
})
