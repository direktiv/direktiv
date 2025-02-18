import { createNamespace, deleteNamespace } from "../utils/namespace";
import { expect, test } from "@playwright/test";

import { createRegistries } from "../utils/registries";
import { createSecrets } from "../utils/secrets";
import { createVariables } from "../utils/variables";
import { decode } from "js-base64";
import { faker } from "@faker-js/faker";
import { radixClick } from "../utils/testutils";
import { waitForSuccessToast } from "../explorer/workflow/utils";

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

  await page.goto(`/n/${namespace}/settings`);

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
  const firstSecretName = secrets[0]?.data.name;
  const secretToDelete = secrets[1];

  // avoid typescript errors below
  if (!secretToDelete || !firstSecretName) throw "error setting up test data";

  await page.goto(`/n/${namespace}/settings`);
  await page.getByTestId("secret-create").click();
  const newSecret = {
    name: faker.internet.domainWord(),
    value: faker.string.alphanumeric(20),
  };

  await page.getByPlaceholder("secret-name").type(firstSecretName);
  await page.locator("textarea").type(newSecret.value);
  await page.getByRole("button", { name: "Create" }).click();

  await expect(
    page.getByText("The name already exists"),
    "it renders an error message when using a name that already exists"
  ).toBeVisible();

  await page.getByPlaceholder("secret-name").type(newSecret.name);
  await page.getByRole("button", { name: "Create" }).click();

  await waitForSuccessToast(page);

  const secretElements = page.getByTestId("item-name");
  await expect(secretElements, "number of secrets should be 4").toHaveCount(4);

  await page
    .getByTestId(`dropdown-trg-item-${secretToDelete.data.name}`)
    .click();

  await page.getByTestId("dropdown-actions-delete").click();
  await page.getByRole("button", { name: "Delete" }).click();

  await waitForSuccessToast(page);
  await expect(secretElements, "number of secrets should be 3").toHaveCount(3);

  await expect(
    secretElements.filter({ hasText: newSecret.name }),
    "there should remain the newly created secret in the list"
  ).toBeVisible();
  await expect(
    secretElements.filter({ hasText: `${secretToDelete.data.name}` }),
    "the deleted item shouldn't be in the list"
  ).toBeHidden();
});

test("secrets are displayed in alphabetical order", async ({ page }) => {
  await page.goto(`/n/${namespace}/settings`);

  // Firt Create secrets in non-alphabetical order
  const secretNames = ["X", "T", "A", "01"];

  for (const name of secretNames) {
    await page.getByTestId("secret-create").click();
    await page.getByPlaceholder("secret-name").type(name);
    await page.locator("textarea").type("test-value");
    await page.getByRole("button", { name: "Create" }).click();
    await waitForSuccessToast(page);
  }

  // Then get all secret names in the list
  const secretElements = page
    .getByTestId("secrets-section")
    .getByTestId("item-name");
  const displayedNames = await secretElements.allTextContents();

  // And check if the names are in alphabetical order
  expect(displayedNames).toEqual(["01", "A", "T", "X"]);
});

test("it is possible to create and delete registries", async ({ page }) => {
  const registries = await createRegistries(namespace, 3);
  const registryToDelete = registries[2];

  // avoid typescript errors below
  if (!registryToDelete) throw "error setting up test data";

  await page.goto(`/n/${namespace}/settings`);
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
  await waitForSuccessToast(page);

  const registryElements = page.getByTestId("item-name");
  await expect(
    registryElements,
    "number of registry elements rendered should be 4"
  ).toHaveCount(4);

  await page.getByTestId(`dropdown-trg-item-${registryToDelete.url}`).click();
  await page.getByTestId("dropdown-actions-delete").click();
  await page.getByTestId("registry-delete-confirm").click();

  await waitForSuccessToast(page);
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
  /* set up test data */
  const variables = await createVariables(namespace, 3);
  const variableToDelete = variables[2];

  if (!variableToDelete) throw "error setting up test data";

  /* visit page and edit variable*/
  await page.goto(`/n/${namespace}/settings`);
  await page.getByTestId("variable-create").click();

  await page.getByTestId("variable-name").type("awesome-variable");

  await page.locator(".view-lines").click();
  await page.locator(".view-lines").type("<div>Hello world</div>");
  await page.getByLabel("Mimetype").click();
  await page.getByLabel("HTML").click();
  await page.getByRole("button", { name: "create" }).click();
  await waitForSuccessToast(page);

  /* reload and make sure changes have been persisted */
  await page.reload({
    waitUntil: "networkidle",
  });

  await page.getByTestId("dropdown-trg-item-awesome-variable").click();
  await page.getByTestId("dropdown-actions-edit").click();

  await expect(
    page.locator(".view-lines"),
    "the variable's content is loaded into the editor"
  ).toContainText("<div>Hello world</div>");

  await expect(
    page.locator("select"),
    "MimeTypeSelect is set to the subject's mimeType"
  ).toHaveValue("text/html");

  const cancelButton = page.getByRole("button", { name: "Cancel" });
  await radixClick(browserName, cancelButton);

  /* delete one variable */
  await expect(
    page.getByTestId("item-name"),
    "there are 4 variables"
  ).toHaveCount(4);

  await page
    .getByTestId(`dropdown-trg-item-${variableToDelete.data.name}`)
    .click();
  await page.getByTestId("dropdown-actions-delete").click();
  await page.getByTestId("registry-delete-confirm").click();

  await waitForSuccessToast(page);
  await expect(
    page.getByTestId("item-name"),
    "after deleting a variable, there are 3 variables left"
  ).toHaveCount(3);

  await expect(
    page
      .getByTestId("item-name")
      .filter({ hasText: variableToDelete.data.name }),
    "the deleted variable is no longer in the list"
  ).toHaveCount(0);

  await expect(
    page.getByTestId("item-name").filter({ hasText: "awesome-variable" }),
    "the new variable is still in the list"
  ).toHaveCount(1);
});

