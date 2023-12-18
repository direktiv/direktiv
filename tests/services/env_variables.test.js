import request from 'supertest'
import retry from "jest-retries";
import common from "../common";

const testNamespace = "test-services"

describe('Test services crud operations', () => {
    beforeAll(common.helpers.deleteAllNamespaces)

    common.helpers.itShouldCreateNamespace(it, expect, testNamespace)

    common.helpers.itShouldCreateFile(it, expect, testNamespace,
        "/s1.yaml", `
direktiv_api: service/v1
image: redis
cmd: redis-server
scale: 1
envs:
- name: foo1
  value: bar1
- name: foo2
  value: bar2
`)

    common.helpers.itShouldCreateFile(it, expect, testNamespace,
        "/s2.yaml", `
direktiv_api: service/v1
image: redis
cmd: redis-server
scale: 2
`)

    let listRes;
    retry(`should list all services`, 10, async () => {
        await sleep(500)
        listRes = await request(common.config.getDirektivHost())
            .get(`/api/v2/namespaces/${testNamespace}/services`)
        expect(listRes.statusCode).toEqual(200)
        expect(listRes.body).toMatchObject({
            data: [
                {
                    "cmd": "redis-server",
                    "error": null,
                    "filePath": "/s1.yaml",
                    "id": "test-services-s1-yaml-913535492b",
                    "image": "redis",
                    "namespace": "test-services",
                    "scale": 1,
                    "size": "medium",
                    "type": "namespace-service",
                    "envs": [
                        {name: "foo1", value: "bar1"},
                        {name: "foo2", value: "bar2"},
                    ]
                },
                {
                    "cmd": "redis-server",
                    "error": null,
                    "filePath": "/s2.yaml",
                    "id": "test-services-s2-yaml-d6f019ac00",
                    "image": "redis",
                    "namespace": "test-services",
                    "scale": 2,
                    "size": "medium",
                    "type": "namespace-service",
                    "envs": []
                }
            ]
        })
    })
});

function sleep(ms) {
    return new Promise(resolve => setTimeout(resolve, ms));
}