import request from 'supertest'
import retry from "jest-retries";
import common from "../common";

const testNamespace = "git-test-services"

describe('Test services crud operations', () => {
    beforeAll(common.helpers.deleteAllNamespaces)

    it(`should create a new git mirrored namespace`, async () => {
        const res = await request(common.config.getDirektivHost())
            .put(`/api/namespaces/${testNamespace}`)
            .send({
                url: "https://github.com/direktiv/direktiv-examples.git",
                ref: "main",
                cron: "",
                passphrase: "",
                publicKey: "",
                privateKey: ""
            })
        expect(res.statusCode).toEqual(200)
    })

    retry(`should list all services`, 10, async () => {
        await sleep(500)
        const listRes = await request(common.config.getDirektivHost())
            .get(`/api/v2/namespaces/${testNamespace}/services`)
        expect(listRes.statusCode).toEqual(200)

        let reduced = listRes.body.data.map(item => {return {id: item.id, error: item.error}})

        expect(reduced).toEqual(expect.arrayContaining([
            {
                error: null,
                id: 'git-test-services-hello-world-greeting-event-liste-6acf6e6da3',
            },
            {
                error: null,
                id: 'git-test-services-greeter-greeting-greeting-yaml-a09fc061bb',
            },
            {
                error: null,
                id: 'git-test-services-csvkit-input-convert-workflow-ya-6c50acea98',
            },
            {
                error: expect.anything(),
                id: 'git-test-services-build-patching-wf-build-yaml-6909196d31',
            },
            {
                error: null,
                id: 'git-test-services-patch-patching-wf-yaml-f1cd98cbce',
            },
        ]))
    })
});

function sleep(ms) {
    return new Promise(resolve => setTimeout(resolve, ms));
}