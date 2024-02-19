import common from "../common";
import request from "../common/request"

const testNamespace = "test-git-namespace"

describe('Test namespace git mirroring', () => {
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
        expect(res.body).toMatchObject({
            namespace: {
                name: testNamespace,
                // regex /^2.*Z$/ matches timestamps like 2023-03-01T14:19:52.383871512Z
                createdAt: expect.stringMatching(/^2.*Z$/),
                updatedAt: expect.stringMatching(/^2.*Z$/),
            }
        })
    })

    it(`should get the new git namespace`, async () => {
        await sleep(7000)
        const res = await request(common.config.getDirektivHost()).get(`/api/namespaces/${testNamespace}/tree/aws`)
        expect(res.statusCode).toEqual(200)
        expect(res.body).toMatchObject({
            namespace: testNamespace,
        })
    })

    it(`should delete the new git namespace`, async () => {
        const res = await request(common.config.getDirektivHost()).delete(`/api/namespaces/${testNamespace}?recursive=true`)
        expect(res.statusCode).toEqual(200)
        expect(res.body).toMatchObject({})
    })

    it(`should get 404 after the new git namespace deletion`, async () => {
        const res = await request(common.config.getDirektivHost()).get(`/api/namespaces/${testNamespace}/tree`)
        expect(res.statusCode).toEqual(404)
        expect(res.body).toMatchObject({
            code: 404,
            message: "ErrNotFound",
        })
    })
})

function sleep(time) {
    return new Promise((resolve) => setTimeout(resolve, time));
}