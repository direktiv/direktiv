import common from "../common";
import request from "supertest";
import retry from "jest-retries";

const testNamespace = "gateway_namespace";

const endpoint1 = `
direktiv_api: endpoint/v1
plugins:
  auth:
  - type: key-auth
    configuration:
        key_name: secret
  target:
    type: instant-response
    configuration:
        status_code: 201
        status_message: "TEST1"
methods: 
  - GET
path: /endpoint1`


const endpoint2 = `
direktiv_api: endpoint/v1
allow_anonymous: true
plugins:
  auth:
  - type: basic-auth
  - type: key-auth
    configuration:
        key_name: secret
  target:
    type: instant-response
    configuration:
        status_code: 202
        status_message: "TEST2"
methods: 
  - GET
path: /endpoint2`

const consumer1 = `
direktiv_api: "consumer/v1"
username: consumer1
password: pwd
api_key: key1
tags:
- tag1
groups:
- group1`

const consumer2 = `
direktiv_api: "consumer/v1"
username: consumer2
password: pwd
api_key: key2
tags:
- tag2
groups:
- group2`


const endpointBroken = `direktiv_api: endpoint/v1
allow_anonymous: true
plugins:
  outbound:
    type: js-outbound
methods: 
  - GET
path: ep4`

function sleep(ms) {
  return new Promise(resolve => setTimeout(resolve, ms));
}

describe("Test wrong endpoint config", () => {
    beforeAll(common.helpers.deleteAllNamespaces);

    common.helpers.itShouldCreateNamespace(it, expect, testNamespace);

    common.helpers.itShouldCreateFile(
        it,
        expect,
        testNamespace,
        "/endpointbroken.yaml",
        endpointBroken
    );

    retry(`should list all endpoints`, 10, async () => {
        const listRes = await request(common.config.getDirektivHost()).get(
            `/api/v2/namespaces/${testNamespace}/gateway/routes`
        );
        expect(listRes.statusCode).toEqual(200);
        expect(listRes.body.data.length).toEqual(1);
        expect(listRes.body.data).toEqual(
            expect.arrayContaining(
                [
                    {
                        file_path: '/endpointbroken.yaml',
                        server_path: '',
                        methods: [],
                        allow_anonymous: false,
                        timeout: 0,
                        errors: [
                            'yaml: unmarshal errors:\n' +
                            '  line 5: cannot unmarshal !!map into []core.PluginConfig'
                        ],
                        warnings: [],
                        plugins: {
                            target: {
                                type: ""
                            }
                        }
                    }
                ]
            )
        );
    });

});


describe("Test gateway endpoints on create", () => {
  beforeAll(common.helpers.deleteAllNamespaces);

    common.helpers.itShouldCreateNamespace(it, expect, testNamespace);

    retry(`should list all endpoints`, 10, async () => {
      const listRes = await request(common.config.getDirektivHost()).get(
        `/api/v2/namespaces/${testNamespace}/gateway/routes`
      );
      expect(listRes.statusCode).toEqual(200);
      expect(listRes.body.data.length).toEqual(0);
      expect(listRes.body.data).toEqual(
        expect.arrayContaining(
          [
          ]
        )
      );
    });


});

