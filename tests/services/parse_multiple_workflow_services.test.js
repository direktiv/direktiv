import request from 'supertest'
import retry from "jest-retries";
import common from "../common";

const testNamespace = "test-services"

describe('Test services crud operations', () => {
    beforeAll(common.helpers.deleteAllNamespaces)

    common.helpers.itShouldCreateNamespace(it, expect, testNamespace)

    common.helpers.itShouldCreateFile(it, expect, testNamespace,
        "/s2.yaml", `
direktiv_api: workflow/v1

functions:
- id: image-magick
  image: gcr.io/direktiv/functions/image-magick:1.0
  type: knative-workflow
- id: call-btc
  workflow: get-bc.yaml
  type: subflow

states:
- id: subflow
  type: action
  action:
    function: call-btc
- id: draw
  type: action
  action:
    function: image-magick
    files:
    - key: happy.png
      scope: file
    input: 
      commands:
      - ls -la
`)

    retry(`should list all services`, 10, async () => {
        await sleep(500)
        const listRes = await request(common.config.getDirektivHost())
            .get(`/api/v2/namespaces/${testNamespace}/services`)
        expect(listRes.statusCode).toEqual(200)
        expect(listRes.body).toMatchObject({
            data: [
                {
                    type: 'workflow-service',
                    namespace: 'test-services',
                    filePath: '/s2.yaml',
                    name: 'image-magick',
                    image: 'gcr.io/direktiv/functions/image-magick:1.0',
                    cmd: '',
                    size: 'small',
                    scale: 0,
                    error: null,
                    id: 'test-services-image-magick-s2-yaml-864bf960ad',
                }
            ]
        })
    })
});

function sleep(ms) {
    return new Promise(resolve => setTimeout(resolve, ms));
}