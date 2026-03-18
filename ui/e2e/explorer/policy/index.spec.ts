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

test("it is possible to create a policy and view the result in the policy editor", async ({
  page,
}) => {
  /* prepare data */
  const filename = "mypolicy.yaml";
  const headlinevalue = "my-headline";
  const textvalue = "my-text";

  /* visit policy */
  await page.goto(`/n/${namespace}/explorer/tree`, {
    waitUntil: "networkidle",
  });
  await expect(
    page.getByTestId("breadcrumb-namespace"),
    "it navigates to the test namespace in the explorer"
  ).toHaveText(namespace);

  /* create policy */
  await page.getByRole("button", { name: "New" }).first().click();
  await page.getByRole("menuitem", { name: "policy" }).click();

  await expect(page.getByRole("button", { name: "Create" })).toBeDisabled();
  await page.getByPlaceholder("policy-name.yaml").fill(filename);
  await page.getByRole("button", { name: "Create" }).click();

  await expect(
    page,
    "it creates the policy and opens the file in the explorer"
  ).toHaveURL(`/n/${namespace}/explorer/policy/${filename}`);

  await expect(
    page.getByRole("heading", { level: 3, name: "mypolicy.yaml" }),
    "it displays the name of the policy file in the editor header"
  ).toBeVisible();

  await expect(
    page.getByText("Policy content is:"),
    "it shows the content of the policy file in the editor"
  ).toBeVisible();
});
