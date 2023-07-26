import request from 'supertest'

import common from "../common"

const namespaceName = "nooptest"


describe('Test noop state behaviour', () => {
    beforeAll(common.helpers.deleteAllNamespaces)
    afterAll(common.helpers.deleteAllNamespaces)

    it(`should create a namespace`, async () => {
        var req = await request(common.config.getDirektivHost()).put(`/api/namespaces/${namespaceName}`)

        expect(req.statusCode).toEqual(200)
        expect(req.body).toMatchObject({
            namespace: {
                createdAt: expect.stringMatching(common.regex.timestampRegex),
                updatedAt: expect.stringMatching(common.regex.timestampRegex),
                name: namespaceName,
                oid: expect.stringMatching(common.regex.uuidRegex),
            },
        })
    })

    it(`should create a workflow called /noop.yaml`, async () => {

        const res = await request(common.config.getDirektivHost())
            .put(`/api/namespaces/${namespaceName}/tree/noop.yaml?op=create-workflow`)
            .set({
                'Content-Type': 'text/plain',
            })
            .send(`
states:
- id: a
  type: noop
  transform:
    result: x`)

        expect(res.statusCode).toEqual(200)
        expect(res.body).toMatchObject({
            namespace: namespaceName,
        })
    })

    it(`should invoke the '/noop.yaml' workflow`, async () => {
        const req = await request(common.config.getDirektivHost()).get(`/api/namespaces/${namespaceName}/tree/noop.yaml?op=wait`)
        expect(req.statusCode).toEqual(200)
        expect(req.body).toMatchObject({
            result: 'x',
        })
    })

})