import request from 'supertest'
import common from "../common";
import regex from "../common/regex";

const testNamespace = "test-file-namespace"

describe('Test filesystem tree read operations', () => {
    beforeAll(common.helpers.deleteAllNamespaces)

    common.helpers.itShouldCreateNamespace(it, expect, testNamespace)

    it(`should read empty root dir`, async () => {
        const res = await request(common.config.getDirektivHost())
            .get(`/api/v2/namespaces/${testNamespace}/files-tree`)
        expect(res.statusCode).toEqual(200)
        expect(res.body).toMatchObject({
            data: {
                file: {
                    path: "/",
                    type: "directory",
                    createdAt: expect.stringMatching(regex.timestampRegex),
                    updatedAt: expect.stringMatching(regex.timestampRegex),
                },
                paths: []
            }
        })
    })

    common.helpers.itShouldCreateDirectory(it, expect, testNamespace, "/dir1")
    common.helpers.itShouldCreateDirectory(it, expect, testNamespace, "/dir2")
    common.helpers.itShouldCreateFile(it, expect, testNamespace, "/foo.yaml", common.helpers.dummyWorkflow("foo"))
    common.helpers.itShouldCreateFile(it, expect, testNamespace, "/dir1/foo11.yaml", common.helpers.dummyWorkflow("foo11"))
    common.helpers.itShouldCreateFile(it, expect, testNamespace, "/dir1/foo12.yaml", common.helpers.dummyWorkflow("foo12"))

    it(`should read root dir two files`, async () => {
        const res = await request(common.config.getDirektivHost())
            .get(`/api/v2/namespaces/${testNamespace}/files-tree`)
        expect(res.statusCode).toEqual(200)
        expect(res.body).toMatchObject({
            data: {
                file: {
                    path: "/",
                    type: "directory",
                    createdAt: expect.stringMatching(regex.timestampRegex),
                    updatedAt: expect.stringMatching(regex.timestampRegex),
                },
                paths: [
                    {
                        path: "/dir1",
                        type: "directory",
                        createdAt: expect.stringMatching(regex.timestampRegex),
                        updatedAt: expect.stringMatching(regex.timestampRegex),

                    },
                    {
                        path: "/dir2",
                        type: "directory",
                        createdAt: expect.stringMatching(regex.timestampRegex),
                        updatedAt: expect.stringMatching(regex.timestampRegex),

                    },
                    {
                        path: "/foo.yaml",
                        type: "workflow",
                        mimeType: "application/direktiv",
                        createdAt: expect.stringMatching(regex.timestampRegex),
                        updatedAt: expect.stringMatching(regex.timestampRegex),

                    },
                ]
            }
        })
    })

    it(`should read dir1 with two files`, async () => {
        const res = await request(common.config.getDirektivHost())
            .get(`/api/v2/namespaces/${testNamespace}/files-tree/dir1`)
        expect(res.statusCode).toEqual(200)
        expect(res.body).toMatchObject({
            data: {
                file: {
                    path: "/dir1",
                    type: "directory",
                    createdAt: expect.stringMatching(regex.timestampRegex),
                    updatedAt: expect.stringMatching(regex.timestampRegex),
                },
                paths: [
                    {
                        mimeType: "application/direktiv",
                        path: "/dir1/foo11.yaml",
                        type: "workflow",
                        createdAt: expect.stringMatching(regex.timestampRegex),
                        updatedAt: expect.stringMatching(regex.timestampRegex),

                    },
                    {
                        mimeType: "application/direktiv",
                        path: "/dir1/foo12.yaml",
                        type: "workflow",
                        createdAt: expect.stringMatching(regex.timestampRegex),
                        updatedAt: expect.stringMatching(regex.timestampRegex),

                    }
                ]
            }
        })
    })

    it(`should read dir2 with zero files`, async () => {
        const res = await request(common.config.getDirektivHost())
            .get(`/api/v2/namespaces/${testNamespace}/files-tree/dir2`)
        expect(res.statusCode).toEqual(200)
        expect(res.body).toMatchObject({
            data: {
                file: {
                    path: "/dir2",
                    type: "directory",
                    createdAt: expect.stringMatching(regex.timestampRegex),
                    updatedAt: expect.stringMatching(regex.timestampRegex),
                },
                paths: []
            }
        })
    })
})
