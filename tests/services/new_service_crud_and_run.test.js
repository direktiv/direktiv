import request from 'supertest'
import retry from "jest-retries";
import common from "../common";


const testNamespace = "test-services"

describe('Test services crud operations', () => {
    beforeAll(common.helpers.deleteAllNamespaces)

    common.helpers.itShouldCreateNamespace(it, expect, testNamespace)

    common.helpers.itShouldCreateServiceFile(it, expect, testNamespace,
        "/s1.yaml", `
direktiv_api: service/v1
name: s1
image: redis
cmd: redis-server
scale: 1
`)

    common.helpers.itShouldCreateServiceFile(it, expect, testNamespace,
        "/s2.yaml", `
direktiv_api: service/v1
name: s2
image: redis
cmd: redis-server
scale: 2
`)

    let listRes;
    it(`should list all services`, async () => {
        listRes = await request(common.config.getDirektivHost())
            .get(`/api/v2/namespaces/${testNamespace}/services`)
        expect(listRes.statusCode).toEqual(200)
        expect(listRes.body).toMatchObject({
            data: [
                {
                    "cmd": "redis-server",
                    "conditions": null,
                    "error": null,
                    "filePath": "/s1.yaml",
                    "id": "objf8a48067049cad1cdc29obj",
                    "image": "redis",
                    "name": "s1",
                    "namespace": "test-services",
                    "scale": 1,
                    "size": "medium",
                    "type": "namespace-service",

                },
                {
                    "cmd": "redis-server",
                    "conditions": null,
                    "error": null,
                    "filePath": "/s2.yaml",
                    "id": "obj9eca47019fa69482e1a8obj",
                    "image": "redis",
                    "name": "s2",
                    "namespace": "test-services",
                    "scale": 2,
                    "size": "medium",
                    "type": "namespace-service",
                }
            ]
        })
    })

    retry(`should list all service pods`, 10,async () => {
        await sleep(500)

        let sID = listRes.body.data[0].id
        let res = await request(common.config.getDirektivHost())
            .get(`/api/v2/namespaces/${testNamespace}/services/${sID}/pods`)
        expect(res.statusCode).toEqual(200)
        expect(res.body).toMatchObject({
            data: [
                {id: expect.stringMatching(`^${sID}(_|-)`)},
            ]
        })

        sID = listRes.body.data[1].id
        res = await request(common.config.getDirektivHost())
            .get(`/api/v2/namespaces/${testNamespace}/services/${sID}/pods`)
        expect(res.statusCode).toEqual(200)
        expect(res.body).toMatchObject({
            data: [
                {id: expect.stringMatching(`^${sID}(_|-)`)},
                {id: expect.stringMatching(`^${sID}(_|-)`)},
            ]
        })
    })

    retry(`should list all services`, 100, async () => {
        await sleep(500)

        const res = await request(common.config.getDirektivHost())
            .get(`/api/v2/namespaces/${testNamespace}/services`)
        expect(res.statusCode).toEqual(200)
        expect(res.body).toMatchObject({
            data: [
                {
                    "cmd": "redis-server",
                    "conditions": [{
                        message: expect.stringContaining("Up "),
                        status: "True",
                        type: "UpAndReady",
                    }],
                    "error": null,
                    "filePath": "/s1.yaml",
                    "id": "objf8a48067049cad1cdc29obj",
                    "image": "redis",
                    "name": "s1",
                    "namespace": "test-services",
                    "scale": 1,
                    "size": "medium",
                    "type": "namespace-service",
                },
                {
                    "cmd": "redis-server",
                    "conditions": [{
                        message: expect.stringContaining("Up "),
                        status: "True",
                        type: "UpAndReady",
                    }],
                    "error": null,
                    "filePath": "/s2.yaml",
                    "id": "obj9eca47019fa69482e1a8obj",
                    "image": "redis",
                    "name": "s2",
                    "namespace": "test-services",
                    "scale": 2,
                    "size": "medium",
                    "type": "namespace-service",
                }
            ]
        })
    })
});

function sleep(ms) {
    return new Promise(resolve => setTimeout(resolve, ms));
}