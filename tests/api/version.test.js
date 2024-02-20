import request from "../common/request"

import common from "../common"

describe('Test the version information API', () => {
    it(`should request version information`, async () => {
        var r = await request(common.config.getDirektivHost()).get(`/api/v2/version`)
        expect(r.statusCode).toEqual(200)

        expect(r.body).toMatchObject({
            data: expect.anything()
        })
    })
})
