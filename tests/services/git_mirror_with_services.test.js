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
        expect(listRes.body).toMatchObject({
            data: [
                {
                    error: null,
                    id: 'git-test-services-aws-cli-aws-aws-run-instance-yam-4843e86e6a',
                },
                {
                    error: null,
                    id: 'git-test-services-s3-aws-aws-s3-upload-yaml-d68a65802c',
                },
                {
                    error: null,
                    id: 'git-test-services-envs-wf-svc-yaml-1a4b341b98',
                },
                {
                    error: null,
                    id: 'git-test-services-bash-envs-wf-wf-yaml-244eb0cc07',
                },
                {
                    error: null,
                    id: 'git-test-services-echo-foreach-foreach-jq-yaml-30d6b9dff8',
                },
                {
                    error: null,
                    id: 'git-test-services-echo-foreach-foreach-js-yaml-3d6e83ff12',
                },
                {
                    error: null,
                    id: 'git-test-services-echo-foreach-printer-yaml-1df588f6a6',
                },
                {
                    error: null,
                    id: 'git-test-services-gcp-gcp-vm-destroy-create-yaml-63c9914a0c',
                },
                {
                    error: null,
                    id: 'git-test-services-gcp-gcp-vm-destroy-deleter-yaml-d2e0568517',
                },
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
                {
                    error: null,
                    id: 'git-test-services-http-request-request-external-ap-d81abd315a',
                },
                {
                    error: null,
                    id: 'git-test-services-python-scripting-file-yaml-eb8d7c4890',
                },
                {
                    error: null,
                    id: 'git-test-services-services-s1-yaml-ebcc7ad135',
                },
                {
                    error: null,
                    id: 'git-test-services-solve-math-expression-solving-ma-f242d276b6',
                },
                {
                    error: null,
                    id: 'git-test-services-bash-variables-workflow-scope-ya-8c6b60ad75',
                }
            ]
        })
    })
});

function sleep(ms) {
    return new Promise(resolve => setTimeout(resolve, ms));
}