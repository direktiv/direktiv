import request from 'supertest'

import config from "./config";
import common from "./index";

async function deleteAllNamespaces() {
    let listResponse = await request(config.getDirektivHost()).get(`/api/namespaces`)
    if(listResponse.statusCode !== 200) {
        throw Error(`none ok namespaces list statusCode(${listResponse.statusCode})`)
    }

    for (const namespace of listResponse.body.results) {
        let response = await request(config.getDirektivHost()).delete(`/api/namespaces/${namespace.name}?recursive=true`)

        if(response.statusCode !== 200) {
            throw Error(`none ok namespace(${namespace.name}) delete statusCode(${response.statusCode})`)
        }
    }
}

async function itShouldCreateNamespace(it, expect, ns) {
    it(`should create a new namespace ${ns}`, async () => {
        const res = await request(common.config.getDirektivHost()).put(`/api/namespaces/${ns}`)
        expect(res.statusCode).toEqual(200)
        expect(res.body).toMatchObject({
            namespace: {
                name: ns,
                oid: expect.stringMatching(common.regex.uuidRegex),
                // regex /^2.*Z$/ matches timestamps like 2023-03-01T14:19:52.383871512Z
                createdAt: expect.stringMatching(/^2.*Z$/),
                updatedAt: expect.stringMatching(/^2.*Z$/),
            }
        })
    })
}

async function itShouldCreateServiceFile(it, expect, ns, path, content) {
    it(`should create a new service file ${path}`, async () => {
        const res = await request(common.config.getDirektivHost())
            .put(`/api/namespaces/${ns}/tree${path}?op=create-workflow`)
            .set({
                'Content-Type': 'text/plain',
            })

            .send(content)

        expect(res.statusCode).toEqual(200)
        expect(res.body).toMatchObject({
            namespace: ns,
        })
    })
}

async function itShouldDeleteFile(it, expect, ns, path) {
    it(`should create a new service file ${path}`, async () => {
        const res = await request(common.config.getDirektivHost())
            .delete(`/api/namespaces/${ns}/tree${path}?op=delete-nod`)

        expect(res.statusCode).toEqual(200)
        expect(res.body).toMatchObject({
            namespace: ns,
        })
    })
}

let itShouldCreateEndpointFile = itShouldCreateServiceFile

export default {
    deleteAllNamespaces,
    itShouldCreateNamespace,
    itShouldCreateServiceFile,
    itShouldCreateEndpointFile,
    itShouldDeleteFile
}