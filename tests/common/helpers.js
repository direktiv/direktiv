import request from 'supertest'

import config from "./config";
import common from "./index";

async function deleteAllNamespaces() {
    var listResponse = await request(config.getDirektivHost()).get(`/api/namespaces`)
    expect(listResponse.statusCode).toEqual(200)
    var namespaces = listResponse.body.results

    if (namespaces.length == 0) {
        return
    }

    for (let i = 0; i < namespaces.length; i++) {
        var name = namespaces[i].name
        var req = await request(config.getDirektivHost()).delete(`/api/namespaces/${name}?recursive=true`)
    }

    var listResponse = await request(config.getDirektivHost()).get(`/api/namespaces`)
    expect(listResponse.statusCode).toEqual(200)
    expect(listResponse.body.results.length).toEqual(0)
}

async function expectCreateNamespace(expect, testNamespace) {
    const res = await request(common.config.getDirektivHost()).put(`/api/namespaces/${testNamespace}`)
    expect(res.statusCode).toEqual(200)
    expect(res.body).toMatchObject({
        namespace: {
            name: testNamespace,
            oid: expect.stringMatching(common.regex.uuidRegex),
            // regex /^2.*Z$/ matches timestamps like 2023-03-01T14:19:52.383871512Z
            createdAt: expect.stringMatching(/^2.*Z$/),
            updatedAt: expect.stringMatching(/^2.*Z$/),
        }
    })
}

export default {
    deleteAllNamespaces,
    expectCreateNamespace,
}