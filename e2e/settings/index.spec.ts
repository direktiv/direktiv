import { createNamespace, deleteNamespace } from "../utils/namespace";
import { expect, test } from "@playwright/test";

import { BroadcastsSchemaKeys } from "~/api/broadcasts/schema";
import { EditorMimeTypeSchema } from "~/pages/namespace/Settings/Variables/MimeTypeSelect";
import { actionWaitForSuccessToast } from "../explorer/workflow/utils";
import { createBroadcasts } from "../utils/broadcasts";
import { createRegistries } from "../utils/registries";
import { createSecrets } from "../utils/secrets";
import { createVariables } from "../utils/variables";
import { faker } from "@faker-js/faker";
import { radixClick } from "../utils/testutils";

const { options } = EditorMimeTypeSchema;

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
  const secrets = await createSecrets(namespace, 3);
  const secretToDelete = secrets[1];

  // avoid typescript errors below
  if (!secretToDelete) throw "error setting up test data";

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

  const secretElements = page.getByTestId("item-name");
  await expect(secretElements, "number of secrets should be 4").toHaveCount(4);

  await page.getByTestId(`dropdown-trg-item-${secretToDelete.key}`).click();
  await page.getByTestId("dropdown-actions-delete").click();
  await page.getByTestId("secret-delete-confirm").click();

  await actionWaitForSuccessToast(page);
  await expect(secretElements, "number of secrets should be 3").toHaveCount(3);

  await expect(
    secretElements.filter({ hasText: newSecret.name }),
    "there should remain the newly created secret in the list"
  ).toBeVisible();
  await expect(
    secretElements.filter({ hasText: `${secretToDelete.key}` }),
    "the deleted item shouldn't be in the list"
  ).toBeHidden();
});

test("it is possible to create and delete registries", async ({ page }) => {
  const registries = await createRegistries(namespace, 3);
  const registryToDelete = registries[2];

  // avoid typescript errors below
  if (!registryToDelete) throw "error setting up test data";

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

  const registryElements = page.getByTestId("item-name");
  await expect(
    registryElements,
    "number of registry elements rendered should be 4"
  ).toHaveCount(4);

  await page.getByTestId(`dropdown-trg-item-${registryToDelete.url}`).click();
  await page.getByTestId("dropdown-actions-delete").click();
  await page.getByTestId("registry-delete-confirm").click();

  await actionWaitForSuccessToast(page);
  await expect(
    registryElements,
    "number of registry elements rendered should be 3"
  ).toHaveCount(3);

  await expect(
    registryElements.filter({ hasText: newRegistry.url }),
    "there should remain the newly created registry in the list"
  ).toHaveCount(1);
  await expect(
    registryElements.filter({ hasText: registryToDelete.url }),
    "the deleted item shouldn't be in the list"
  ).toHaveCount(0);
});

test("it is possible to create and delete variables", async ({
  page,
  browserName,
}) => {
  // set up test data
  const variables = await createVariables(namespace, 3);
  const variableToDelete = variables[2];

  const newVariable = {
    name: faker.internet.domainWord(),
    value: faker.random.words(20),
    mimeType: options[Math.floor(Math.random() * options.length)] || options[0],
  };

  // handle error to avoid typescript errors below
  if (!variableToDelete) throw "error setting up test data";

  // perform test
  await page.goto(`/${namespace}/settings`);
  await page.getByTestId("variable-create").click();

  await page.getByTestId("new-variable-name").type(newVariable.name);

  const editor = page.locator(".view-lines");
  await editor.click();

  await editor.type(newVariable.value);
  await page.getByTestId("variable-trg-mimetype").click();
  await page.getByTestId(`var-mimetype-${newVariable.mimeType}`).click();
  await page.getByTestId("variable-create-submit").click();
  await actionWaitForSuccessToast(page);

  // reload page after create variable
  await page.reload({
    waitUntil: "networkidle",
  });

  // click on edit and confirm the created variable
  const subjectDropdownSelector = `dropdown-trg-item-${newVariable.name}`;
  await page.getByTestId(subjectDropdownSelector).click();
  await page.getByTestId("dropdown-actions-edit").click();

  await expect(
    editor,
    "the variable's content is loaded into the editor"
  ).toContainText(newVariable.value);

  await expect(
    page.locator("select"),
    "MimeTypeSelect is set to the subject's mimeType"
  ).toHaveValue(newVariable.mimeType);

  const cancelButton = page.getByTestId("var-edit-cancel");
  await radixClick(browserName, cancelButton);

  //delete one item
  await expect(
    page.getByTestId("item-name"),
    "there are 4 variables"
  ).toHaveCount(4);

  await page.getByTestId(`dropdown-trg-item-${variableToDelete.key}`).click();
  await page.getByTestId("dropdown-actions-delete").click();
  await page.getByTestId("registry-delete-confirm").click();

  await actionWaitForSuccessToast(page);
  await expect(
    page.getByTestId("item-name"),
    "after deleting a variable, there are 3 variables left"
  ).toHaveCount(3);

  await expect(
    page.getByTestId("item-name").filter({ hasText: variableToDelete.key }),
    "the deleted variable is no longer in the list"
  ).toHaveCount(0);

  await expect(
    page.getByTestId("item-name").filter({ hasText: newVariable.name }),
    "the new variable is still in the list"
  ).toHaveCount(1);
});

