import request from 'supertest'

import common from "../common"


const namespaceName = "vars"

const workflowName = "wf.yaml"

const simpleWorkflow = `
states:
- id: hello
  type: getter
  variables:
  - key: plain
    scope: workflow
  - key: json
    scope: workflow
  - key: binary
    scope: workflow
  transition: value

- id: value
  type: noop
  transform:
    binary: jq(.var.binary)
    text: jq(.var.plain | @base64d)
    json: jq(.var.json)
`

const binData = "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"

const plainText = `this is plain text`

const jsonData = `{ "hello": "world" }`


const jdata = JSON.stringify(JSON.parse(jsonData))


describe('Test workflow variable operations', () => {
    beforeAll(common.helpers.deleteAllNamespaces)


    it(`should create a namespace`, async () => {
        var createNamespaceResponse = await request(common.config.getDirektivHost()).put(`/api/namespaces/${namespaceName}`)
        expect(createNamespaceResponse.statusCode).toEqual(200)
    })

    it(`should create a workflow`, async () => {
        var createWorkflowResponse = await request(common.config.getDirektivHost()).put(`/api/namespaces/${namespaceName}/tree/${workflowName}?op=create-workflow`)
            .send(simpleWorkflow)

        expect(createWorkflowResponse.statusCode).toEqual(200)
        var buf = Buffer.from(createWorkflowResponse.body.source, 'base64')
        expect(buf.toString()).toEqual(simpleWorkflow)
    })

    it(`should fail invalid name`, async () => {
        var workflowVarResponse = await request(common.config.getDirektivHost()).put(`/api/namespaces/${namespaceName}/tree/${workflowName}?op=set-var&var=hel$$o`)
            .set('Content-Type', 'application/json')
            .send(jdata)

        expect(workflowVarResponse.statusCode).toEqual(406)
    })


    it(`should set plain text variable`, async () => {
        var workflowVarResponse = await request(common.config.getDirektivHost()).put(`/api/namespaces/${namespaceName}/tree/${workflowName}?op=set-var&var=plain`)
            .set('Content-Type', 'text/plain')
            .send(plainText)

        expect(workflowVarResponse.statusCode).toEqual(200)
        expect(workflowVarResponse.body.key).toEqual("plain")
        expect(workflowVarResponse.body.totalSize).toEqual(plainText.length.toString())

    })


    it(`should set json variable`, async () => {
        var workflowVarResponse = await request(common.config.getDirektivHost()).put(`/api/namespaces/${namespaceName}/tree/${workflowName}?op=set-var&var=json`)
            .set('Content-Type', 'application/json')
            .send(jdata)

        expect(workflowVarResponse.statusCode).toEqual(200)
        expect(workflowVarResponse.body.key).toEqual("json")
        expect(workflowVarResponse.body.totalSize).toEqual(jdata.length.toString())
    })

    it(`should set binary variable`, async () => {

        var buf = Buffer.from(binData, 'base64')

        var workflowVarResponse = await request(common.config.getDirektivHost()).put(`/api/namespaces/${namespaceName}/tree/${workflowName}?op=set-var&var=binary`)
            .set('Content-Type', 'image/png')
            .send(buf)

        expect(workflowVarResponse.statusCode).toEqual(200)
        expect(workflowVarResponse.body.key).toEqual("binary")
        expect(workflowVarResponse.body.totalSize).toEqual(Buffer.byteLength(buf).toString())
    })


    it(`should list all variable`, async () => {
        var workflowVarResponse = await request(common.config.getDirektivHost()).get(`/api/namespaces/${namespaceName}/tree/${workflowName}?op=vars`)

        expect(workflowVarResponse.statusCode).toEqual(200)
        expect(workflowVarResponse.body.variables.results.length).toEqual(3)
    })


    it(`should get json variable`, async () => {
        var workflowVarResponse = await request(common.config.getDirektivHost()).get(`/api/namespaces/${namespaceName}/tree/${workflowName}?op=var&var=json`)

        expect(workflowVarResponse.statusCode).toEqual(200)
        expect(workflowVarResponse.body).toEqual(JSON.parse(jsonData))
    })

    it(`should get text variable`, async () => {
        var workflowVarResponse = await request(common.config.getDirektivHost()).get(`/api/namespaces/${namespaceName}/tree/${workflowName}?op=var&var=plain`)

        expect(workflowVarResponse.statusCode).toEqual(200)
        expect(workflowVarResponse.res.text).toEqual(plainText)
    })

    it(`should get binary variable`, async () => {
        var workflowVarResponse = await request(common.config.getDirektivHost()).get(`/api/namespaces/${namespaceName}/tree/${workflowName}?op=var&var=binary`)

        var buf = Buffer.from(workflowVarResponse.body).toString('base64')

        expect(workflowVarResponse.statusCode).toEqual(200)
        expect(buf).toEqual(binData)
    })


    it(`should get variables from workflow getter`, async () => {
        var workflowVarResponse = await request(common.config.getDirektivHost()).post(`/api/namespaces/${namespaceName}/tree/${workflowName}?op=wait&ref=latest`)

        expect(workflowVarResponse.statusCode).toEqual(200)
        expect(workflowVarResponse.body.text).toEqual(plainText)
        expect(workflowVarResponse.body.json).toEqual(JSON.parse(jsonData))
        expect(workflowVarResponse.body.binary).toEqual(binData)

    })

    it(`should delete one variable`, async () => {
        var workflowVarResponse = await request(common.config.getDirektivHost()).delete(`/api/namespaces/${namespaceName}/tree/${workflowName}?op=delete-var&var=json`)

        expect(workflowVarResponse.statusCode).toEqual(200)
    })

    it(`should have less variables`, async () => {
        var workflowVarListResponse = await request(common.config.getDirektivHost()).get(`/api/namespaces/${namespaceName}/tree/${workflowName}?op=vars`)

        expect(workflowVarListResponse.statusCode).toEqual(200)
        expect(workflowVarListResponse.body.variables.results.length).toEqual(2)

    })


})