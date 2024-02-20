import request from "../common/request"

import common from "../common"

const namespaceName = "datatest"


describe('Test instance data behaviour', () => {
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

    it(`should create a workflow called /data.yaml`, async () => {

        const res = await request(common.config.getDirektivHost())
            .put(`/api/namespaces/${namespaceName}/tree/data.yaml?op=create-workflow`)
            .set({
                'Content-Type': 'text/plain',
            })
            .send(`
states:
- id: a
  type: noop
`)

        expect(res.statusCode).toEqual(200)
        expect(res.body).toMatchObject({
            namespace: namespaceName,
        })
    })

    it(`should invoke the '/data.yaml' workflow with no input`, async () => {
        const req = await request(common.config.getDirektivHost()).get(`/api/namespaces/${namespaceName}/tree/data.yaml?op=wait`)
        expect(req.statusCode).toEqual(200)
        expect(req.body).toMatchObject({})
    })

    it(`should invoke the '/data.yaml' workflow with a simple object input`, async () => {
        const req = await request(common.config.getDirektivHost()).get(`/api/namespaces/${namespaceName}/tree/data.yaml?op=wait`)
        .send(`{"x": 5}`)
        expect(req.statusCode).toEqual(200)
        expect(req.body).toMatchObject({
            x: 5,
        })
    })

    it(`should invoke the '/data.yaml' workflow with a json non-object input`, async () => {
        const req = await request(common.config.getDirektivHost()).get(`/api/namespaces/${namespaceName}/tree/data.yaml?op=wait`)
        .send(`[1, 2, 3]`)
        expect(req.statusCode).toEqual(200)
        expect(req.body).toMatchObject({
            input: [1, 2, 3],
        })
    })

    it(`should invoke the '/data.yaml' workflow with a non-json input`, async () => {
        const req = await request(common.config.getDirektivHost()).get(`/api/namespaces/${namespaceName}/tree/data.yaml?op=wait`)
        .send(`Hello, world!`)
        expect(req.statusCode).toEqual(200)
        expect(req.body).toMatchObject({
            input: "SGVsbG8sIHdvcmxkIQ==",
        })
    })

})