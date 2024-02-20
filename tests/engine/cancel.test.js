import request from "../common/request"

import common from "../common"

const namespaceName = "canceltest"

function sleep(ms) {
    return new Promise(resolve => setTimeout(resolve, ms))
}

describe('Test cancel state behaviour', () => {
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

    it(`should create a workflow called /cancel.yaml`, async () => {

        const res = await request(common.config.getDirektivHost())
            .put(`/api/namespaces/${namespaceName}/tree/cancel.yaml?op=create-workflow`)
            .set({
                'Content-Type': 'text/plain',
            })
            .send(`
states:
- id: a
  type: delay
  duration: PT5S
  transform:
    result: x`)

        expect(res.statusCode).toEqual(200)
        expect(res.body).toMatchObject({
            namespace: namespaceName,
        })
    })

    it(`should invoke the '/cancel.yaml' workflow`, async () => {
        const xreq = await request(common.config.getDirektivHost()).post(`/api/namespaces/${namespaceName}/tree/cancel.yaml?op=execute`)
        expect(xreq.statusCode).toEqual(200)

        var instanceID = xreq.body.instance

        await sleep(50)

        const creq = await request(common.config.getDirektivHost()).post(`/api/namespaces/${namespaceName}/instances/${instanceID}/cancel`)
        expect(creq.statusCode).toEqual(200)

        await sleep(50)

        const ireq = await request(common.config.getDirektivHost()).get(`/api/namespaces/${namespaceName}/instances/${instanceID}`)
        expect(ireq.statusCode).toEqual(200)
        expect(ireq.body.instance.status).toEqual("failed")
        expect(ireq.body.instance.errorCode).toEqual("direktiv.cancels.api")
        expect(ireq.body.instance.errorMessage).toEqual("cancelled by api request")
    })

    it(`should create a workflow called /handle-cancel.yaml`, async () => {

        const res = await request(common.config.getDirektivHost())
            .put(`/api/namespaces/${namespaceName}/tree/handle-cancel.yaml?op=create-workflow`)
            .set({
                'Content-Type': 'text/plain',
            })
            .send(`
states:
- id: a
  type: delay
  duration: PT5S
  catch:
  - error: 'direktiv.cancels.api'
  transform:
    result: x`)

        expect(res.statusCode).toEqual(200)
        expect(res.body).toMatchObject({
            namespace: namespaceName,
        })
    })

    it(`should invoke the '/handle-cancel.yaml' workflow`, async () => {
        const xreq = await request(common.config.getDirektivHost()).post(`/api/namespaces/${namespaceName}/tree/handle-cancel.yaml?op=execute`)
        expect(xreq.statusCode).toEqual(200)

        var instanceID = xreq.body.instance

        await sleep(50)

        const creq = await request(common.config.getDirektivHost()).post(`/api/namespaces/${namespaceName}/instances/${instanceID}/cancel`)
        expect(creq.statusCode).toEqual(200)

        await sleep(50)

        const ireq = await request(common.config.getDirektivHost()).get(`/api/namespaces/${namespaceName}/instances/${instanceID}`)
        expect(ireq.statusCode).toEqual(200)
        expect(ireq.body.instance.status).toEqual("failed")
        expect(ireq.body.instance.errorCode).toEqual("direktiv.cancels.api")
        expect(ireq.body.instance.errorMessage).toEqual("cancelled by api request")
    })

    // TODO: test that a parent can catch a child that was cancelled
    // TODO: test that cancelling a parent recurses down to all children
})