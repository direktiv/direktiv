import { createNamespace, deleteNamespace } from "e2e/utils/namespace";
import { expect, test } from "@playwright/test";

import { createFile } from "e2e/utils/files";
import { createGatewayFile } from "../info/utils";
import { createRouteYaml } from "../../explorer/route/utils";
import yaml from "yaml";

let namespace = "";

test.beforeEach(async () => {
  namespace = await createNamespace();
});

test.afterEach(async () => {
  await deleteNamespace(namespace);
  namespace = "";
});

test("Documentation section show no documentation available when no route endpoints are set", async ({
  page,
}) => {
  await page.goto(`/n/${namespace}/gateway/openapiDoc`, {
    waitUntil: "networkidle",
  });

  await page.goto(`/n/${namespace}/gateway/openapiDoc`, {
    waitUntil: "networkidle",
  });

  await expect(
    page.getByTestId("breadcrumb-documentation"),
    "it renders the 'Documentation' breadcrumb"
  ).toBeVisible();

  await expect(
    page.getByTestId("breadcrumb-namespace"),
    "it displays the current namespace in the breadcrumb"
  ).toHaveText(namespace);

  await expect(
    page.getByText("No documentation found"),
    "Notification for no available documentation visible"
  ).toBeVisible();
});

test("Make sure Rapidoc component is displaying basic OpenAPI doc", async ({
  page,
}) => {
  const testOpenApiObject = {
    openapi: "3.0.0",
    info: {
      title: namespace,
      version: "2.0.0",
      description: "testDescription",
    },
  };

  await createGatewayFile({
    name: "testname",
    fileContent: yaml.stringify(testOpenApiObject),
    namespace,
  });

  await page.goto(`/n/${namespace}/gateway/openapiDoc`, {
    waitUntil: "networkidle",
  });

  await expect(
    page.getByText("No documentation found"),
    "Notification for no available documentation visible"
  ).toBeVisible();

  const filename = "myroute.yaml";

  type CreateRouteYamlParam = Parameters<typeof createRouteYaml>[0];
  const minimalRouteConfig: Omit<CreateRouteYamlParam, "plugins"> = {
    path: "path",
    timeout: 3000,
    methods: {
      get: {},
      post: {},
    },
    allow_anonymous: true,
  };

  const basicTargetPlugin = `
    type: instant-response
    configuration:
      status_code: 200`;

  const initialRouteYaml = createRouteYaml({
    ...minimalRouteConfig,
    plugins: {
      target: basicTargetPlugin,
    },
  });

  await createFile({
    namespace,
    name: filename,
    type: "endpoint",
    content: initialRouteYaml,
    mimeType: "application/yaml",
  });

  await page.goto(`/n/${namespace}/gateway/openapiDoc`, {
    waitUntil: "networkidle",
  });

  await expect(
    page.getByText(`${namespace} 2.0.0 ${testOpenApiObject.info.description}`)
  ).toBeVisible();

  await expect(page.getByText("get", { exact: true })).toBeVisible();
  await expect(page.getByText("post", { exact: true })).toBeVisible();

  await expect(page.locator('text="/path"')).toHaveCount(2);
});