test("it is possible to edit variables", async ({ page }) => {
  const variables = await createVariables(namespace, 3);
  const subject = variables[2];

  if (!subject) throw "There was an error setting up test data";

  const subjectDropdownSelector = `dropdown-trg-item-${subject.key}`;

  await page.goto(`/${namespace}/settings`);
  await page.getByTestId(subjectDropdownSelector).click();
  await page.getByTestId("dropdown-actions-edit").click();

  const textArea = page
    .getByTestId("variable-editor-card")
    .getByRole("textbox");
  await expect
    .poll(
      async () => await textArea.inputValue(),
      "the variable's content is loaded into the editor"
    )
    .toBe(subject.content);

  await expect(
    page.locator("select"),
    "MimeTypeSelect is set to the subject's mimeType"
  ).toHaveValue(subject.mimeType);

  await textArea.type(faker.random.alphaNumeric(10));
  const updatedValue = await textArea.inputValue();
  const updatedType =
    options[Math.floor(Math.random() * options.length)] || options[0];
  await page.getByTestId("variable-trg-mimetype").click();
  await page.getByTestId(`var-mimetype-${updatedType}`).click();

  await page.getByTestId("var-edit-submit").click();
  await actionWaitForSuccessToast(page);
  await page.reload({
    waitUntil: "networkidle",
  });

  await page.getByTestId(subjectDropdownSelector).click();
  await page.getByTestId("dropdown-actions-edit").click();

  await expect
    .poll(
      async () => await textArea.inputValue(),
      "the updated variable content is loaded into the editor"
    )
    .toBe(updatedValue);

  await expect(
    page.locator("select"),
    "MimeTypeSelect is set to the updated mimeType"
  ).toHaveValue(updatedType);
});

test("it is possible to update broadcasts", async ({ page }) => {
  const { broadcast: broadcasts } = await createBroadcasts(namespace);
  expect(broadcasts, "test data has been created").toBeTruthy();

  await page.goto(`/${namespace}/settings`);

  // check the initial state
  for (let i = 0; i < BroadcastsSchemaKeys.length; i++) {
    const key = BroadcastsSchemaKeys[i] || "directory.create";
    const checkbox = page.getByTestId(`check.${key}`);
    const isChecked = await checkbox.isChecked();
    expect(isChecked, `checkbox for ${key} has the expected value`).toBe(
      broadcasts[key]
    );
  }

  // update random fields and check their status
  const randomElements = faker.helpers.arrayElements(BroadcastsSchemaKeys, 3);

  for (let i = 0; i < randomElements.length; i++) {
    const key = randomElements[i] || "directory.create";
    const checkbox = page.getByTestId(`check.${key}`);
    await checkbox.click();

    // expect() without .poll() would fail, because it would not wait for the DOM to update
    await expect
      .poll(
        async () => await checkbox.isChecked(),
        `checkbox for ${key} has been toggled to the inverse value: `
      )
      .toBe(!broadcasts[key]);
  }
});

test("it is possible to sucessfully test the connection when creating a registry", async ({
  page,
}) => {
  await page.goto(`/${namespace}/settings`);
  await page.getByTestId("registry-create").click();
  const testConnectionButton = page.getByTestId("btn-test-connection");

  const validRegistryUrl = "https://hub.docker.com/r/stefandirektiv/test";
  const user = faker.internet.userName();
  const password = faker.internet.password();

  await page.getByTestId("new-registry-url").fill(validRegistryUrl);
  await page.getByTestId("new-registry-user").fill(user);

  await expect(
    testConnectionButton,
    "the test connection button should be disabled until we fill in the url, user and password"
  ).toBeDisabled();

  await page.getByTestId("new-registry-pwd").fill(password);

  await expect(
    testConnectionButton,
    "the test connection button should be enabled after we filled in url, user and password"
  ).toBeEnabled();

  await expect(
    testConnectionButton,
    'the test connection button should be labeled "Test connection" by default'
  ).toHaveText("Test connection");

  await testConnectionButton.click();

  await expect(
    testConnectionButton,
    "the test connection button should be disabled while checking"
  ).toBeDisabled();

  await expect(
    testConnectionButton,
    'the test connection button should be labeled "Connection successful" when the test was successful'
  ).toHaveText("Connection successful");

  await expect(
    testConnectionButton,
    "the test connection button should be enabled after the check finished"
  ).toBeEnabled();

  await expect(
    testConnectionButton,
    'the test connection button should be labeled "Test connection" some time after the check finished'
  ).toHaveText("Test connection");
});

test("it is possible to unsucessfully test the connection when creating a registry", async ({
  page,
}) => {
  await page.goto(`/${namespace}/settings`);
  await page.getByTestId("registry-create").click();
  const testConnectionButton = page.getByTestId("btn-test-connection");

  const invalidRegistryUrl = "https://google.com";
  const user = faker.internet.userName();
  const password = faker.internet.password();

  await page.getByTestId("new-registry-url").fill(invalidRegistryUrl);
  await page.getByTestId("new-registry-user").fill(user);
  await page.getByTestId("new-registry-pwd").fill(password);

  await expect(
    testConnectionButton,
    'the test connection button should be labeled "Test connection" by default'
  ).toHaveText("Test connection");

  await testConnectionButton.click();

  await expect(
    testConnectionButton,
    "the test connection button should be disabled while checking"
  ).toBeDisabled();

  await expect(
    testConnectionButton,
    'the test connection button should be labeled "Connection failed" when the test was not successful'
  ).toHaveText("Connection failed");

  await expect(
    testConnectionButton,
    "the test connection button should be enabled after the check finished"
  ).toBeEnabled();

  await expect(
    testConnectionButton,
    'the test connection button should be labeled "Test connection" some time after the check finished'
  ).toHaveText("Test connection");
});