test("bulk delete variables", async ({ page }) => {
  /* set up test data */
  const variables = await createVariables(namespace, 4);
  const variableToDelete = variables[3];

  if (!variableToDelete) throw "error setting up test data";

  await page.goto(`/n/${namespace}/settings`);

  // Check first checkbox and click delete
  await page.getByTestId("item-name").getByRole("checkbox").first().check();
  await page.getByRole("button", { name: "Delete selected" }).click();
  await expect(
    page.getByText(`Are you sure you want to delete variable`, { exact: false })
  ).toBeVisible();
  await page.getByRole("button", { name: "Delete" }).click();
  await waitForSuccessToast(page);
  await expect(page.getByTestId("item-name")).toHaveCount(3);

  // Check second checkbox and click delete
  await page.getByTestId("item-name").getByRole("checkbox").nth(0).check();
  await page.getByTestId("item-name").getByRole("checkbox").nth(1).check();
  await page.getByRole("button", { name: "Delete selected" }).click();
  await expect(
    page.getByText("Are you sure you want to delete 2 variables?", {
      exact: true,
    })
  ).toBeVisible();
  await page.getByRole("button", { name: "Cancel" }).click();

  // Check third checkbox and click delete
  await page.getByTestId("item-name").getByRole("checkbox").nth(0).check();
  await page.getByTestId("item-name").getByRole("checkbox").nth(1).check();
  await page.getByTestId("item-name").getByRole("checkbox").nth(2).check();
  await page.getByRole("button", { name: "Delete selected" }).click();
  await expect(
    page.getByText("Are you sure you want to delete all variables?", {
      exact: true,
    })
  ).toBeVisible();

  // Confirm deletion
  await page.getByRole("button", { name: "Delete" }).click();
  await waitForSuccessToast(page);

  await expect(page.getByTestId("item-name")).toHaveCount(0);
});

test("it is possible to edit variables", async ({ page }) => {
  /* set up test data */
  const variables = await createVariables(namespace, 3);
  const subject = variables[2];

  const newName = "new-name";

  if (!subject) throw "There was an error setting up test data";

  /* visit page and edit variable */
  await page.goto(`/n/${namespace}/settings`);
  await page.getByTestId(`dropdown-trg-item-${subject.data.name}`).click();
  await page.getByRole("button", { name: "edit" }).click();

  const textArea = page
    .getByTestId("variable-editor-card")
    .getByRole("textbox");

  await expect
    .poll(
      async () => await textArea.inputValue(),
      "the variable's content is loaded into the editor"
    )
    .toBe(decode(subject.content));

  await page.getByTestId("variable-name").fill(newName);

  await expect(
    page.locator("select"),
    "MimeTypeSelect is set to the subject's mimeType"
  ).toHaveValue(subject.data.mimeType);

  await page.locator(".view-lines").click();
  await page.locator(".view-lines").click();

  /* delete and replace existing content */
  for (let i = 0; i < subject.content.length; i++) {
    await page.locator(".view-lines").press("Backspace");
  }

  await page.locator(".view-lines").type("data: this is supposed to be YAML");
  await page.getByLabel("Mimetype").click();
  await page.getByLabel("YAML").click();

  await page.getByRole("button", { name: "save" }).click();
  await waitForSuccessToast(page);

  /* reload and make sure changes have been persisted */
  await page.reload({
    waitUntil: "networkidle",
  });

  await page.getByTestId(`dropdown-trg-item-${newName}`).click();
  await page.getByRole("button", { name: "edit" }).click();

  await expect
    .poll(
      async () => await textArea.inputValue(),
      "the updated variable content is loaded into the editor"
    )
    .toBe("data: this is supposed to be YAML");

  await expect(
    await page.getByTestId("variable-name").inputValue(),
    "the updated variable name is shown in the name input"
  ).toBe(newName);

  await expect(
    page.locator("select"),
    "MimeTypeSelect is set to the updated mimeType"
  ).toHaveValue("application/yaml");
});

test("it is not possible to create a variable with a name that already exists", async ({
  page,
}) => {
  /* set up test data */
  const variables = await createVariables(namespace, 3);
  const reservedName = variables[0]?.data.name ?? "";

  await page.goto(`/n/${namespace}/settings`);
  await page.getByTestId("variable-create").click();

  page.getByTestId("variable-name").fill(reservedName);

  await page.getByRole("button", { name: "Create" }).click();

  await expect(
    page.getByText("The name already exists"),
    "it renders an error message"
  ).toBeVisible();
});

test("it is not possible to set a variables name to a name that already exists", async ({
  page,
}) => {
  /* set up test data */
  const variables = await createVariables(namespace, 3);
  const subject = variables[2];

  if (!subject) throw "There was an error setting up test data";

  const reservedName = variables[0]?.data.name ?? "";

  await page.goto(`/n/${namespace}/settings`);
  await page.getByTestId(`dropdown-trg-item-${subject.data.name}`).click();
  await page.getByRole("button", { name: "edit" }).click();

  page.getByTestId("variable-name").fill(reservedName);

  await page.getByRole("button", { name: "Save" }).click();

  await expect(
    page.getByText("The name already exists"),
    "it renders an error message"
  ).toBeVisible();
});
