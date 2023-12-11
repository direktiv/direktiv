import common from "../common";
import request from 'supertest'

const testNamespace = "test-secrets-namespace"
let testWorkflow = "test-secret"

describe('Test secret read operations', () => {
    beforeAll(common.helpers.deleteAllNamespaces)

    it(`should create a new namespace`, async () => {
        const res = await request(common.config.getDirektivHost()).put(`/api/namespaces/${testNamespace}`)
        expect(res.statusCode).toEqual(200)
        expect(res.body).toMatchObject({
            namespace: {
                name: testNamespace,
                oid: expect.stringMatching(common.regex.uuidRegex),
                // regex /^2.*Z$/ matches timestamps like 2023-03-01T14:19:52.383871512Z
                createdAt: expect.stringMatching(/^2.*Z$/),
                updatedAt: expect.stringMatching(/^2.*Z$/),
            }
        })
    })

    it(`should create a new secret`, async () => {
        const res = await request(common.config.getDirektivHost())
            .put(`/api/namespaces/${testNamespace}/secrets/key1`)
            .set({
                'Content-Type': 'text/plain',
            })

            .send(`value1`)
        expect(res.statusCode).toEqual(200)
        expect(res.body).toMatchObject({
            namespace: testNamespace,
            key: 'key1'
        })
    })

    it(`should create the first workflow in a pair to test secret read`, async () => {

        const res = await request(common.config.getDirektivHost())
        .put(`/api/namespaces/${testNamespace}/tree/${testWorkflow}-parent.yaml?op=create-workflow`)
        .set({
            'Content-Type': 'text/plain',
        })

        .send(`
functions:
- id: echo
  workflow: ${testWorkflow}-child.yaml
  type: subflow
states:
- id: echo
  type: action
  action:
    function: echo
    secrets: [key1]
    input: 
      secret: 'jq(.secrets.key1)'
  transform: 
    result: 'jq(.return.secret)'
`)

    expect(res.statusCode).toEqual(200)
    expect(res.body).toMatchObject({
        namespace: testNamespace,
    })
})

    it(`should create the second workflow in a pair to test secret read`, async () => {

        const res = await request(common.config.getDirektivHost())
            .put(`/api/namespaces/${testNamespace}/tree/${testWorkflow}-child.yaml?op=create-workflow`)
            .set({
                'Content-Type': 'text/plain',
            })

            .send(`
states:
- id: helloworld
  type: noop
`)

        expect(res.statusCode).toEqual(200)
        expect(res.body).toMatchObject({
            namespace: testNamespace,
        })
    })

    it(`should invoke the '/${testWorkflow}-parent.yaml' workflow`, async () => {
        const req = await request(common.config.getDirektivHost()).get(`/api/namespaces/${testNamespace}/tree/${testWorkflow}-parent.yaml?op=wait`)
        expect(req.statusCode).toEqual(200)
        expect(req.body).toMatchObject({
            result: 'value1',
        })
    })
});

