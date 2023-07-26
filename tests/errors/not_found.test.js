import request from 'supertest'
import common from "../common";

describe('Test path not found', () => {
    const methods = [
        "get",
        "post",
        "put",
        "delete",
    ];

    const paths = [
        "/api/something",
        "/api/something/",
        "/api/v1/something",
        "/api/v1/something/",
        "/api/v1/something/not/found",
        "/api/v1/something/not/found/",
    ]

    paths.forEach(path => {
        methods.forEach(method => {
            it(`should return not_found for path:${path} with method:${method}`, async () => {
                const res = await request(common.config.getDirektivHost())[method](path)

                // currently api returns 405 for not found resources, but by http standards, it should return 404.
                expect(res.statusCode).toEqual(405)
                expect(res.body).toMatchObject({})
            })
        })
    })
})
