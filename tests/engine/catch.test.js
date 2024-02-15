import request from "../common/request"

import common from "../common"

const namespaceName = "catchtest"


describe('Test catch state behaviour', () => {
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

    it(`should create a workflow called /catch.yaml`, async () => {

        const res = await request(common.config.getDirektivHost())
            .put(`/api/namespaces/${namespaceName}/tree/catch.yaml?op=create-workflow`)
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
  catch:
  - error: 'testcode'
`)

        expect(res.statusCode).toEqual(200)
        expect(res.body).toMatchObject({
            namespace: namespaceName,
        })
    })

    it(`should invoke the '/catch.yaml' workflow`, async () => {
        const req = await request(common.config.getDirektivHost()).get(`/api/namespaces/${namespaceName}/tree/catch.yaml?op=wait`)
        expect(req.statusCode).toEqual(200)
        expect(req.body).toMatchObject({
            error: {
                code: "testcode",
                msg: "this is a test error"
            },
        })
    })

})