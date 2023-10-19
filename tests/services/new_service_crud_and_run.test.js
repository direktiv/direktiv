import request from 'supertest'
import common from "../common";

const testNamespace = "test-services"

describe('Test services crud operations', () => {
    beforeAll(common.helpers.deleteAllNamespaces)

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