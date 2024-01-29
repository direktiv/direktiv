import request from 'supertest'

import common from "../common"

const namespaceName = "errorstatetest"


describe('Test error state behaviour', () => {
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

    it(`should create a workflow called /error.yaml`, async () => {

        const res = await request(common.config.getDirektivHost())
            .put(`/api/namespaces/${namespaceName}/tree/error.yaml?op=create-workflow`)
            .set({
                'Content-Type': 'text/plain',
            })
            .send(`
states:
- id: a
  type: error
  error: testcode
  message: 'this is a test error'
  transform: 
    result: x
`)

        expect(res.statusCode).toEqual(200)
        expect(res.body).toMatchObject({
            namespace: namespaceName,
        })
    })

    it(`should invoke the '/error.yaml' workflow`, async () => {
        const req = await request(common.config.getDirektivHost()).get(`/api/namespaces/${namespaceName}/tree/error.yaml?op=wait`)

        expect(req.statusCode).toEqual(500)
        expect(req.headers["direktiv-instance-error-code"]).toEqual('testcode')
        expect(req.headers["direktiv-instance-error-message"]).toEqual('this is a test error')
        expect(req.body).toMatchObject({})
    })

    it(`should create a workflow called /caller.yaml`, async () => {

        const res = await request(common.config.getDirektivHost())
            .put(`/api/namespaces/${namespaceName}/tree/caller.yaml?op=create-workflow`)
            .set({
                'Content-Type': 'text/plain',
            })
            .send(`
functions:
- id: child
  type: subflow
  workflow: '/error.yaml'
states:
- id: a
  type: action
  action:
    function: child
`)

        expect(res.statusCode).toEqual(200)
        expect(res.body).toMatchObject({
            namespace: namespaceName,
        })
    })

    it(`should invoke the '/caller.yaml' workflow`, async () => {
        const req = await request(common.config.getDirektivHost()).get(`/api/namespaces/${namespaceName}/tree/caller.yaml?op=wait`)

        expect(req.statusCode).toEqual(500)
        expect(req.headers["direktiv-instance-error-code"]).toEqual('testcode')
        expect(req.headers["direktiv-instance-error-message"]).toEqual('this is a test error')
        expect(req.body).toMatchObject({})
    })

    it(`should create a workflow called /error-and-continue.yaml`, async () => {

        const res = await request(common.config.getDirektivHost())
            .put(`/api/namespaces/${namespaceName}/tree/error-and-continue.yaml?op=create-workflow`)
            .set({
                'Content-Type': 'text/plain',
            })
            .send(`
states:
- id: a
  type: error
  error: testcode
  message: 'this is a test error'
  transition: b
- id: b
  type: noop
`)

        expect(res.statusCode).toEqual(200)
        expect(res.body).toMatchObject({
            namespace: namespaceName,
        })
    })

    it(`should invoke the '/error-and-continue.yaml' workflow`, async () => {
        const req = await request(common.config.getDirektivHost()).get(`/api/namespaces/${namespaceName}/tree/error-and-continue.yaml?op=wait`)

        expect(req.statusCode).toEqual(500)
        expect(req.headers["direktiv-instance-error-code"]).toEqual('testcode')
        expect(req.headers["direktiv-instance-error-message"]).toEqual('this is a test error')
        expect(req.body).toMatchObject({})
    })

    it(`should create a workflow called /double-error.yaml`, async () => {

        const res = await request(common.config.getDirektivHost())
            .put(`/api/namespaces/${namespaceName}/tree/double-error.yaml?op=create-workflow`)
            .set({
                'Content-Type': 'text/plain',
            })
            .send(`
states:
- id: a
  type: error
  error: testcode
  message: 'this is a test error'
  transition: b
- id: b
  type: error
  error: testcode2
  message: 'this is a test error 2'
`)

        expect(res.statusCode).toEqual(200)
        expect(res.body).toMatchObject({
            namespace: namespaceName,
        })
    })

    // it(`should invoke the '/double-error.yaml' workflow`, async () => {
    //     const req = await request(common.config.getDirektivHost()).get(`/api/namespaces/${namespaceName}/tree/double-error.yaml?op=wait`)
    //
    //     expect(req.statusCode).toEqual(500)
    //     expect(req.headers["direktiv-instance-error-code"]).toEqual('direktiv.workflow.multipleErrors')
    //     expect(req.headers["direktiv-instance-error-message"]).toEqual('the workflow instance tried to throw multiple errors')
    //     expect(req.body).toMatchObject({})
    // })

})