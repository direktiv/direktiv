import { createNamespace, deleteNamespace } from "e2e/utils/namespace";
import { expect, test } from "@playwright/test";

import { faker } from "@faker-js/faker";

let namespace = "";

test.beforeEach(async () => {
  namespace = await createNamespace();
});

test.afterEach(async () => {
  await deleteNamespace(namespace);
  namespace = "";
});

test("it is possible to create a service", async ({ page }) => {
  /* prepare data */
  const filename = "mynewservice.yaml";

  const envVariables = Array.from({ length: 5 }, () => ({
    name: faker.lorem.word(),
    value: faker.git.shortSha(),
  }));

  const envsYaml = envVariables
    .map((item) => `\n  - name: "${item.name}"\n    value: "${item.value}"`)
    .join("");

  const expectedYaml = `direktiv_api: "service/v1"
image: "bash"
scale: 2
size: "medium"
cmd: "hello"
envs:${envsYaml}`;

  /* visit page */
  await page.goto(`/${namespace}/explorer/tree`, { waitUntil: "networkidle" });
  await expect(
    page.getByTestId("breadcrumb-namespace"),
    "it navigates to the test namespace in the explorer"
  ).toHaveText(namespace);

  /* create service */
  await page.getByRole("button", { name: "New" }).first().click();
  await page.getByRole("button", { name: "New Service" }).click();

  await expect(page.getByRole("button", { name: "Create" })).toBeDisabled();
  await page.getByPlaceholder("service-name.yaml").fill(filename);
  await page.getByRole("button", { name: "Create" }).click();

  await expect(
    page,
    "it creates the service and opens the file in the explorer"
  ).toHaveURL(`/${namespace}/explorer/service/${filename}`);

  /* fill in form */
  await page.getByLabel("Image").fill("bash");
  await page.locator("button").filter({ hasText: "Select a scale" }).click();
  await page.getByLabel("2").click();
  await page.locator("button").filter({ hasText: "Select a size" }).click();
  await page.getByLabel("medium").click();

  await page.getByLabel("Cmd").fill("hello");

  const envsElement = page
    .locator("fieldset")
    .filter({ hasText: "Environment variables" });

  await Promise.all(
    envVariables.map(async (item, index) => {
      await expect(
        envsElement.getByPlaceholder("NAME"),
        "it renders one set of inputs for every existing env +1 empty set"
      ).toHaveCount(index + 1);

      await envsElement.getByPlaceholder("NAME").last().fill(item.name);
      await envsElement.getByPlaceholder("VALUE").last().fill(item.value);
      await envsElement.getByRole("button").last().click();
    })
  );

  const editor = page.locator(".lines-content");

  await expect(
    editor,
    "all entered data is represented in the editor preview"
  ).toContainText(expectedYaml, { useInnerText: true });

  await expect(
    page.getByTestId("unsaved-note"),
    "it renders a hint that there are unsaved changes"
  ).toBeVisible();
  await page.getByRole("button", { name: "Save" }).click();

  await expect(
    page.getByTestId("unsaved-note"),
    "it does not render a hint that there are unsaved changes"
  ).not.toBeVisible();

  /* reload and assert data has been persisted */
  await page.reload({ waitUntil: "domcontentloaded" });

  await expect(
    editor,
    "after reloading, the entered data is still in the editor preview"
  ).toContainText(expectedYaml, { useInnerText: true });

  await expect(page.getByLabel("Image")).toHaveValue("bash");
  await expect(page.locator("button").filter({ hasText: "2" })).toBeVisible();
  await expect(
    page.locator("button").filter({ hasText: "medium" })
  ).toBeVisible();
  await expect(page.getByLabel("Cmd")).toHaveValue("hello");

  await Promise.all(
    envVariables.map(async (item, index) => {
      await expect(
        page.getByTestId("env-item-form").getByTestId("env-name").nth(index)
      ).toHaveValue(item.name);
      await expect(
        page.getByTestId("env-item-form").getByTestId("env-value").nth(index)
      ).toHaveValue(item.value);
    })
  );
});
