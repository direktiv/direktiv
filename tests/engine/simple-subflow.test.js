import request from 'supertest'

import common from "../common"

const namespaceName = "simplesubflowtest"


describe('Test subflow behaviour', () => {
    beforeAll(common.helpers.deleteAllNamespaces)

    it(`should create a namespace`, async () => {
        var req = await request(common.config.getDirektivHost()).put(`/api/namespaces/${namespaceName}`)

        expect(req.statusCode).toEqual(200)
        expect(req.body).toMatchObject({
            namespace: {
                createdAt: expect.stringMatching(common.regex.timestampRegex),
                updatedAt: expect.stringMatching(common.regex.timestampRegex),
                name: namespaceName,
            },
        })
    })

    it(`should create a directory called /a`, async () => {
        var createDirectoryResponse = await request(common.config.getDirektivHost()).put(`/api/namespaces/${namespaceName}/tree/a?op=create-directory`)
        expect(createDirectoryResponse.statusCode).toEqual(200)
    })

    it(`should create a workflow called /a/child.yaml`, async () => {

        const res = await request(common.config.getDirektivHost())
            .put(`/api/namespaces/${namespaceName}/tree/a/child.yaml?op=create-workflow`)
            .set({
                'Content-Type': 'text/plain',
            })
            .send(`
states:
- id: a
  type: noop
  transform:
    result: 'jq(.input + 1)'`)

        expect(res.statusCode).toEqual(200)
        expect(res.body).toMatchObject({
            namespace: namespaceName,
        })
    })

    it(`should create a workflow called /a/parent1.yaml`, async () => {

        const res = await request(common.config.getDirektivHost())
            .put(`/api/namespaces/${namespaceName}/tree/a/parent1.yaml?op=create-workflow`)
            .set({
                'Content-Type': 'text/plain',
            })
            .send(`
functions:
- id: child
  type: subflow
  workflow: '/a/child.yaml'
states:
- id: a
  type: action
  action:
    function: child
    input: 
      input: 1
  transform:
    result: 'jq(.return.result)'
`)

        expect(res.statusCode).toEqual(200)
        expect(res.body).toMatchObject({
            namespace: namespaceName,
        })
    })

    it(`should invoke the '/a/parent1.yaml' workflow`, async () => {
        const req = await request(common.config.getDirektivHost()).get(`/api/namespaces/${namespaceName}/tree/a/parent1.yaml?op=wait`)
        expect(req.statusCode).toEqual(200)
        expect(req.body).toMatchObject({
            result: 2,
        })
    })

    it(`should create a workflow called /a/parent2.yaml`, async () => {

        const res = await request(common.config.getDirektivHost())
            .put(`/api/namespaces/${namespaceName}/tree/a/parent2.yaml?op=create-workflow`)
            .set({
                'Content-Type': 'text/plain',
            })
            .send(`
functions:
- id: child
  type: subflow
  workflow: 'child.yaml'
states:
- id: a
  type: action
  action:
    function: child
    input: 
      input: 1
  transform:
    result: 'jq(.return.result)'
`)

        expect(res.statusCode).toEqual(200)
        expect(res.body).toMatchObject({
            namespace: namespaceName,
        })
    })

    it(`should invoke the '/a/parent2.yaml' workflow`, async () => {
        const req = await request(common.config.getDirektivHost()).get(`/api/namespaces/${namespaceName}/tree/a/parent2.yaml?op=wait`)
        expect(req.statusCode).toEqual(200)
        expect(req.body).toMatchObject({
            result: 2,
        })
    })

    it(`should create a workflow called /a/parent3.yaml`, async () => {

        const res = await request(common.config.getDirektivHost())
            .put(`/api/namespaces/${namespaceName}/tree/a/parent3.yaml?op=create-workflow`)
            .set({
                'Content-Type': 'text/plain',
            })
            .send(`
functions:
- id: child
  type: subflow
  workflow: './child.yaml'
states:
- id: a
  type: action
  action:
    function: child
    input: 
      input: 1
  transform:
    result: 'jq(.return.result)'
`)

        expect(res.statusCode).toEqual(200)
        expect(res.body).toMatchObject({
            namespace: namespaceName,
        })
    })

    it(`should invoke the '/a/parent3.yaml' workflow`, async () => {
        const req = await request(common.config.getDirektivHost()).get(`/api/namespaces/${namespaceName}/tree/a/parent3.yaml?op=wait`)
        expect(req.statusCode).toEqual(200)
        expect(req.body).toMatchObject({
            result: 2,
        })
    })

    it(`should create a workflow called /a/parent4.yaml`, async () => {

        const res = await request(common.config.getDirektivHost())
            .put(`/api/namespaces/${namespaceName}/tree/a/parent4.yaml?op=create-workflow`)
            .set({
                'Content-Type': 'text/plain',
            })
            .send(`
functions:
- id: child
  type: subflow
  workflow: '../a/child.yaml'
states:
- id: a
  type: action
  action:
    function: child
    input: 
      input: 1
  transform:
    result: 'jq(.return.result)'
`)

        expect(res.statusCode).toEqual(200)
        expect(res.body).toMatchObject({
            namespace: namespaceName,
        })
    })

    it(`should invoke the '/a/parent4.yaml' workflow`, async () => {
        const req = await request(common.config.getDirektivHost()).get(`/api/namespaces/${namespaceName}/tree/a/parent4.yaml?op=wait`)
        expect(req.statusCode).toEqual(200)
        expect(req.body).toMatchObject({
            result: 2,
        })
    })

    it(`check if instances are present`, async () => {
        var instances = await request(common.config.getDirektivHost()).get(`/api/namespaces/${namespaceName}/instances`)
        expect(instances.statusCode).toEqual(200)
        expect(instances.body.instances.results.length).not.toBeLessThan(1)
    })


    it(`check if instance logs are present`, async () => {
        var instances = await request(common.config.getDirektivHost()).get(`/api/namespaces/${namespaceName}/instances`)
        // create a new array containing the ids
        const ids = instances.body.instances.results.map((result) => result.id)

        // iterate over that array 
        await Promise.all(ids.map(async (id) => {
            const logsResponse = await request(common.config.getDirektivHost()).get(`/api/namespaces/${namespaceName}/instances/${id}/logs`)
            expect(logsResponse.statusCode).toEqual(200)
            expect(logsResponse.body.results.length).not.toBeLessThan(1)
        }))
    })

    it(`check if namespace logs contains some workflow operations`, async () => {
        var logsResponse = await request(common.config.getDirektivHost()).get(`/api/namespaces/${namespaceName}/logs`)
        expect(logsResponse.statusCode).toEqual(200)
        expect(logsResponse.body.results.length).not.toBeLessThan(1)
    })

})
