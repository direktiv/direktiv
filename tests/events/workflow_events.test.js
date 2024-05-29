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

	helpers.itShouldCreateNamespace(it, expect, namespaceName)
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
		const req = await request(common.config.getDirektivHost()).post(`/api/v2/namespaces/${ namespaceName }/instances?path=listener.yml&wait=true`)

		expect(req.statusCode).toEqual(500)
		expect(req.body).toMatchObject({
			error: {
				code: 'cannot manually invoke event-based workflow',
				message: 'cannot manually invoke event-based workflow',
			},
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

	helpers.itShouldCreateNamespace(it, expect, namespaceName)

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

const workflowContextMultipleName = 'waitmulticontext.yaml'
const eventContextMultipleWorkflow = `
states:
- id: ce
  type: consumeEvent
  event:
    type: waitformulti
    context:
      hello: world1
      hello2: world2
  transition: greet
- id: greet
  type: noop
  log: jq(.)
`

const baseventMultipleContext = (type, id) => `{
    "specversion" : "1.0",
    "type" : "${ type }",
    "id": "${ id }",
    "source" : "https://direktiv.io/test",
    "datacontenttype" : "application/json",
	"hello": "world1",
	"hello2": "world2",
    "data" : {
		"hello": "world",
        "123": 456
    }
}`

describe('Test workflow events', () => {
	beforeAll(common.helpers.deleteAllNamespaces)

	helpers.itShouldCreateNamespace(it, expect, namespaceName)

	// workflow with start
	helpers.itShouldCreateYamlFileV2(it, expect, namespaceName,
		'', startWorkflowName, 'workflow',
		startWorkflow)

	it(`should have one event listeners`, async () => {
		await helpers.sleep(1000)

		const getEventListenerResponse = await request(common.config.getDirektivHost()).get(`/api/v2/namespaces/${ namespaceName }/events/listeners?limit=8&offset=0`)
			.send()
		expect(getEventListenerResponse.body.data[0]).toMatchObject({
			triggerWorkflow: '/start.yaml',
			triggerType: 'StartSimple',
			namespace: { namespaceName }.namespaceName,
			createdAt: expect.stringMatching(common.regex.timestampRegex),
			updatedAt: expect.stringMatching(common.regex.timestampRegex),
			eventContextFilters: [ {
				type: 'hello',
			} ],
		})

		expect(getEventListenerResponse.body.meta.total).toEqual(1)
	})

	// workflow with start
	helpers.itShouldCreateYamlFileV2(it, expect, namespaceName,
		'', waitWorkflowName, 'workflow',
		waitWorkflow)

	it(`should have two event listeners`, async () => {
		await helpers.sleep(1000)

		// start workflow
		const runWorkflowResponse = await request(common.config.getDirektivHost()).post(`/api/v2/namespaces/${ namespaceName }/instances?path=${ waitWorkflowName }`)
			.send()
		expect(runWorkflowResponse.statusCode).toEqual(200)

		await new Promise(r => setTimeout(r, 250))

		const getEventListenerResponse = await request(common.config.getDirektivHost()).get(`/api/v2/namespaces/${ namespaceName }/events/listeners?limit=8&offset=0`)
			.send()
		expect(getEventListenerResponse.body.meta.total).toEqual(2)

		const result = getEventListenerResponse.body.data.find(item => item.hasOwnProperty('triggerInstance'))
		expect(result).toMatchObject({
			triggerType: 'WaitSimple',
			triggerInstance: expect.stringMatching(common.regex.uuidRegex),
			createdAt: expect.stringMatching(common.regex.timestampRegex),
			updatedAt: expect.stringMatching(common.regex.timestampRegex),
			eventContextFilters: [ {
				type: 'hellowait',
			} ],
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

		const instanceOutput = await request(common.config.getDirektivHost()).get(`/api/v2/namespaces/${ namespaceName }/instances/${ instancesResponse.id }/output`)
		const output = Buffer.from(instanceOutput.body.data.output, 'base64')
		const outputJSON = JSON.parse(output.toString())

		// custom value set
		expect(outputJSON.hellowait.hello).toEqual('world')
	})

	it(`should kick off start event workflow`, async () => {
		await events.sendEventAndList(namespaceName, basevent('hello', 'start-event'))
		const instance = await events.listInstancesAndFilter(namespaceName, startWorkflowName, 'complete')
		expect(instance).not.toBeFalsy()

		const instanceOutput = await request(common.config.getDirektivHost()).get(`/api/v2/namespaces/${ namespaceName }/instances/${ instance.id }/output`)
		const output = Buffer.from(instanceOutput.body.data.output, 'base64')
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
		const runWorkflowResponse = await request(common.config.getDirektivHost()).post(`/api/v2/namespaces/${ namespaceName }/instances?path=${ waitWorkflowNameContext }`)
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

	helpers.itShouldCreateYamlFileV2(it, expect, namespaceName,
		'', startThenWaitWorkflowNameContext, 'workflow',
		starthenWaitWorkflowContext)

	it(`should start by event and kick off running workflow with context filter`, async () => {
		await helpers.sleep(1000)

		await events.sendEventAndList(namespaceName, basevent('hello', 'wait-ctx2', 'condition1'))
		let instancesResponse = await events.listInstancesAndFilter(namespaceName, startThenWaitWorkflowNameContext, 'pending')

		// no instance fired, still pending
		expect(instancesResponse).not.toBeFalsy()
		await events.sendEventAndList(namespaceName, basevent('hellowait', 'wait-ctx-run3', 'condition2'))
		instancesResponse = await events.listInstancesAndFilter(namespaceName, startThenWaitWorkflowNameContext, 'pending')
		// no instance fired, still pending
		expect(instancesResponse).not.toBeFalsy()
		await events.sendEventAndList(namespaceName, basevent('hellowait', 'wait-ctx-run4', 'condition1'))
		instancesResponse = await events.listInstancesAndFilter(namespaceName, startThenWaitWorkflowNameContext, 'complete')

		// instance fired
		expect(instancesResponse).not.toBeFalsy()
	})

	// workflow with multiple context-filters
	helpers.itShouldCreateYamlFileV2(it, expect, namespaceName,
		'', workflowContextMultipleName, 'workflow',
		eventContextMultipleWorkflow)

	it(`should not start by event due to context filter`, async () => {
		await helpers.sleep(2000)
		const runWorkflowResponse = await request(common.config.getDirektivHost()).post(`/api/v2/namespaces/${ namespaceName }/instances?path=${ workflowContextMultipleName }`)
			.send()
		expect(runWorkflowResponse.statusCode).toEqual(200)

		await events.sendEventAndList(namespaceName, basevent('waitformulti', 'wait-ctx65', 'world1'))
		let instancesResponse = await events.listInstancesAndFilter(namespaceName, workflowContextMultipleName, 'pending')
		expect(instancesResponse).not.toBeFalsy()
		const workflowEventResponse = await request(common.config.getDirektivHost()).post(`/api/v2/namespaces/${ namespaceName }/events/broadcast`)
			.set('Content-Type', 'application/json')
			.send(baseventMultipleContext('waitformulti', 'wait-c3432tx7'))
		expect(workflowEventResponse.statusCode).toEqual(200)
		instancesResponse = await events.listInstancesAndFilter(namespaceName, workflowContextMultipleName, 'complete')
		expect(instancesResponse).not.toBeFalsy()
	})
})
