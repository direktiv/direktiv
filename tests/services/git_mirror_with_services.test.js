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
                    type: 'workflow-service',
                    namespace: 'git-test-services',
                    filePath: '/aws/aws-run-instance.yaml',
                    name: 'aws-cli',
                    image: 'direktiv/aws-cli:dev',
                    id: 'git-test-services-aws-cli-aws-aws-run-instance-yam-4843e86e6a',
                    cmd: "",
                    scale: 0,
                    size: "small",
                    error: null,
                    conditions: expect.anything(),
                },
                {
                    type: 'workflow-service',
                    namespace: 'git-test-services',
                    filePath: '/aws/aws-s3-upload.yaml',
                    name: 's3',
                    image: 'direktiv/aws-cli:dev',
                    id: 'git-test-services-s3-aws-aws-s3-upload-yaml-d68a65802c',
                    cmd: "",
                    scale: 0,
                    size: "small",
                    error: null,
                    conditions: expect.anything(),
                },
                {
                    type: 'namespace-service',
                    namespace: 'git-test-services',
                    filePath: '/envs-wf/svc.yaml',
                    name: '',
                    image: 'gcr.io/direktiv/functions/http-request:1.0',
                    id: 'git-test-services-envs-wf-svc-yaml-1a4b341b98',
                    cmd: "",
                    scale: 0,
                    size: "small",
                    error: null,
                    conditions: expect.anything(),
                },
                {
                    type: 'workflow-service',
                    namespace: 'git-test-services',
                    filePath: '/envs-wf/wf.yaml',
                    name: 'bash',
                    image: 'gcr.io/direktiv/functions/bash:1.0',
                    id: 'git-test-services-bash-envs-wf-wf-yaml-244eb0cc07',
                    cmd: "",
                    scale: 0,
                    size: "small",
                    error: null,
                    conditions: expect.anything(),
                },
                {
                    type: 'workflow-service',
                    namespace: 'git-test-services',
                    filePath: '/foreach/foreach-jq.yaml',
                    name: 'echo',
                    image: 'direktiv/echo:dev',
                    id: 'git-test-services-echo-foreach-foreach-jq-yaml-30d6b9dff8',
                    cmd: "",
                    scale: 0,
                    size: "small",
                    error: null,
                    conditions: expect.anything(),
                },
                {
                    type: 'workflow-service',
                    namespace: 'git-test-services',
                    filePath: '/foreach/foreach-js.yaml',
                    name: 'echo',
                    image: 'direktiv/echo:dev',
                    id: 'git-test-services-echo-foreach-foreach-js-yaml-3d6e83ff12',
                    cmd: "",
                    scale: 0,
                    size: "small",
                    error: null,
                    conditions: expect.anything(),
                },
                {
                    type: 'workflow-service',
                    namespace: 'git-test-services',
                    filePath: '/foreach/printer.yaml',
                    name: 'echo',
                    image: 'direktiv/echo:dev',
                    id: 'git-test-services-echo-foreach-printer-yaml-1df588f6a6',
                    cmd: "",
                    scale: 0,
                    size: "small",
                    error: null,
                    conditions: expect.anything(),
                },
                {
                    type: 'workflow-service',
                    namespace: 'git-test-services',
                    filePath: '/gcp-vm-destroy/create.yaml',
                    name: 'gcp',
                    image: 'gcr.io/direktiv/functions/gcp:1.0',
                    id: 'git-test-services-gcp-gcp-vm-destroy-create-yaml-63c9914a0c',
                    cmd: "",
                    scale: 0,
                    size: "small",
                    error: null,
                    conditions: expect.anything(),
                },
                {
                    type: 'workflow-service',
                    namespace: 'git-test-services',
                    filePath: '/gcp-vm-destroy/deleter.yaml',
                    name: 'gcp',
                    image: 'gcr.io/direktiv/functions/gcp:1.0',
                    id: 'git-test-services-gcp-gcp-vm-destroy-deleter-yaml-d2e0568517',
                    cmd: "",
                    scale: 0,
                    size: "small",
                    error: null,
                    conditions: expect.anything(),
                },
                {
                    type: 'workflow-service',
                    namespace: 'git-test-services',
                    filePath: '/greeting-event-listener/greeting-listener.yaml',
                    name: 'hello-world',
                    image: 'direktiv/hello-world:dev',
                    id: 'git-test-services-hello-world-greeting-event-liste-6acf6e6da3',
                    cmd: "",
                    scale: 0,
                    size: "small",
                    error: null,
                    conditions: expect.anything(),
                },
                {
                    type: 'workflow-service',
                    namespace: 'git-test-services',
                    filePath: '/greeting/greeting.yaml',
                    name: 'greeter',
                    image: 'direktiv/hello-world:dev',
                    id: 'git-test-services-greeter-greeting-greeting-yaml-a09fc061bb',
                    cmd: "",
                    scale: 0,
                    size: "small",
                    error: null,
                    conditions: expect.anything(),
                },
                {
                    type: 'workflow-service',
                    namespace: 'git-test-services',
                    filePath: '/input-convert/workflow.yaml',
                    name: 'csvkit',
                    image: 'direktiv/csvkit:dev',
                    id: 'git-test-services-csvkit-input-convert-workflow-ya-6c50acea98',
                    cmd: "",
                    scale: 0,
                    size: "small",
                    error: null,
                    conditions: expect.anything(),
                },
                {
                    type: 'workflow-service',
                    namespace: 'git-test-services',
                    filePath: '/request-external-api/fetch-api.yaml',
                    name: 'http-request',
                    image: 'direktiv/http-request:dev',
                    id: 'git-test-services-http-request-request-external-ap-d81abd315a',
                    cmd: "",
                    scale: 0,
                    size: "small",
                    error: null,
                    conditions: expect.anything(),
                },
                {
                    type: 'workflow-service',
                    namespace: 'git-test-services',
                    filePath: '/scripting/file.yaml',
                    name: 'python',
                    image: 'direktiv/python:dev',
                    id: 'git-test-services-python-scripting-file-yaml-eb8d7c4890',
                    cmd: "",
                    scale: 0,
                    size: "small",
                    error: null,
                    conditions: expect.anything(),
                },
                {
                    type: 'namespace-service',
                    namespace: 'git-test-services',
                    filePath: '/services/s1.yaml',
                    name: '',
                    image: 'gcr.io/direktiv/functions/http-request:1.0',
                    id: 'git-test-services-services-s1-yaml-ebcc7ad135',
                    cmd: "",
                    scale: 2,
                    size: "small",
                    error: null,
                    conditions: expect.anything(),
                },
                {
                    type: 'workflow-service',
                    namespace: 'git-test-services',
                    filePath: '/solving-math-expressions/solve-math.yaml',
                    name: 'solve-math-expression',
                    image: 'direktiv/bash:dev',
                    id: 'git-test-services-solve-math-expression-solving-ma-f242d276b6',
                    cmd: "",
                    scale: 0,
                    size: "small",
                    error: null,
                    conditions: expect.anything(),
                },
                {
                    type: 'workflow-service',
                    namespace: 'git-test-services',
                    filePath: '/variables/workflow-scope.yaml',
                    name: 'bash',
                    image: 'direktiv/bash:dev',
                    id: 'git-test-services-bash-variables-workflow-scope-ya-8c6b60ad75',
                    cmd: "",
                    scale: 0,
                    size: "small",
                    error: null,
                    conditions: expect.anything(),
                }
            ]
        })
    })
});

function sleep(ms) {
    return new Promise(resolve => setTimeout(resolve, ms));
}