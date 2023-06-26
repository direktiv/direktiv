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
    name: faker.internet.domainWord(),
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
  const removedItemName = await itemName.nth(2).innerText();
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
    itemName.filter({ hasText: removedItemName }),
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
  await page.getByTestId("variable-create-card").click();
  await page.type("textarea", newVariable.value);
  await page.getByTestId("variable-trg-mimetype").click();
  await page.getByTestId(`var-mimetype-${newVariable.mimeType}`).click();
  await page.getByTestId("variable-create-submit").click();
  await actionWaitForSuccessToast(page);

  //reload page after create variable
  await page.reload({
    waitUntil: "networkidle",
  });

  //click on edit and confirm the created variable
  const subjectDropdownSelector = `dropdown-trg-item-${newVariable.name}`;
  await page.getByTestId(subjectDropdownSelector).click();
  await page.getByTestId("dropdown-actions-edit").click();

  await expect(
    page.getByTestId("variable-editor-card"),
    "the variable's content is loaded into the editor"
  ).toContainText(newVariable.value);

  await expect(
    page.locator("select"),
    "MimeTypeSelect is set to the subject's mimeType"
  ).toHaveValue(newVariable.mimeType || "");
  await page.getByTestId("var-edit-cancel").click();

  //delete one item
  const menuButtons = page.getByTestId(/dropdown-trg-item-/);
  await expect(menuButtons, "number of menuButtons should be 4").toHaveCount(4);
  const itemName = page.getByTestId("item-name");
  const removedItemName = await itemName.nth(2).innerText();

  await page
    .getByTestId(/dropdown-trg-item/)
    .nth(2)
    .click();
  await page.getByTestId("dropdown-actions-delete").click();
  await page.getByTestId("registry-delete-confirm").click();

  await actionWaitForSuccessToast(page);
  await expect(menuButtons, "number of menuButtons should be 3").toHaveCount(3);

  await expect(
    itemName.filter({ hasText: removedItemName }),
    "the deleted item shouldn't be in the list"
  ).toBeHidden();
  if (newVariable.name !== removedItemName) {
    await expect(
      itemName.filter({ hasText: newVariable.name }),
      "there should remain the newly created secret in the list"
    ).toBeVisible();
  }
});

test("it is possible to edit variables", async ({ page }) => {
  const variables = await createVariables(namespace, 3);
  const subject = variables[2];

  if (!subject) throw "There was an error setting up test data";

  const subjectDropdownSelector = `dropdown-trg-item-${subject.key}`;

  await page.goto(`/${namespace}/settings`);
  await page.getByTestId(subjectDropdownSelector).click();
  await page.getByTestId("dropdown-actions-edit").click();

  await expect(
    page.getByTestId("variable-editor-card"),
    "the variable's content is loaded into the editor"
  ).toContainText(subject.content);

  await expect(
    page.locator("select"),
    "MimeTypeSelect is set to the subject's mimeType"
  ).toHaveValue(subject.mimeType);

  // This was needed previously to make sure the editor is initialized
  // before updating the value, but it should no longer be needed.
  // await page.getByTestId("variable-editor-card").click();

  const textArea = page.getByRole("textbox");
  await textArea.type(faker.random.alphaNumeric(10));
  const updatedValue = await textArea.inputValue();
  const updatedType = options[Math.floor(Math.random() * 5)];
  await page.getByTestId("variable-trg-mimetype").click();
  await page.getByTestId(`var-mimetype-${updatedType}`).click();

  await page.getByTestId("var-edit-submit").click();
  await actionWaitForSuccessToast(page);
  await page.reload({
    waitUntil: "networkidle",
  });

  await page.getByTestId(subjectDropdownSelector).click();
  await page.getByTestId("dropdown-actions-edit").click();
  await page.getByTestId("variable-editor-card").click();

  await expect(
    page.getByTestId("variable-editor-card"),
    "the variable's content is loaded into the editor"
  ).toContainText(updatedValue);

  await expect(
    page.locator("select"),
    "MimeTypeSelect is set to the subject's mimeType"
  ).toHaveValue(updatedType || "");
});
