import request from 'supertest'

import common from "../common"

import events from "./send_helper.js"

const namespaceName = "wfevents"

const basevent = (type, id, value) => `{
    "specversion" : "1.0",
    "type" : "${type}",
    "id": "${id}",
    "source" : "https://direktiv.io/test",
    "datacontenttype" : "application/json",
    "hello": "${value}",
    "data" : {
        "hello": "world",
        "123": 456
    }
}`

const startWorkflowName = "start.yaml"
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

const waitWorkflowName = "wait.yaml"
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

const startWorkflowNameContext = "startcontext.yaml"
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

const waitWorkflowNameContext = "waitcontext.yaml"
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
        var createNamespaceResponse = await request(common.config.getDirektivHost()).put(`/api/namespaces/${namespaceName}`)
        expect(createNamespaceResponse.statusCode).toEqual(200)
    })

    it(`should have one event listeners`, async () => {

        // workflow with start
        var createWorkflowResponse = await request(common.config.getDirektivHost()).put(`/api/namespaces/${namespaceName}/tree/${startWorkflowName}?op=create-workflow`)
            .send(startWorkflow)

        expect(createWorkflowResponse.statusCode).toEqual(200)

        var getEventListenerResponse = await request(common.config.getDirektivHost()).get(`/api/namespaces/${namespaceName}/event-listeners?limit=8&offset=0`)
            .send()

        expect(getEventListenerResponse.body.results[0]).toMatchObject({
            workflow: "/start.yaml",
            "mode": "simple",
            "instance": "",
            "createdAt": expect.stringMatching(common.regex.timestampRegex),
            "updatedAt": expect.stringMatching(common.regex.timestampRegex),
            "events": [{"type": "hello", "filters": {}}]
        });

        expect(getEventListenerResponse.body.pageInfo.total).toEqual(1)

    })

    it(`should have two event listeners`, async () => {

        // workflow with start
        var createWorkflowResponse = await request(common.config.getDirektivHost()).put(`/api/namespaces/${namespaceName}/tree/${waitWorkflowName}?op=create-workflow`)
            .send(waitWorkflow)
        expect(createWorkflowResponse.statusCode).toEqual(200)

        // start workflow
        var runWorkflowResponse = await request(common.config.getDirektivHost()).post(`/api/namespaces/${namespaceName}/tree/${waitWorkflowName}?op=execute`)
            .send()
        expect(runWorkflowResponse.statusCode).toEqual(200)

        await new Promise((r) => setTimeout(r, 250));

        var getEventListenerResponse = await request(common.config.getDirektivHost()).get(`/api/namespaces/${namespaceName}/event-listeners?limit=8&offset=0`)
            .send()

        expect(getEventListenerResponse.body.pageInfo.total).toEqual(2)

        var result = getEventListenerResponse.body.results.find(item => item.workflow === "");

        expect(result).toMatchObject({
            workflow: "",
            "mode": "simple",
            "instance": expect.stringMatching(common.regex.uuidRegex),
            "createdAt": expect.stringMatching(common.regex.timestampRegex),
            "updatedAt": expect.stringMatching(common.regex.timestampRegex),
            "events": [{"type": "hellowait", "filters": {}}]
        });


    })

    it(`should kick off in flow workflow with custom attributes`, async () => {

        // should not continue workflow
        await events.sendEventAndList(namespaceName, basevent("no-kick", "json-event"))

        // the waiting workflow is running but nothing triggered by event, state pending
        var instancesResponse = await events.listInstancesAndFilter(namespaceName, waitWorkflowName, "pending")
        expect(instancesResponse).not.toBeFalsy();

        await events.sendEventAndList(namespaceName, basevent("hellowait", "testinflow", "world"))

        var instancesResponse = await events.listInstancesAndFilter(namespaceName, waitWorkflowName, "complete")
        expect(instancesResponse).not.toBeFalsy();

        var instanceOutput = await request(common.config.getDirektivHost()).get(`/api/namespaces/${namespaceName}/instances/${instancesResponse.id}/output`)
            .send()

        var output = Buffer.from(instanceOutput.body.data, 'base64');
        var outputJSON = JSON.parse(output.toString())

        // custom value set
        expect(outputJSON["hellowait"].hello).toEqual("world")

    })


    it(`should kick off start event workflow`, async () => {


        await events.sendEventAndList(namespaceName, basevent("hello", "start-event"))
        var instance = await events.listInstancesAndFilter(namespaceName, startWorkflowName)
        expect(instance).not.toBeFalsy();

        var instanceOutput = await request(common.config.getDirektivHost()).get(`/api/namespaces/${namespaceName}/instances/${instance.id}/output`)
            .send()

        var output = Buffer.from(instanceOutput.body.data, 'base64');
        var outputJSON = JSON.parse(output.toString())

        // custom data set
        expect(outputJSON["hello"].data.hello).toEqual("world")

    })


    it(`should kick off start event workflow with context filter`, async () => {

        var createWorkflowResponse = await request(common.config.getDirektivHost()).put(`/api/namespaces/${namespaceName}/tree/${startWorkflowNameContext}?op=create-workflow`)
            .send(startEventWorkflowContext)
        expect(createWorkflowResponse.statusCode).toEqual(200)

        // send event with same type but without context
        await events.sendEventAndList(namespaceName, basevent("helloctx", "ctx-test"))
        var instancesResponse = await events.listInstancesAndFilter(namespaceName, startWorkflowNameContext)

        // no instance fired
        expect(instancesResponse).toBeFalsy();

        await events.sendEventAndList(namespaceName, basevent("helloctx", "ctx-test-fire", "world"))
        var instancesResponse = await events.listInstancesAndFilter(namespaceName, startWorkflowNameContext)

        // instance fired
        expect(instancesResponse).not.toBeFalsy();

    })


    it(`should kick off running workflow with context filter`, async () => {

        var createWorkflowResponse = await request(common.config.getDirektivHost()).put(`/api/namespaces/${namespaceName}/tree/${waitWorkflowNameContext}?op=create-workflow`)
            .send(waitWorkflowContext)
        expect(createWorkflowResponse.statusCode).toEqual(200)

        // start workflow
        var runWorkflowResponse = await request(common.config.getDirektivHost()).post(`/api/namespaces/${namespaceName}/tree/${waitWorkflowNameContext}?op=execute`)
            .send()
        expect(runWorkflowResponse.statusCode).toEqual(200)

        // send event with same type but without context
        await events.sendEventAndList(namespaceName, basevent("hellowait", "wait-ctx", "dummy"))
        var instancesResponse = await events.listInstancesAndFilter(namespaceName, waitWorkflowNameContext, "pending")

        // no instance fired, still pending
        expect(instancesResponse).not.toBeFalsy();

        await events.sendEventAndList(namespaceName, basevent("hellowait", "wait-ctx-run", "world"))

        var instancesResponse = await events.listInstancesAndFilter(namespaceName, waitWorkflowNameContext, "complete")

        // instance fired
        expect(instancesResponse).not.toBeFalsy();

    })

})
