import { createNamespace, deleteNamespace } from "e2e/utils/namespace";
import { expect, test } from "@playwright/test";

let namespace = "";

test.beforeEach(async () => {
  namespace = await createNamespace();
});

test.afterEach(async () => {
  await deleteNamespace(namespace);
  namespace = "";
});

test("it is possible to create a route file", async ({ page }) => {
  /* prepare data */
  const filename = "myroute.yaml";

  const expectedYaml = `direktiv_api: "endpoint/v1"
  path: "path"
  timeout: 3000
  methods:
    - "GET"
    - "POST"
  plugins:
    inbound: []
    outbound: []
    auth: []
    target:
      type: "instant-response"
      configuration:
        status_code: 200`;

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

  const editor = page.locator(".lines-content");
  await expect(
    editor,
    "all entered data is represented in the editor preview"
  ).toContainText(expectedYaml, { useInnerText: true });
});
