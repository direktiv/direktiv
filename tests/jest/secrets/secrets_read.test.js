import request from 'supertest'
import common from "../common";

const testNamespace = "test-secrets-namespace"
let testWorkflow = "test-secret.yaml"

beforeAll(async () => {
    // delete a 'test-namespace' if it's already exit.
    await request(common.config.getDirektivHost()).delete(`/api/namespaces/${testNamespace}?recursive=true`)
});

describe('Test secret read operations', () => {
    it(`should create a new namespace`, async () => {
        const res = await request(common.config.getDirektivHost()).put(`/api/namespaces/${testNamespace}`)
        expect(res.statusCode).toEqual(200)
        expect(res.body).toMatchObject({
            namespace: {
                name: testNamespace,
                oid: "",
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

    it(`should create a workflow to test secret read`, async () => {

        const res = await request(common.config.getDirektivHost())
            .put(`/api/namespaces/${testNamespace}/tree/${testWorkflow}?op=create-workflow`)
            .set({
                'Content-Type': 'text/plain',
            })

            // TODO: Alan, please write a workflow that test reading 'key1' secret ant print it.
            .send(`
description: A simple that sould test secret read'
states:
- id: helloworld
  type: noop
  transform:
    result: value1`)

        expect(res.statusCode).toEqual(200)
        expect(res.body).toMatchObject({
            namespace: testNamespace,
        })
    })

    it(`should invoke the '/${testWorkflow}' workflow`, async () => {
        const req = await request(common.config.getDirektivHost()).get(`/api/namespaces/${testNamespace}/tree/${testWorkflow}?op=wait`)
        expect(req.statusCode).toEqual(200)
        expect(req.body).toMatchObject({
            result: 'value1',
        })
    })
});

