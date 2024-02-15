import request from "../common/request"

import common from "../common"

const namespaceName = "simplelooptest"


describe('Test a simple loop', () => {
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

    it(`should create a workflow called /simple-loop.yaml`, async () => {

        const res = await request(common.config.getDirektivHost())
            .put(`/api/namespaces/${namespaceName}/tree/simple-loop.yaml?op=create-workflow`)
            .set({
                'Content-Type': 'text/plain',
            })
            .send(`
states:
- id: init
  type: noop
  transform:
    i: 5
    result: []
  transition: check
- id: check
  type: switch
  conditions:
  - condition: 'jq(.i)'
    transform: 'jq(.result += [.i] | .i -= 1)'
    transition: check
`)

        expect(res.statusCode).toEqual(200)
        expect(res.body).toMatchObject({
            namespace: namespaceName,
        })
    })

    it(`should invoke the '/simple-loop.yaml' workflow`, async () => {
        const req = await request(common.config.getDirektivHost()).get(`/api/namespaces/${namespaceName}/tree/simple-loop.yaml?op=wait`)
        expect(req.statusCode).toEqual(200)
        expect(req.body).toMatchObject({
            i: 0,
            result: [5, 4, 3, 2, 1],
        })
    })

})