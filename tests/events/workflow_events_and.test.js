import common from "../common"
import events from "./send_helper.js"
import request from 'supertest'

const namespaceName = "sendeventsand"

const waitWorkflowName = "waitand.yaml"
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

const startWorkflowName = "startand.yaml"
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

const startWorkflowContextName = "startandcontext.yaml"
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

const waitWorkflowContextName = "waitandcontext.yaml"
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

describe('Test workflow events and', () => {
    beforeAll(common.helpers.deleteAllNamespaces)


    it(`should create namespace`, async () => {
        var createNamespaceResponse = await request(common.config.getDirektivHost()).put(`/api/namespaces/${namespaceName}`)
        expect(createNamespaceResponse.statusCode).toEqual(200)
    })

    it(`should have one event listeners`, async () => {

        // workflow with start
        var createWorkflowResponse = await request(common.config.getDirektivHost()).put(`/api/namespaces/${namespaceName}/tree/${startWorkflowName}?op=create-workflow`)
            .send(startEventWorkflow)
        expect(createWorkflowResponse.statusCode).toEqual(200)

        await sleep(1000)
        var getEventListenerResponse = await request(common.config.getDirektivHost()).get(`/api/namespaces/${namespaceName}/event-listeners?limit=8&offset=0`)
            .send()

        expect(getEventListenerResponse.body.results[0]).toMatchObject({
            workflow: "/startand.yaml",
            "mode": "and",
            "instance": "",
            "createdAt": expect.stringMatching(common.regex.timestampRegex),
            "updatedAt": expect.stringMatching(common.regex.timestampRegex),
            "events": [{"type": "eventtype3", "filters": {}}, {"type": "eventtype4", "filters": {}}]
        });

        expect(getEventListenerResponse.body.pageInfo.total).toEqual(1)

    })

    it(`should have two event listeners`, async () => {

        // workflow with start
        var createWorkflowResponse = await request(common.config.getDirektivHost()).put(`/api/namespaces/${namespaceName}/tree/${waitWorkflowName}?op=create-workflow`)
            .send(waitEventWorkflow)
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
            "mode": "and",
            "instance": expect.stringMatching(common.regex.uuidRegex),
            "createdAt": expect.stringMatching(common.regex.timestampRegex),
            "updatedAt": expect.stringMatching(common.regex.timestampRegex),
            "events": [{"type": "eventtype1", "filters": {}}, {"type": "eventtype2", "filters": {}}]
        });


    })

    it(`should kick off in flow workflow`, async () => {

        // should not continue workflow
        await events.sendEventAndList(namespaceName, basevent("eventtype1", "eventtype1", "world1"))

        // the waiting workflow is running but nothing triggered by event, state pending
        var instancesResponse = await events.listInstancesAndFilter(namespaceName, waitWorkflowName, "pending")
        expect(instancesResponse).not.toBeFalsy();

        //  await events.sendEventAndList(namespaceName, basevent("eventtype1", "eventtype1", "world"))
        await events.sendEventAndList(namespaceName, basevent("eventtype2", "eventtype2", "world2"))
        var instancesResponse = await events.listInstancesAndFilter(namespaceName, waitWorkflowName, "complete")
        expect(instancesResponse).not.toBeFalsy();

        var instanceOutput = await request(common.config.getDirektivHost()).get(`/api/namespaces/${namespaceName}/instances/${instancesResponse.id}/output`)
            .send()

        var output = Buffer.from(instanceOutput.body.data, 'base64');
        var outputJSON = JSON.parse(output.toString())

        //  custom value set
        expect(outputJSON["eventtype1"].hello).toEqual("world1")
        expect(outputJSON["eventtype2"].hello).toEqual("world2")

    })


    it(`should kick off start event workflow`, async () => {


        await events.sendEventAndList(namespaceName, basevent("eventtype3", "eventtype3", "world1"))
        var instance = await events.listInstancesAndFilter(namespaceName, startWorkflowName)

        expect(instance).toBeFalsy();

        await events.sendEventAndList(namespaceName, basevent("eventtype4", "eventtype4", "world2"))
        var instance = await events.listInstancesAndFilter(namespaceName, startWorkflowName)

        expect(instance).not.toBeFalsy();


        var instanceOutput = await request(common.config.getDirektivHost()).get(`/api/namespaces/${namespaceName}/instances/${instance.id}/output`)
            .send()

        var output = Buffer.from(instanceOutput.body.data, 'base64');
        var outputJSON = JSON.parse(output.toString())

        // custom data set
        expect(outputJSON["eventtype3"].data.hello).toEqual("world")
        expect(outputJSON["eventtype4"].data.hello).toEqual("world")

    })

    it(`start context`, async () => {

        // timeout workflow
        var createWorkflowResponse = await request(common.config.getDirektivHost()).put(`/api/namespaces/${namespaceName}/tree/${startWorkflowContextName}?op=create-workflow`)
            .send(startEventContextWorkflow)
        expect(createWorkflowResponse.statusCode).toEqual(200)

        await events.sendEventAndList(namespaceName, basevent("eventtype9", "eventtype9", "world1"))
        await events.sendEventAndList(namespaceName, basevent("eventtype10", "eventtype10", "world3"))

        var instance = await events.listInstancesAndFilter(namespaceName, startWorkflowContextName)
        expect(instance).toBeFalsy();

        await events.sendEventAndList(namespaceName, basevent("eventtype10", "eventtype10a", "world2"))

        var instance = await events.listInstancesAndFilter(namespaceName, startWorkflowContextName)
        expect(instance).not.toBeFalsy();

    })


    it(`flow context`, async () => {

        // timeout workflow
        var createWorkflowResponse = await request(common.config.getDirektivHost()).put(`/api/namespaces/${namespaceName}/tree/${waitWorkflowContextName}?op=create-workflow`)
            .send(waitEventContextimeout)
        expect(createWorkflowResponse.statusCode).toEqual(200)


        // start workflow
        var runWorkflowResponse = await request(common.config.getDirektivHost()).post(`/api/namespaces/${namespaceName}/tree/${waitWorkflowContextName}?op=execute`)
            .send()
        expect(runWorkflowResponse.statusCode).toEqual(200)

        await events.sendEventAndList(namespaceName, basevent("eventtype11", "eventtype11", "world1"))
        await events.sendEventAndList(namespaceName, basevent("eventtype12", "eventtype12", "world3"))

        var instance = await events.listInstancesAndFilter(namespaceName, waitWorkflowContextName, "pending")
        expect(instance).not.toBeFalsy();

        await events.sendEventAndList(namespaceName, basevent("eventtype12", "eventtype12a", "world2"))

        var instance = await events.listInstancesAndFilter(namespaceName, startWorkflowContextName, "complete")
        expect(instance).not.toBeFalsy();

    })

})

function sleep(ms) {
    return new Promise(resolve => setTimeout(resolve, ms));
}