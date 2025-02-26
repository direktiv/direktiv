import { createNamespace, deleteNamespace } from "e2e/utils/namespace";
import { expect, test } from "@playwright/test";

import { createGatewayFile } from "./utils";
import yaml from "yaml";

let namespace = "";

test.beforeEach(async () => {
  namespace = await createNamespace();
});

test.afterEach(async () => {
  await deleteNamespace(namespace);
  namespace = "";
});

test("Info page default view", async ({ page }) => {
  await page.goto(`/n/${namespace}/gateway/gatewayInfo`, {
    waitUntil: "networkidle",
  });

  await expect(
    page.getByTestId("breadcrumb-info"),
    "it renders the 'Info' breadcrumb"
  ).toBeVisible();

  await expect(page.getByText("Gateway Info")).toBeVisible();

  await expect(
    page.getByTestId("breadcrumb-namespace"),
    "it displays the current namespace in the breadcrumb"
  ).toHaveText(namespace);

  await expect(
    page.getByRole("cell", { name: namespace }),
    "it displays the current namespace in the info section"
  ).toBeVisible();

  await expect(
    page.getByRole("cell", { name: "1.0" }),
    "it displays the gateway version in the info section"
  ).toBeVisible();

  const editor = page.locator(".lines-content");

  await expect(
    editor,
    "it displays the namespace in the editor preview"
  ).toContainText(`title: ${namespace}`, { useInnerText: true });

  await expect(
    editor,
    "it displays the version in the editor preview"
  ).toContainText(`version: "1.0"`, { useInnerText: true });

  await expect(
    page.getByText("Virtual"),
    "it displays the file path in the editor preview"
  ).toBeVisible();
});

test("Info page when openAPI specification exists", async ({ page }) => {
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

  await page.goto(`/n/${namespace}/gateway/gatewayInfo`, {
    waitUntil: "networkidle",
  });

  await expect(
    page.getByRole("cell", { name: "2.0.0" }),
    "it displays the gateway version in the info section"
  ).toBeVisible();

  await expect(
    page.getByRole("cell", { name: "testDescription" }),
    "it displays the gateway description in the info section"
  ).toBeVisible();

  await expect(
    page.getByRole("cell", { name: "/testname" }),
    "it displays the file path in the editor preview"
  ).toBeVisible();
});
