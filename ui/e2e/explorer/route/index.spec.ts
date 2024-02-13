import { createNamespace, deleteNamespace } from "e2e/utils/namespace";
import { expect, test } from "@playwright/test";

import { createRouteYaml } from "./utils";
import { createWorkflow } from "~/api/tree/mutate/createWorkflow";
import { headers } from "e2e/utils/testutils";

let namespace = "";

test.beforeEach(async () => {
  namespace = await createNamespace();
});

test.afterEach(async () => {
  await deleteNamespace(namespace);
  namespace = "";
});

test("it is possible to create a basic route file", async ({ page }) => {
  /* prepare data */
  const filename = "myroute.yaml";

  const expectedYaml = createRouteYaml({
    path: "path",
    timeout: 3000,
    methods: ["GET", "POST"],
    plugins: {
      target: `
    type: "instant-response"
    configuration:
        status_code: 200`,
    },
  });

  /* visit page */
  await page.goto(`/${namespace}/explorer/tree`, { waitUntil: "networkidle" });
  await expect(
    page.getByTestId("breadcrumb-namespace"),
    "it navigates to the test namespace in the explorer"
  ).toHaveText(namespace);

  /* create consumer */
  await page.getByRole("button", { name: "New" }).first().click();
  await page.getByRole("menuitem", { name: "Gateway" }).click();
  await page.getByRole("button", { name: "New Route" }).click();

  await expect(page.getByRole("button", { name: "Create" })).toBeDisabled();
  await page.getByPlaceholder("route-name.yaml").fill(filename);
  await page.getByRole("button", { name: "Create" }).click();

  /* close the toast, that covers the save button (makes this test 4 seconds faster) */
  await page.getByTestId("toast-close").click();

  await expect(
    page,
    "it creates the service and opens the file in the explorer"
  ).toHaveURL(`/${namespace}/explorer/endpoint/${filename}`);

  /* fill in form */
  await page.getByLabel("path").fill("path");
  await page.getByLabel("timeout").fill("3000");
  await page.getByLabel("GET").click();
  await page.getByLabel("POST").click();

  /* try to save incomplete form */
  await page.getByRole("button", { name: "Save" }).click();

  await expect(
    page.getByText("plugins : this field is invalid"),
    "it can not save the route without a valid plugin"
  ).toBeVisible();

  await page.getByRole("button", { name: "set target plugin" }).click();

  /* add an empty instant response plugin */
  await page.getByRole("combobox").click();
  await page.getByLabel("Instant Response").click();
  await page.getByRole("button", { name: "Save" }).click();

  /* check editor content */
  const editor = page.locator(".lines-content");
  await expect(
    editor,
    "all entered data is represented in the editor preview"
  ).toContainText(expectedYaml, { useInnerText: true });

  await expect(
    page.getByText("unsaved changes"),
    "it renders a hint that there are unsaved changes"
  ).toBeVisible();

  await page.getByRole("button", { name: "Save" }).click();

  await expect(
    page.getByText("unsaved changes"),
    "it does not render a hint that there are unsaved changes"
  ).not.toBeVisible();

  /* reload */
  await page.reload({ waitUntil: "networkidle" });

  await expect(
    editor,
    "the editor shows the same content after reloading the page"
  ).toContainText(expectedYaml, { useInnerText: true });
});

test("it is possible to add plugins to a route file", async ({ page }) => {
  /* prepare data */
  const filename = "myroute.yaml";

  type CreateRouteYamlParam = Parameters<typeof createRouteYaml>[0];
  const minimalRouteConfig: Omit<CreateRouteYamlParam, "plugins"> = {
    path: "path",
    timeout: 3000,
    methods: ["GET", "POST"],
  };

  const basicTargetPlugin = `
    type: "instant-response"
    configuration:
      status_code: 200`;

  const initialRouteYaml = createRouteYaml({
    ...minimalRouteConfig,
    plugins: {
      target: basicTargetPlugin,
    },
  });

  await createWorkflow({
    payload: initialRouteYaml,
    urlParams: {
      baseUrl: process.env.VITE_DEV_API_DOMAIN,
      namespace,
      name: filename,
    },
    headers,
  });

  await page.goto(`/${namespace}/explorer/endpoint/${filename}`, {
    waitUntil: "networkidle",
  });

  /* configure inbound plugin: ACL */
  await page.getByRole("button", { name: "add inbound plugin" }).click();
  await page.getByRole("combobox").click();
  await page.getByLabel("Access control list (acl)").click();
  await page
    .locator("fieldset")
    .filter({ hasText: "Allow Groups (optional)" })
    .getByPlaceholder("Enter a group")
    .fill("allow this group 1");

  await page
    .locator("fieldset")
    .filter({ hasText: "Allow Groups (optional)" })
    .getByPlaceholder("Enter a group")
    .press("Enter");

  await page
    .locator("fieldset")
    .filter({ hasText: "Allow Groups (optional)" })
    .getByPlaceholder("Enter a group")
    .nth(1)
    .fill("allow this group 2");

  await page
    .locator("fieldset")
    .filter({ hasText: "Allow Groups (optional)" })
    .getByPlaceholder("Enter a group")
    .nth(1)
    .press("Enter");

  await page.getByRole("button", { name: "Save" }).click();

  /* configure inbound plugin: Request Convert  */
  await page.getByRole("button", { name: "add inbound plugin" }).click();
  await page.getByRole("combobox").click();
  await page.getByLabel("Request Convert").click();
  await page.getByText("Omit Queries").click();
  await page.getByText("Omit Consumer").click();
  await page.getByRole("button", { name: "Save" }).click();

  /* configure outbound plugin: JavaScript */
  await page.getByRole("button", { name: "add outbound plugin" }).click();
  await page.getByRole("combobox").click();
  await page.getByLabel("JavaScript").click();
  await page.getByRole("textbox").fill("// execute some JavaScript here");
  await page.getByRole("button", { name: "Save" }).click();

  /* configure auth plugin: Request Convert */
  await page.getByRole("button", { name: "add auth plugin" }).click();
  await page.getByRole("combobox").click();
  await page.getByLabel("Github Webhook").click();
  await page.getByLabel("secret").fill("my github secret");
  await page.getByRole("button", { name: "Save" }).click();

  // const editor = page.locator(".lines-content");
  // await expect(
  //   editor,
  //   "all entered data is represented in the editor preview"
  // ).toContainText(initialRouteYaml, { useInnerText: true });
});
