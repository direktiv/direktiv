import request from 'supertest'
import common from "../common";

const testNamespace = "test-services"
const testWorkflow = "test-workflow.yaml"

beforeAll(async () => {
    // delete a 'test-namespace' if it's already exit.
    await request(common.config.getDirektivHost()).delete(`/api/namespaces/${testNamespace}?recursive=true`)
});

describe('Test services read operations', () => {
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

    it(`should create a new service`, async () => {
        const res = await request(common.config.getDirektivHost())
            .post(`/api/functions/namespaces/${testNamespace}`)
            .send({
                cmd: "",
                image: "direktiv/request",
                minScale: 1,
                name: "mysvc1",
                size: 1
            })
        expect(res.statusCode).toEqual(200)
        expect(res.body).toEqual({})
    })

    it(`should create a second service`, async () => {
        const res = await request(common.config.getDirektivHost())
            .post(`/api/functions/namespaces/${testNamespace}`)
            .send({
                cmd: "",
                image: "direktiv/request",
                minScale: 1,
                name: "mysvc2",
                size: 1
            })
        expect(res.statusCode).toEqual(200)
        expect(res.body).toEqual({})
    })

    it(`should list all services`, async () => {
        const res = await request(common.config.getDirektivHost())
            .get(`/api/functions/namespaces/${testNamespace}`)
        expect(res.statusCode).toEqual(200)
        expect(res.body).toMatchObject({
            config: { maxscale: 3 },
            functions: expect.arrayContaining([
                {
                    conditions: expect.anything(),
                    info: expect.objectContaining({
                        "name": "mysvc1",
                        "image": "direktiv/request",
                        "cmd": "",
                        "size": 1,
                        "minScale": 1,
                        "namespace": expect.stringMatching(common.regex.uuidRegex),
                        "namespaceName": testNamespace,
                        "workflow": "",
                        "path": "",
                        "revision": "",
                        "envs": {}
                    }),
                    serviceName: expect.anything(),
                    status: expect.anything(),
                }
            ])
        })
    })

    it(`should all services have status true`, async () => {
        let res;
        for(let i=1; i<10; i++) {
            await sleep(1000);
            res = await request(common.config.getDirektivHost())
                .get(`/api/functions/namespaces/${testNamespace}`)
            if(res.statusCode !== 200) {
                break;
            }
            if (res.body.functions[0].status === "Unknown") {
                continue;
            }
            if (res.body.functions[1].status === "Unknown") {
                continue;
            }

            break;
        }

        expect(res.statusCode).toEqual(200)
        expect(res.body.functions[0].status).toBe("True")
        expect(res.body.functions[1].status).toBe("True")
    })


    it(`should create a workflow /${testWorkflow} to invoke the service`, async () => {
        const res = await request(common.config.getDirektivHost())
            .put(`/api/namespaces/${testNamespace}/tree/${testWorkflow}?op=create-workflow`)
            .set({
                'Content-Type': 'text/plain',
            })

            .send(`
description: A simple 'action' state that test services
functions:
- id: get
  service: mysvc1
  type: knative-namespace
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
        const res = await request(common.config.getDirektivHost()).get(`/api/namespaces/${testNamespace}/tree/${testWorkflow}?op=wait`)
        expect(res.statusCode).toEqual(200)
        expect(res.body.return.status).toBe("200 OK")
    })
});


function sleep(ms) {
    return new Promise(resolve => setTimeout(resolve, ms));
}