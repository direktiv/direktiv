import {
  PatchOperationType,
  PatchOperations,
  PatchSchemaType,
} from "~/pages/namespace/Explorer/Service/ServiceEditor/schema";
import { createNamespace, deleteNamespace } from "e2e/utils/namespace";
import { createService, createServiceYaml } from "./utils";
import { expect, test } from "@playwright/test";

import { EnvironementVariableSchemaType } from "~/api/services/schema/services";
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

  /**
   * note: keep number of variables and patches low because we only
   * compare the yaml that is visible in the editor at one time
   **/

  const envs = Array.from({ length: 3 }, () => ({
    name: faker.lorem.word(),
    value: faker.git.shortSha(),
  }));

  const patches = Array.from({ length: 2 }, () => ({
    op: PatchOperations[Math.floor(Math.random() * 3)] as PatchOperationType,
    path: faker.internet.url(),
    value: faker.lorem.words(3),
  }));

  const service = {
    name: "mynewservice.yaml",
    image: "bash",
    scale: 2,
    size: "medium",
    cmd: "hello",
    envs,
    patches,
  };

  const expectedYaml = createServiceYaml(service);

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
  await page.getByPlaceholder("service-name.yaml").fill(service.name);
  await page.getByRole("button", { name: "Create" }).click();

  await expect(
    page,
    "it creates the service and opens the file in the explorer"
  ).toHaveURL(`/${namespace}/explorer/service/${service.name}`);

  /* fill in form */
  await page.getByLabel("Image").fill("bash");
  await page.locator("button").filter({ hasText: "Select a scale" }).click();
  await page.getByLabel(service.scale.toString()).click();
  await page.locator("button").filter({ hasText: "Select a size" }).click();
  await page.getByLabel(service.size).click();

  await page.getByLabel("Cmd").fill(service.cmd);

  /* add patches */
  for (let i = 0; i < patches.length; i++) {
    const item = patches[i] as PatchSchemaType;
    await page.getByRole("button", { name: "add patch" }).click();
    await page.getByLabel("Operation").click();
    await page.getByLabel(item.op).click();
    await page.getByLabel("path").fill(item.path);
    await page.keyboard.press("Tab");
    await page.type("textarea", item.value);
    await page.getByRole("button", { name: "Save" }).click();
  }

  /* add env variables */
  const envsElement = page
    .locator("fieldset")
    .filter({ hasText: "Environment variables" });

  for (let i = 0; i < envs.length; i++) {
    const item = envs[i] as EnvironementVariableSchemaType;
    await expect(
      envsElement.getByPlaceholder("NAME"),
      "it renders one set of inputs for every existing env +1 empty set"
    ).toHaveCount(i + 1);

    await envsElement.getByPlaceholder("NAME").last().fill(item.name);
    await envsElement.getByPlaceholder("VALUE").last().fill(item.value);
    await envsElement.getByRole("button").last().click();
  }

  /**
   * assert preview editor content
   * note that only the visible part of the yaml is compared, so
   * this will fail if the document gets too long.
   */
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
    envs.map(async (item, index) => {
      const currentElement = page.getByTestId("env-item-form").nth(index);
      await expect(currentElement.getByTestId("env-name")).toHaveValue(
        item.name
      );
      await expect(currentElement.getByTestId("env-value")).toHaveValue(
        item.value
      );
    })
  );

  await expect(
    page.getByRole("cell", { name: `${patches.length} Patches` }),
    "It renders a table of patches, displaying the number of patches in the header"
  ).toBeVisible();

  await Promise.all(
    patches.map(async (item, index) => {
      const currentElement = page.getByTestId("patch-row").nth(index);
      await expect(currentElement).toContainText(item.op);
      await expect(currentElement).toContainText(item.path);
    })
  );
});

