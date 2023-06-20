import { createNamespace, deleteNamespace } from "../utils/namespace";
import { expect, test } from "@playwright/test";

import { MimeTypeSchema } from "~/pages/namespace/Settings/Variables/MimeTypeSelect";
import { actionWaitForSuccessToast } from "../explorer/workflow/utils";
import { createRegistries } from "../utils/registries";
import { createSecrets } from "../utils/secrets";
import { createVariables } from "../utils/variables";
import { faker } from "@faker-js/faker";

const { options } = MimeTypeSchema;

let namespace = "";

test.beforeEach(async () => {
  namespace = await createNamespace();
});

test.afterEach(async () => {
  await deleteNamespace(namespace);
  namespace = "";
});

test("it renders secrets, variables and registries", async ({ page }) => {
  const secrets = await createSecrets(namespace, 6);
  const registries = await createRegistries(namespace, 4);
  const variables = await createVariables(namespace);

  await page.goto(`/${namespace}/settings`);

  await expect(
    page
      .getByTestId("secrets-section")
      .getByRole("heading", { name: "Secrets" })
  ).toBeVisible();

  await expect(
    page.getByTestId("secrets-section").getByTestId("item-name")
  ).toHaveCount(secrets.length);

  await expect(
    page
      .getByTestId("registries-section")
      .getByRole("heading", { name: "Registries" })
  ).toBeVisible();

  await expect(
    page.getByTestId("registries-section").getByTestId("item-name")
  ).toHaveCount(registries.length);

  await expect(
    page
      .getByTestId("variables-section")
      .getByRole("heading", { name: "Variables" })
  ).toBeVisible();

  await expect(
    page.getByTestId("variables-section").getByTestId("item-name")
  ).toHaveCount(variables.length);
});

test("it is possible to create and delete secrets", async ({ page }) => {
  const defaultSecrets = await createSecrets(namespace, 3);
  await page.goto(`/${namespace}/settings`);
  await page.getByTestId("secret-create").click();
  const newSecret = {
    name: faker.random.alphaNumeric(7),
    value: faker.random.alphaNumeric(20),
  };
  await page.getByTestId("new-secret-name").type(newSecret.name);
  await page.getByTestId("new-secret-editor").type(newSecret.value);
  await page.getByTestId("secret-create-submit").click();
  await actionWaitForSuccessToast(page);

  const menuButtons = page.getByTestId(/dropdown-trg-item-/);
  await expect(menuButtons, "number of menuButtons should be 4").toHaveCount(4);

  await page.getByTestId(`dropdown-trg-item-${defaultSecrets[1]?.key}`).click();
  await page.getByTestId("dropdown-actions-delete").click();
  await page.getByTestId("secret-delete-confirm").click();

  await actionWaitForSuccessToast(page);
  await expect(menuButtons, "number of menuButtons should be 3").toHaveCount(3);
  const itemName = page.getByTestId("item-name");

  await expect(
    itemName.filter({ hasText: newSecret.name }),
    "there should remain the newly created secret in the list"
  ).toBeVisible();
  await expect(
    itemName.filter({ hasText: `${defaultSecrets[1]?.key}` }),
    "the deleted item shouldn't be in the list"
  ).toBeHidden();
});

test("it is possible to create and delete registries", async ({ page }) => {
  await createRegistries(namespace, 3);
  await page.goto(`/${namespace}/settings`);
  await page.getByTestId("registry-create").click();

  const newRegistry = {
    url: faker.internet.url(),
    user: faker.internet.userName(),
    password: faker.internet.password(),
  };

  await page.getByTestId("new-registry-url").type(newRegistry.url);
  await page.getByTestId("new-registry-pwd").type(newRegistry.password);
  await page.getByTestId("new-registry-user").type(newRegistry.user);

  await page.getByTestId("registry-create-submit").click();
  await actionWaitForSuccessToast(page);

  const menuButtons = page.getByTestId(/dropdown-trg-item-/);
  await expect(menuButtons, "number of menuButtons should be 4").toHaveCount(4);
  const itemName = page.getByTestId("item-name");
  const removing = await itemName.nth(2).innerText();
  await page
    .getByTestId(/dropdown-trg-item/)
    .nth(2)
    .click();
  await page.getByTestId("dropdown-actions-delete").click();
  await page.getByTestId("registry-delete-confirm").click();

  await actionWaitForSuccessToast(page);
  await expect(menuButtons, "number of menuButtons should be 3").toHaveCount(3);

  await expect(
    itemName.filter({ hasText: newRegistry.url }),
    "there should remain the newly created secret in the list"
  ).toBeVisible();
  await expect(
    itemName.filter({ hasText: removing }),
    "the deleted item shouldn't be in the list"
  ).toBeHidden();
});

test("it is possible to create and delete variables", async ({ page }) => {
  await createVariables(namespace, 3);
  await page.goto(`/${namespace}/settings`);
  await page.getByTestId("variable-create").click();

  const newVariable = {
    name: faker.random.word(),
    value: faker.random.words(20),
    mimeType: options[Math.floor(Math.random() * 5)],
  };
  await page.getByTestId("new-variable-name").type(newVariable.name);
  await page.waitForTimeout(5000);
  await page.type("textarea", newVariable.value);
  await page.getByTestId("variable-create-submit").click();
  await actionWaitForSuccessToast(page);

  const menuButtons = page.getByTestId(/dropdown-trg-item-/);
  await expect(menuButtons, "number of menuButtons should be 4").toHaveCount(4);
  const itemName = page.getByTestId("item-name");
  const removing = await itemName.nth(2).innerText();
  await page
    .getByTestId(/dropdown-trg-item/)
    .nth(2)
    .click();
  await page.getByTestId("dropdown-actions-delete").click();
  await page.getByTestId("registry-delete-confirm").click();

  await actionWaitForSuccessToast(page);
  await expect(menuButtons, "number of menuButtons should be 3").toHaveCount(3);

  await expect(
    itemName.filter({ hasText: removing }),
    "the deleted item shouldn't be in the list"
  ).toBeHidden();
  if (newVariable.name !== removing) {
    await expect(
      itemName.filter({ hasText: newVariable.name }),
      "there should remain the newly created secret in the list"
    ).toBeVisible();
  }
});

test("it is possible to edit variables", async ({ page }) => {
  await createVariables(namespace, 3);
  await page.goto(`/${namespace}/settings`);
  // const itemName = page.getByTestId("item-name");
  // const editing = await itemName.nth(2).innerText();
  await page
    .getByTestId(/dropdown-trg-item/)
    .nth(2)
    .click();
  await page.getByTestId("dropdown-actions-edit").click();
  // await page.getByTestId("registry-delete-confirm").click();
  const textArea = page.getByRole("textbox");
  await page.waitForTimeout(5000);
  await textArea.type(faker.random.alphaNumeric(10));
  // const updatedValue = textArea.inputValue();
  await page.getByTestId("var-edit-submit").click();
  actionWaitForSuccessToast(page);
});
