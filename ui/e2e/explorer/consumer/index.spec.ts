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

test("it is possible to create a consumer", async ({ page }) => {
  /* prepare data */
  const filename = "myconsumer.yaml";
  const groups = Array.from({ length: 5 }, () => faker.lorem.word());
  const tags = Array.from({ length: 2 }, () => faker.lorem.word());
  const groupsYaml = groups.map((group) => `\n  - "${group}"`).join("");
  const tagsYaml = tags.map((tag) => `\n  - "${tag}"`).join("");

  const expectedYaml = `direktiv_api: "consumer/v1"
  username: "my-username"
  password: "a-v3ry-g00d-pASSw0rd"
  api_key: "some-api-key"
  tags:${tagsYaml}
  groups:${groupsYaml}
  `;

  /* visit page */
  await page.goto(`/${namespace}/explorer/tree`, { waitUntil: "networkidle" });
  await expect(
    page.getByTestId("breadcrumb-namespace"),
    "it navigates to the test namespace in the explorer"
  ).toHaveText(namespace);

  /* create consumer */
  await page.getByRole("button", { name: "New" }).first().click();
  await page.getByRole("menuitem", { name: "Gateway" }).click();
  await page.getByRole("button", { name: "New Consumer" }).click();

  await expect(page.getByRole("button", { name: "Create" })).toBeDisabled();
  await page.getByPlaceholder("consumer-name.yaml").fill(filename);
  await page.getByRole("button", { name: "Create" }).click();

  await expect(
    page,
    "it creates the service and opens the file in the explorer"
  ).toHaveURL(`/${namespace}/explorer/consumer/${filename}`);

  /* fill in form */
  await page.getByLabel("Username").fill("my-username");
  await page.getByLabel("Password").fill("a-v3ry-g00d-pASSw0rd");
  await page.getByLabel("Api key").fill("some-api-key");

  const groupsElement = page.getByPlaceholder("Enter a group");

  await Promise.all(
    groups.map(async (group, index) => {
      await expect(
        groupsElement,
        "it renders one input for every existing group +1 empty one"
      ).toHaveCount(index + 1);

      await groupsElement.last().fill(group);
      await page
        .locator("fieldset")
        .filter({ hasText: "Groups (optional)" })
        .getByRole("button")
        .last()
        .click();
    })
  );

  const tagsElement = page.getByPlaceholder("Enter a tag");
  await Promise.all(
    tags.map(async (tag, index) => {
      await expect(
        tagsElement,
        "it renders one input for every existing tag +1 empty one"
      ).toHaveCount(index + 1);

      await tagsElement.last().fill(tag);
      await page
        .locator("fieldset")
        .filter({ hasText: "Tags (optional)" })
        .getByRole("button")
        .last()
        .click();
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

  await expect(page.getByLabel("Username")).toHaveValue("my-username");
  await expect(page.getByLabel("Password")).toHaveValue("a-v3ry-g00d-pASSw0rd");
  await expect(page.getByLabel("Api Key")).toHaveValue("some-api-key");

  await Promise.all(
    groups.map(async (group, index) => {
      await expect(
        page
          .locator("fieldset")
          .filter({ hasText: "Groups (optional)" })
          .getByRole("textbox")
          .nth(index)
      ).toHaveValue(group);
    })
  );

  await Promise.all(
    tags.map(async (group, index) => {
      await expect(
        page
          .locator("fieldset")
          .filter({ hasText: "Tags (optional)" })
          .getByRole("textbox")
          .nth(index)
      ).toHaveValue(group);
    })
  );
});
