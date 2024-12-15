import { beforeAll, describe, expect, it } from "@jest/globals";

import common from "../common";
import request from "../common/request";
import { retry10 } from "../common/retry";

const testNamespace = "system";

const endpoint1 = `
get:
    summary: "Fetch endpoint details"
    description: "Retrieves example details."
    responses:
        "200":
            description: "Successful response"
x-extensions:
    direktiv: "api_path/v1"
    path: "endpoint1"
    allow-anonymous: true
    timeout: 30
    plugins:
        auth: []
        inbound: []
        target:
            type: "instant-response"
            configuration:
                status_code: 201
                status_message: "TEST1"
        outbound: []
`;

describe("Test Gateway Endpoint Functions", () => {
  beforeAll(common.helpers.deleteAllNamespaces);
  common.helpers.itShouldCreateNamespace(it, expect, testNamespace);

  describe("Validate route creation, listing and ns doc generation", () => {
    // Create and verify the addition of endpoint
    common.helpers.itShouldCreateYamlFile(
      it,
      expect,
      testNamespace,
      "/",
      "endpoint1.yaml",
      "api_path",
      endpoint1
    );

    retry10(`should list single endpoint`, async () => {
      const listRes = await request(common.config.getDirektivHost()).get(
        `/api/v2/namespaces/${testNamespace}/gateway/routes?path=/endpoint1`
      );

      expect(listRes.statusCode).toEqual(200);
      expect(listRes.body.data.length).toEqual(1);
      expect(listRes.body.data[0]).toEqual({
        allow_anonymous: true,
        errors: [],
        warnings: [],
        server_path: `/ns/${testNamespace}/endpoint1`,
        file_path: "/endpoint1.yaml",
        methods: ["GET"],
        path: "/endpoint1",
        plugins: {
          target: {
            configuration: {
              status_code: 201,
              status_message: "TEST1",
            },
            type: "instant-response",
          },
        },
        timeout: 30,
      });
    });

    it(`should generate basic namespace OpenAPI spec`, async () => {
      const res = await request(common.config.getDirektivHost()).get(
        `/api/v2/namespaces/${testNamespace}/doc`
      );
      expect(res.statusCode).toEqual(200);
      expect(res.body.data.paths).toHaveProperty(
        "/api/v2/namespaces/system/endpoint1"
      );

      const endpointData =
        res.body.data.paths["/api/v2/namespaces/system/endpoint1"];
      function validateExtensionProperties(actual, expected) {
        expect(actual["allow-anonymous"]).toEqual(expected["allow-anonymous"]);
        expect(actual.direktiv).toEqual(expected.direktiv);
        expect(actual.path).toEqual(expected.path);
        expect(actual.timeout).toEqual(expected.timeout);
        expect(actual.plugins).toEqual(expected.plugins);
      }
      const expected = {
        openapi: "3.0.3", // demo data
        info: {
          title: "Example API",
          description: "This is an example API to test OpenAPI specs",
          version: "1.0.0", // demo data
        },
        servers: [
          {
            url: "https://api.example.com/v1", // demo data
            description: "Production server",
          },
        ],
        paths: {
          "/api/v2/namespaces/system/endpoint1": {
            get: {
              summary: "Fetch endpoint details",
              description: "Retrieves example details.",
              responses: {
                200: { description: "Successful response" },
              },
              "x-extensions": {
                direktiv: "api_path/v1",
                "allow-anonymous": true,
                path: "endpoint1",
                plugins: {
                  auth: [],
                  inbound: [],
                  outbound: [],
                  target: {
                    type: "instant-response",
                    configuration: {
                      status_code: 201,
                      status_message: "TEST1",
                    },
                  },
                },
                timeout: 30,
              },
            },
          },
        },
      };
      // Compare expected values with actual ones
      expect(endpointData.get.summary).toEqual(
        expected.paths["/api/v2/namespaces/system/endpoint1"].get.summary
      );
      expect(endpointData.get.description).toEqual(
        expected.paths["/api/v2/namespaces/system/endpoint1"].get.description
      );
      expect(endpointData.get.responses).toEqual(
        expected.paths["/api/v2/namespaces/system/endpoint1"].get.responses
      );
      validateExtensionProperties(
        endpointData["x-extensions"],
        expected.paths["/api/v2/namespaces/system/endpoint1"].get[
          "x-extensions"
        ]
      );
      expect(res.body.data.openapi).toEqual(expected.openapi);
      expect(res.body.data.info).toEqual(expected.info);
      expect(res.body.data.servers).toEqual(expected.servers);
    });
  });
});
