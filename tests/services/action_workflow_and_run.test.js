import request from "../common/request"
import common from "../common";

const testNamespace = "test-services"
const testWorkflow = "test-workflow.yaml"

describe('Test workflow function invoke', () => {
    beforeAll(common.helpers.deleteAllNamespaces)

    it(`should create a new namespace`, async () => {
        const res = await request(common.config.getDirektivHost()).put(`/api/namespaces/${testNamespace}`)
        expect(res.statusCode).toEqual(200)
        expect(res.body).toMatchObject({
            namespace: {
                name: testNamespace,
                // regex /^2.*Z$/ matches timestamps like 2023-03-01T14:19:52.383871512Z
                createdAt: expect.stringMatching(/^2.*Z$/),
                updatedAt: expect.stringMatching(/^2.*Z$/),
            }
        })
    })

    it(`should create a workflow /${testWorkflow} to invoke the a function`, async () => {
        const res = await request(common.config.getDirektivHost())
            .put(`/api/namespaces/${testNamespace}/tree/${testWorkflow}?op=create-workflow`)
            .set({
                'Content-Type': 'text/plain',
            })

            .send(`
description: A simple 'action' state that sends a get request
functions:
- id: get
  image: direktiv/request:v4
  type: knative-workflow
states:
- id: getter 
  type: action
  action:
    function: get
    input: 
      method: "GET"
      url: "https://jsonplaceholder.typicode.com/todos/1"
`)

        expect(res.statusCode).toEqual(200)
        expect(res.body).toMatchObject({
            namespace: testNamespace,
        })
    })

    it(`should invoke the ${testWorkflow} workflow`, async () => {
        await sleep(500);
        const res = await request(common.config.getDirektivHost()).get(`/api/namespaces/${testNamespace}/tree/${testWorkflow}?op=wait`)
        expect(res.statusCode).toEqual(200)
        expect(res.body.return.status).toBe("200 OK")
    })
});

function sleep(ms) {
    return new Promise(resolve => setTimeout(resolve, ms));
}