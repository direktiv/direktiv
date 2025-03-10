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

  /* visit page */
  await page.goto(`/n/${namespace}/explorer/tree`, {
    waitUntil: "load",
  });
  await expect(
    page.getByTestId("breadcrumb-namespace"),
    "it navigates to the test namespace in the explorer"
  ).toHaveText(namespace);

  /* create consumer */
  await page.getByRole("button", { name: "New" }).first().click();
  await page.getByRole("menuitem", { name: "Gateway" }).focus();
  await page.keyboard.press("Enter");
  await page.getByRole("button", { name: "Consumer" }).focus();
  await page.keyboard.press("Enter");

  await expect(page.getByRole("button", { name: "Create" })).toBeDisabled();
  await page.getByPlaceholder("consumer-name.yaml").fill(filename);
  await page.getByRole("button", { name: "Create" }).click();

  await expect(
    page,
    "it creates the consumer and opens the file in the explorer"
  ).toHaveURL(`/n/${namespace}/explorer/consumer/${filename}`);

  /* fill in form */
  await page.getByLabel("Username").fill("my-username");
  await page.getByLabel("Password").fill("a-v3ry-g00d-pASSw0rd");
  await page.getByLabel("Api key").fill("some-api-key");

  const groupsElement = page.getByPlaceholder("Enter a group");

  for (let i = 0; i < groups.length; i++) {
    const group = groups?.[i];
    if (!group) {
      throw new Error("group is undefined");
    }

    await expect(
      groupsElement,
      "it renders one input for every existing group +1 empty one"
    ).toHaveCount(i + 1);

    await expect(
      groupsElement,
      "it renders one input for every existing group +1 empty one"
    ).toHaveCount(i + 1);
    await groupsElement.last().fill(group);
    await page
      .locator("fieldset")
      .filter({ hasText: "Groups (optional)" })
      .getByRole("button")
      .last()
      .focus();
    await page.keyboard.press("Enter");
  }

  const tagsElement = page.getByPlaceholder("Enter a tag");

  for (let i = 0; i < tags.length; i++) {
    const tag = tags?.[i];
    if (!tag) {
      throw new Error("tag is undefined");
    }

    await expect(
      tagsElement,
      "it renders one input for every existing tag +1 empty one"
    ).toHaveCount(i + 1);
    await tagsElement.last().fill(tag);
    await page
      .locator("fieldset")
      .filter({ hasText: "Tags (optional)" })
      .getByRole("button")
      .last()
      .focus();
    await page.keyboard.press("Enter");
  }

  const editor = page.locator(".lines-content");

  function normalizeText(text: string) {
    return text.replace(/\s+/g, " ").trim();
  }

  const actualText = normalizeText(await editor.innerText());

  const keyElements = [
    "direktiv_api: consumer/v1",
    "username: my-username",
    "password: a-v3ry-g00d-pASSw0rd",
    "api_key: some-api-key",
    "tags:",
    "groups:",
  ];

  for (const element of keyElements) {
    expect(actualText).toContain(element);
  }

  const actualGroups = groups.map((group) => `- ${group}`);
  const actualTags = tags.map((tag) => `- ${tag}`);

  for (const group of actualGroups) {
    expect(actualText).toContain(group);
  }

  for (const tag of actualTags) {
    expect(actualText).toContain(tag);
  }

  await expect(
    page.getByTestId("unsaved-note"),
    "it renders a hint that there are unsaved changes"
  ).toBeVisible();

  const saveButton = page.getByRole("button", { name: "Save" });

  await saveButton.focus();
  await saveButton.waitFor({ state: "visible" });
  await page.keyboard.press("Enter");

  await expect(
    page.getByTestId("unsaved-note"),
    "it does not render a hint that there are unsaved changes"
  ).not.toBeVisible();

  /* reload and assert data has been persisted */
  await page.reload({ waitUntil: "domcontentloaded" });

  await editor.waitFor({ state: "visible" });

  // Wait a bit more for content to render (WebKit needs this)
  await page.waitForTimeout(500);

  const actualTextAfterReload = normalizeText(await editor.innerText());

  const fixedElements = [
    "direktiv_api: consumer/v1",
    "username: my-username",
    "password: a-v3ry-g00d-pASSw0rd",
    "api_key: some-api-key",
  ];

  for (const element of fixedElements) {
    expect(actualTextAfterReload).toContain(element);
  }

  for (const group of groups) {
    expect(actualTextAfterReload).toContain(`- ${group}`);
  }

  for (const tag of tags) {
    expect(actualTextAfterReload).toContain(`- ${tag}`);
  }

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
