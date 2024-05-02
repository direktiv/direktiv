import { describe, expect, it } from '@jest/globals'

import config from '../../common/config'
import request from '../../common/request'

describe('Test the jx API with a valid body.', () => {
	it(`should perform a simple jx query`, async () => {
		const r = await request(config.getDirektivHost()).post(`/api/v2/jx`)
        .send(`{
            "jx": "e30=",
            "data": "e30="
        }`)

		expect(r.statusCode).toEqual(200)
		expect(r.body.data).toEqual({
			jx: 'e30=',
            data: 'e30=',
            logs: '',
            output: ['e30='],
		})
	})
})

describe('Test the jx API with jx string query.', () => {
	it(`should perform a simple jx query`, async () => {
        // query: "jq(5)"
		const r = await request(config.getDirektivHost()).post(`/api/v2/jx`)
        .send(`{
            "jx": "ImpxKDUpIg==",
            "data": "e30="
        }`)

		expect(r.statusCode).toEqual(200)
		expect(r.body.data).toEqual({
			jx: 'ImpxKDUpIg==',
            data: 'e30=',
            logs: '',
            output: ['NQ=='],
		})
	})
})

describe('Test the jx API with broken jx string query.', () => {
	it(`should perform a simple jx query`, async () => {
		const r = await request(config.getDirektivHost()).post(`/api/v2/jx`)
        // query: "jq("
        .send(`{
            "jx": "ImpxKCI=",
            "data": "e30="
        }`)

		expect(r.statusCode).toEqual(200)
		expect(r.body.data).toEqual({
			jx: 'ImpxKCI=',
            data: 'e30=',
            logs: 'ZmFpbHVyZToganEvanMgc2NyaXB0IG1pc3NpbmcgYnJhY2tldApxdWVyeSBwcm9kdWNlZCB6ZXJvIHJlc3VsdHMK',
            output: null,
		})
	})
})

describe('Test the jx API with broken javascript query.', () => {
    it(`should perform a simple jx query`, async () => {
        // query: "js(5)"
        const r = await request(config.getDirektivHost()).post(`/api/v2/jx`)
        .send(`{
            "jx": "ImpzKDUpIg==",
            "data": "e30="
        }`)

        expect(r.statusCode).toEqual(200)
        expect(r.body.data).toEqual({
            jx: 'ImpzKDUpIg==',
            data: 'e30=',
            logs: 'ZmFpbHVyZTogZXJyb3IgaW4ganMgcXVlcnkgNTogbm8gcmVzdWx0cwpxdWVyeSBwcm9kdWNlZCB6ZXJvIHJlc3VsdHMK',
            output: null,
        })
    })
})

describe('Test the jx API with javascript query.', () => {
    it(`should perform a simple jx query`, async () => {
        // query: "js(return 5)"
        const r = await request(config.getDirektivHost()).post(`/api/v2/jx`)
        .send(`{
            "jx": "ImpzKHJldHVybiA1KSI=",
            "data": "e30="
        }`)

        expect(r.statusCode).toEqual(200)
        expect(r.body.data).toEqual({
            jx: 'ImpzKHJldHVybiA1KSI=',
            data: 'e30=',
            logs: '',
            output: ['NQ=='],
        })
    })
})

describe('Test the jx API with a structured query.', () => {
    it(`should perform a simple jx query`, async () => {
        // query: {"x": jq(5)}
        const r = await request(config.getDirektivHost()).post(`/api/v2/jx`)
        .send(`{
            "jx": "eyJ4IjogImpxKDUpIn0=",
            "data": "e30="
        }`)

        expect(r.statusCode).toEqual(200)
        expect(r.body.data).toEqual({
            jx: 'eyJ4IjogImpxKDUpIn0=',
            data: 'e30=',
            logs: '',
            output: ['ewogICJ4IjogNQp9'],
        })
    })
})


describe('Test the jx API with passing assertions.', () => {
    it(`should perform a simple jx query`, async () => {
        // query: {"x": jq(5)}
        const r = await request(config.getDirektivHost()).post(`/api/v2/jx?assert=object&assert=success`)
        .send(`{
            "jx": "eyJ4IjogImpxKDUpIn0=",
            "data": "e30="
        }`)

        expect(r.statusCode).toEqual(200)
        expect(r.body.data).toEqual({
            jx: 'eyJ4IjogImpxKDUpIn0=',
            data: 'e30=',
            logs: '',
            output: ['ewogICJ4IjogNQp9'],
        })
    })
})

describe('Test the jx API with failing assertions.', () => {
    it(`should perform a simple jx query`, async () => {
        // query: "jq(5)"
		const r = await request(config.getDirektivHost()).post(`/api/v2/jx?assert=object`)
        .send(`{
            "jx": "ImpxKDUpIg==",
            "data": "e30="
        }`)

		expect(r.statusCode).toEqual(400)
        expect(r.body.error).toEqual({
			code: "assert_object",
            message: "result is not an object"
		})
		expect(r.body.data).toEqual({
			jx: 'ImpxKDUpIg==',
            data: 'e30=',
            logs: '',
            output: ['NQ=='],
		})
    })
})
