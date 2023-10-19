import request from 'supertest'
import common from "../common";

const testNamespace = "test-services"

beforeAll(async () => {
    // delete a 'test-namespace' if it's already exit.
    await request(common.config.getDirektivHost()).delete(`/api/namespaces/${testNamespace}?recursive=true`)
});

describe('Test services crud operations', () => {
    common.helpers.itShouldCreateNamespace(it, expect, testNamespace)

    it(`should create a new service file`, async () => {
        const res = await request(common.config.getDirektivHost())
            .put(`/api/namespaces/${testNamespace}/tree/my-workflow.yaml?op=create-workflow`)
            .set({
                'Content-Type': 'text/plain',
            })

            .send(`
direktiv_api: service/v1
name: s1
image: redis
cmd: redis-server
scale: 2
`)

        expect(res.statusCode).toEqual(200)
        expect(res.body).toMatchObject({
            namespace: testNamespace,
        })
    })

});


function sleep(ms) {
    return new Promise(resolve => setTimeout(resolve, ms));
}