import common from "../common";
import request from "supertest";
import retry from "jest-retries";


const testNamespace = "gateway";

const limitedNamespace = "limited_namespace";

const endpointNSFile = `
direktiv_api: endpoint/v1
allow_anonymous: true
plugins:
  target:
    type: target-namespace-file
    configuration:
        namespace: ` + testNamespace + `
        file: /endpoint1.yaml
methods: 
  - GET
path: /endpoint1`

const endpointNSFileAllowed = `
  direktiv_api: endpoint/v1
  allow_anonymous: true
  plugins:
    target:
      type: target-namespace-file
      configuration:
          file: /endpoint1.yaml
  methods: 
    - GET
  path: /endpoint2`
  

  const endpointBroken= `
  direktiv_api: endpoint/v1
  allow_anonymous: true
  plugins:
    target:
      type: target-namespace-file
  methods: 
    - GET
  path: /endpoint3`

describe("Test target file wrong config", () => {
    beforeAll(common.helpers.deleteAllNamespaces);

    common.helpers.itShouldCreateNamespace(it, expect, testNamespace);

    common.helpers.itShouldCreateFile(
      it,
      expect,
      testNamespace,
      "/ep3.yaml",
      endpointBroken
    );

    retry(`should list all services`, 10, async () => {
      await sleep(500)
      const listRes = await request(common.config.getDirektivHost()).get(
        `/api/v2/namespaces/${testNamespace}/gateway/routes`
      );
      expect(listRes.statusCode).toEqual(200);
      expect(listRes.body.data.length).toEqual(1);
      expect(listRes.body.data).toEqual(
        expect.arrayContaining(
          [
            {
              file_path: '/ep3.yaml',
              path: '/endpoint3',
              methods: [ 'GET' ],
              server_path: '/gw/endpoint3',
              allow_anonymous: true,
              timeout: 0,
              errors: [ 'file is required' ],
              warnings: [],
              plugins: { target: {"type": "target-namespace-file"} }
            }
          ]
        )
      );
    })

});


function sleep(ms) {
  return new Promise(resolve => setTimeout(resolve, ms));
}



describe("Test target namespace file plugin", () => {
    beforeAll(common.helpers.deleteAllNamespaces);
  
    common.helpers.itShouldCreateNamespace(it, expect, limitedNamespace);
    common.helpers.itShouldCreateNamespace(it, expect, testNamespace);
  
    common.helpers.itShouldCreateFile(
      it,
      expect,
      testNamespace,
      "/endpoint1.yaml",
      endpointNSFile
    );
  
    common.helpers.itShouldCreateFile(
      it,
      expect,
      testNamespace,
      "/endpoint2.yaml",
      endpointNSFileAllowed
    );
  
    common.helpers.itShouldCreateFile(
      it,
      expect,
      limitedNamespace,
      "/endpoint1.yaml",
      endpointNSFile
    );
  
    common.helpers.itShouldCreateFile(
      it,
      expect,
      limitedNamespace,
      "/endpoint2.yaml",
      endpointNSFileAllowed
    );
    
  
    it(`should return a file from magic namespace`, async () => {
      const req = await request(common.config.getDirektivHost()).get(
        `/gw/endpoint1`
      );
      expect(req.statusCode).toEqual(200);
      expect(req.text).toEqual(endpointNSFile)
    });
  
    it(`should return a file from magic namespace without namespace set`, async () => {
      const req = await request(common.config.getDirektivHost()).get(
        `/gw/endpoint2`
      );
      expect(req.statusCode).toEqual(200);
      expect(req.text).toEqual(endpointNSFile)
    });
  
    it(`should return a file from non-magic namespace`, async () => {
      const req = await request(common.config.getDirektivHost()).get(
        `/ns/` + limitedNamespace + `/endpoint2`
      );
      expect(req.statusCode).toEqual(200);
      expect(req.text).toEqual(endpointNSFile)
    });
  
    it(`should not return a file`, async () => {
      const req = await request(common.config.getDirektivHost()).get(
        `/ns/` + limitedNamespace + `/endpoint1`
      );
      expect(req.statusCode).toEqual(500);
    });
    
  
  });
  