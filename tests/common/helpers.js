import request from 'supertest'

import config from "./config";

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

export default {
    deleteAllNamespaces,
}