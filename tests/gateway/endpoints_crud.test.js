import request from 'supertest'
import retry from "jest-retries";
import common from "../common";


const testNamespace = "test-services"

describe('Test gateway endpoints crud operations', () => {
    beforeAll(common.helpers.deleteAllNamespaces)

    common.helpers.itShouldCreateNamespace(it, expect, testNamespace)

    common.helpers.itShouldCreateEndpointFile(it, expect, testNamespace,
        "/g1.yaml", `
direktiv_api: endpoint/v1
method: POST
`)

    common.helpers.itShouldCreateEndpointFile(it, expect, testNamespace,
        "/g2.yaml", `
direktiv_api: endpoint/v1
method: GET
`)

    it(`should list all endpoints`, async () => {
        const listRes = await request(common.config.getDirektivHost())
            .get(`/api/v2/namespaces/${testNamespace}/gateway_endpoints`)
        expect(listRes.statusCode).toEqual(200)
        expect(listRes.body).toMatchObject({
            data: [
                {
                    method: "POST",
                },
                {
                    method: "GET",
                }
            ]
        })
    })

    common.helpers.itShouldDeleteFile(it, expect, testNamespace, "/g1.yaml")

    it(`should list all endpoints`, async () => {
        const listRes = await request(common.config.getDirektivHost())
            .get(`/api/v2/namespaces/${testNamespace}/gateway_endpoints`)
        expect(listRes.statusCode).toEqual(200)
        expect(listRes.body).toMatchObject({
            data: [
                {
                    method: "GET",
                }
            ]
        })
    })
});