describe("Test gateway endpoints crud operations", () => {
  beforeAll(common.helpers.deleteAllNamespaces);

    common.helpers.itShouldCreateNamespace(it, expect, testNamespace);

    common.helpers.itShouldCreateFile(
      it,
      expect,
      testNamespace,
      "/endpoint1.yaml",
      endpoint1
    );

    common.helpers.itShouldCreateFile(
      it,
      expect,
      testNamespace,
      "/endpoint2.yaml",
      endpoint2
    );

    common.helpers.itShouldCreateFile(
      it,
      expect,
      testNamespace,
      "/consumer1.yaml",
      consumer1
    );

    common.helpers.itShouldCreateFile(
      it,
      expect,
      testNamespace,
      "/consumer2.yaml",
      consumer2
    );

  retry(`should list all endpoints`, 10, async () => {
    const listRes = await request(common.config.getDirektivHost()).get(
      `/api/v2/namespaces/${testNamespace}/gateway/routes`
    );
    expect(listRes.statusCode).toEqual(200);
    expect(listRes.body.data.length).toEqual(2);
    expect(listRes.body.data).toEqual(
      expect.arrayContaining(
        [
          { "allow_anonymous": false, 
            "errors": [], 
            "file_path": "/endpoint1.yaml", 
            "methods": ["GET"], 
            "path": "/endpoint1", 
            "server_path": "/gw/endpoint1",
            "plugins": {
              "auth": [{"configuration": {"key_name": "secret"}, "type": "key-auth"}], 
              "target": {"configuration": {"status_code": 201, "status_message": "TEST1"}, 
              "type": "instant-response"}
            }, 
            "timeout": 0, 
            "warnings": []
          }, {
            "allow_anonymous": true, 
            "errors": [], 
            "file_path": "/endpoint2.yaml", 
            "methods": ["GET"], 
            "path": "/endpoint2", 
            "server_path": "/gw/endpoint2",
            "plugins": {
              "auth": [
                {"type": "basic-auth"}, 
                {"configuration": {"key_name": "secret"}, "type": "key-auth"}
              ], 
              "target": {"configuration": {"status_code": 202, "status_message": "TEST2"}, "type": "instant-response"}}, 
            "timeout": 0, 
            "warnings": []
          }
        ]
      )
    );
  });

  
  retry(`should list all consumers`, 10, async () => {
    const listRes = await request(common.config.getDirektivHost()).get(
      `/api/v2/namespaces/${testNamespace}/gateway/consumers`
    );
    expect(listRes.statusCode).toEqual(200);
    expect(listRes.body.data.length).toEqual(2);
    expect(listRes.body.data).toEqual(
      expect.arrayContaining(
        [{
         "api_key": "key2", 
         "groups": ["group2"],
         "password": "pwd",
         "tags": ["tag2"],
         "username": "consumer2"}, 
         {
         "api_key": "key1",
         "groups": ["group1"],
         "password": "pwd",
         "tags": ["tag1"],
         "username": "consumer1"
        }]
      )
    );
  });

  common.helpers.itShouldDeleteFile(it, expect, testNamespace, "/endpoint1.yaml");
  common.helpers.itShouldDeleteFile(it, expect, testNamespace, "/consumer1.yaml");

  it(`should list one route after delete`, async () => {
    const listRes = await request(common.config.getDirektivHost()).get(
      `/api/v2/namespaces/${testNamespace}/gateway/routes`
    );
    expect(listRes.statusCode).toEqual(200);
    expect(listRes.body.data.length).toEqual(1);
  });

  it(`should list one consumer after delete`, async () => {
    const listRes = await request(common.config.getDirektivHost()).get(
      `/api/v2/namespaces/${testNamespace}/gateway/consumers`
    );
    expect(listRes.statusCode).toEqual(200);
    expect(listRes.body.data.length).toEqual(1);
  });

});

describe("Test availability of gateway endpoints", () => {
  beforeAll(common.helpers.deleteAllNamespaces);

  common.helpers.itShouldCreateNamespace(it, expect, testNamespace);

  common.helpers.itShouldCreateFile(
    it,
    expect,
    testNamespace,
    "/endpoint1.yaml",
    endpoint1
  );

  common.helpers.itShouldCreateFile(
    it,
    expect,
    testNamespace,
    "/endpoint2.yaml",
    endpoint2
  );

  common.helpers.itShouldCreateFile(
    it,
    expect,
    testNamespace,
    "/consumer1.yaml",
    consumer1
  );

  common.helpers.itShouldCreateFile(
    it,
    expect,
    testNamespace,
    "/consumer2.yaml",
    consumer2
  );

  it(`should not run endpoint without authentication`, async () => {
    const req = await request(common.config.getDirektivHost()).get(
      `/gw/endpoint1`
    );
    expect(req.statusCode).toEqual(401);
  });

  it(`should run endpoint without authentication but allow anonymous`, async () => {
    const req = await request(common.config.getDirektivHost()).get(
      `/gw/endpoint2`
    );
    expect(req.statusCode).toEqual(202);
  });

  it(`should run endpoint with key authentication`, async () => {
    const req = await request(common.config.getDirektivHost()).get(
      `/gw/endpoint1`
    ).set('secret', 'key2');
    expect(req.statusCode).toEqual(201);
  });

  it(`should run endpoint with basic authentication`, async () => {
    const req = await request(common.config.getDirektivHost()).get(
      `/gw/endpoint2`
    ).auth('consumer1', 'pwd');
    expect(req.statusCode).toEqual(202);
  });

});