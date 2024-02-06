import request from 'supertest'
import common from "../common";
import regex from "../common/regex";

const testNamespace = "test-file-namespace"

describe('Test filesystem tree read operations', () => {
    beforeAll(common.helpers.deleteAllNamespaces)

    common.helpers.itShouldCreateNamespace(it, expect, testNamespace)

    it(`should fail creating file with invalid base64 content`, async () => {
        const res = await request(common.config.getDirektivHost())
            .post(`/api/v2/namespaces/${testNamespace}/files-tree`)
            .set('Content-Type', 'application/json')
            .send({
                name: "foo",
                type: "workflow",
                mimeType: "text/plain",
                content:  "some_invalid_base64",
            })
        expect(res.statusCode).toEqual(400)
        expect(res.body).toMatchObject({
            error: {
                code: 'request_data_invalid',
                message: 'file content has invalid base64 string'
            }
        })
    })

    it(`should fail creating file with invalid yaml content`, async () => {
        const res = await request(common.config.getDirektivHost())
            .post(`/api/v2/namespaces/${testNamespace}/files-tree`)
            .set('Content-Type', 'application/json')
            .send({
                name: "foo",
                type: "workflow",
                mimeType: "text/plain",
                content:  btoa("11 some_invalid_yaml 11 \nsome_invalid_yaml"),
            })
        expect(res.statusCode).toEqual(400)
        expect(res.body).toMatchObject({
            error: {
                code: 'request_data_invalid',
                message: 'file content has invalid yaml string'
            }
        })
    })
})
