import common from "../common";
import request from "supertest";

const testNamespace = "gateway_namespace";

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
  - GET`

const endpointNSFileAllowed = `
  direktiv_api: endpoint/v1
  allow_anonymous: true
  plugins:
    target:
      type: target-namespace-file
      configuration:
          file: /endpoint1.yaml
  methods: 
    - GET`
  

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
      expect(req.statusCode).toEqual(403);
    });
    
  
  });
  