test("it is possible to edit patches", async ({ page }) => {
  /* prepare data */
  const patches = Array.from({ length: 4 }, () => ({
    op: PatchOperations[Math.floor(Math.random() * 3)] as PatchOperationType,
    path: faker.internet.url(),
    value: faker.lorem.words(3),
  }));

  const service = {
    name: "mynewservice.yaml",
    image: "bash",
    scale: 2,
    size: "medium",
    cmd: "hello",
    patches,
  };

  await createService(namespace, service);

  /* visit page, assert content rendered */
  await page.goto(`/${namespace}/explorer/service/${service.name}`);

  await Promise.all(
    patches.map(async (item, index) => {
      const currentElement = page.getByTestId("patch-row").nth(index);
      await expect(currentElement).toContainText(item.op);
      await expect(currentElement).toContainText(item.path);
    })
  );

  /* update list and assert content after each manipulation*/
  await page.getByTestId("patch-row").nth(1).getByRole("button").click();
  await page.getByRole("button", { name: "Move down" }).click();

  let expectNewPatches: PatchSchemaType[];

  expectNewPatches = [
    patches[0] as PatchSchemaType,
    patches[2] as PatchSchemaType,
    patches[1] as PatchSchemaType,
    patches[3] as PatchSchemaType,
  ];

  await Promise.all(
    expectNewPatches.map(async (item, index) => {
      const currentElement = page.getByTestId("patch-row").nth(index);
      await expect(currentElement).toContainText(item.op);
      await expect(currentElement).toContainText(item.path);
    })
  );

  await page.getByTestId("patch-row").nth(3).getByRole("button").click();
  await page.getByRole("button", { name: "Move up" }).click();

  expectNewPatches = [
    patches[0] as PatchSchemaType,
    patches[2] as PatchSchemaType,
    patches[3] as PatchSchemaType,
    patches[1] as PatchSchemaType,
  ];

  await Promise.all(
    expectNewPatches.map(async (item, index) => {
      const currentElement = page.getByTestId("patch-row").nth(index);
      await expect(currentElement).toContainText(item.op);
      await expect(currentElement).toContainText(item.path);
    })
  );

  await page.getByTestId("patch-row").nth(1).getByRole("button").click();
  await page.getByRole("button", { name: "Delete" }).click();

  expectNewPatches = [
    patches[0] as PatchSchemaType,
    patches[3] as PatchSchemaType,
    patches[1] as PatchSchemaType,
  ];

  await Promise.all(
    expectNewPatches.map(async (item, index) => {
      const currentElement = page.getByTestId("patch-row").nth(index);
      await expect(currentElement).toContainText(item.op);
      await expect(currentElement).toContainText(item.path);
    })
  );

  /* edit one patch */
  const updatedPatch: PatchSchemaType = {
    op: PatchOperations[Math.floor(Math.random() * 3)] as PatchOperationType,
    path: faker.internet.url(),
    value: faker.lorem.words(3),
  };

  const patchToEdit = expectNewPatches[1];

  if (!patchToEdit) throw Error("patch to edit is undefined");
  await page.getByTestId("patch-row").nth(1).click();

  await page.getByLabel("Operation").click();
  await page.getByLabel(updatedPatch.op).click();

  await page.getByLabel("Path").fill(updatedPatch.path);
  const editorTarget = await page.getByText(patchToEdit.value, {
    exact: true,
  });

  await editorTarget.click();
  await page.locator("textarea").last().fill(updatedPatch.value);

  await page.getByRole("button", { name: "Save" }).click();

  expectNewPatches = [
    patches[0] as PatchSchemaType,
    updatedPatch,
    patches[1] as PatchSchemaType,
  ];

  await Promise.all(
    expectNewPatches.map(async (item, index) => {
      const currentElement = page.getByTestId("patch-row").nth(index);
      await expect(currentElement).toContainText(item.op);
      await expect(currentElement).toContainText(item.path);
    })
  );

  /* assert preview has been updated */
  const updatedService = {
    name: "mynewservice.yaml",
    image: "bash",
    scale: 2,
    size: "medium",
    cmd: "hello",
    patches: expectNewPatches,
  };

  const expectedYaml = createServiceYaml(updatedService);

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
});

test("it is possible to edit environment variables", async ({ page }) => {
  const envs = Array.from({ length: 5 }, () => ({
    name: faker.lorem.word(),
    value: faker.git.shortSha(),
  }));

  const service = {
    name: "mynewservice.yaml",
    image: "bash",
    scale: 2,
    size: "medium",
    cmd: "hello",
    envs,
  };

  await createService(namespace, service);

  /* visit page, assert content rendered */
  await page.goto(`/${namespace}/explorer/service/${service.name}`);

  await Promise.all(
    envs.map(async (item, index) => {
      const currentElement = page.getByTestId("env-item-form").nth(index);
      await expect(currentElement.getByTestId("env-name")).toHaveValue(
        item.name
      );
      await expect(currentElement.getByTestId("env-value")).toHaveValue(
        item.value
      );
    })
  );

  /* edit one item */
  const updatedEnv: EnvironementVariableSchemaType = {
    name: faker.lorem.word(),
    value: faker.git.shortSha(),
  };

  const envToEdit = envs[3];

  if (!envToEdit) throw Error("env to edit is undefined");

  await page.getByTestId("env-name").nth(3).fill(updatedEnv.name);
  await page.getByTestId("env-value").nth(3).fill(updatedEnv.value);

  let expectNewEnvs: EnvironementVariableSchemaType[];

  expectNewEnvs = [
    envs[1] as EnvironementVariableSchemaType,
    envs[2] as EnvironementVariableSchemaType,
    updatedEnv,
    envs[4] as EnvironementVariableSchemaType,
  ];

  /* delete items and assert rendered list is updated*/
  await page.getByTestId("env-item-form").nth(0).getByRole("button").click();
  await page.getByTestId("env-item-form").nth(2).getByRole("button").click();

  expectNewEnvs = [
    envs[1] as EnvironementVariableSchemaType,
    envs[2] as EnvironementVariableSchemaType,
    envs[4] as EnvironementVariableSchemaType,
  ];

  await Promise.all(
    expectNewEnvs.map(async (item, index) => {
      const currentElement = page.getByTestId("env-item-form").nth(index);
      await expect(currentElement.getByTestId("env-name")).toHaveValue(
        item.name
      );
      await expect(currentElement.getByTestId("env-value")).toHaveValue(
        item.value
      );
    })
  );

  /* assert preview has been updated */
  const updatedService = {
    name: "mynewservice.yaml",
    image: "bash",
    scale: 2,
    size: "medium",
    cmd: "hello",
    envs: expectNewEnvs,
  };

  const expectedYaml = createServiceYaml(updatedService);

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
});
