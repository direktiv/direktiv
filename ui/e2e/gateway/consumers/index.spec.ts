import { createNamespace, deleteNamespace } from "e2e/utils/namespace";
import { createRedisConsumerFile, findConsumerWithApiRequest } from "./utils";
import { expect, test } from "@playwright/test";

import { createFile } from "e2e/utils/files";
import { encode } from "js-base64";
import { headers } from "e2e/utils/testutils";
import { patchFile } from "~/api/files/mutate/patchFile";

let namespace = "";

test.beforeEach(async () => {
  namespace = await createNamespace();
});

test.afterEach(async () => {
  await deleteNamespace(namespace);
  namespace = "";
});

test("The consumer list can be visited", async ({ page }) => {
  await page.goto(`/n/${namespace}/gateway/routes`, {
    waitUntil: "networkidle",
  });

  await page.getByRole("tab", { name: "Consumers" }).click();

  await expect(
    page,
    "it is possible to navigate to Consumers by breadcrumb"
  ).toHaveURL(`n/${namespace}/gateway/consumers`);

  await expect(
    page.getByTestId("breadcrumb-gateway"),
    "it renders the 'Gateway' breadcrumb"
  ).toBeVisible();

  await expect(
    page.getByTestId("breadcrumb-consumers"),
    "it renders the 'Consumers' breadcrumb"
  ).toBeVisible();

  await expect(
    page.getByText("No consumers exist yet"),
    "it renders an empty list of consumers"
  ).toBeVisible();

  await page.getByRole("tab", { name: "Routes" }).click();

  await expect(
    page,
    "it is possible to navigate to Routes by breadcrumb"
  ).toHaveURL(`n/${namespace}/gateway/routes`);

  await expect(
    page.getByTestId("breadcrumb-gateway"),
    "it renders the 'Gateway' breadcrumb"
  ).toBeVisible();

  await expect(
    page.getByTestId("breadcrumb-routes"),
    "it renders the 'Routes' breadcrumb"
  ).toBeVisible();
});

test("Consumer list shows all available consumers", async ({ page }) => {
  await createFile({
    name: "redis-consumer.yaml",
    namespace,
    type: "consumer",
    yaml: createRedisConsumerFile({
      username: "userA",
      password: "password",
    }),
  });

  await expect
    .poll(
      async () =>
        await findConsumerWithApiRequest({
          namespace,
          match: (consumer) => consumer.username === "userA",
        }),
      "the consumer was created and is available"
    )
    .toBeTruthy();

  await page.goto(`/n/${namespace}/gateway/consumers`, {
    waitUntil: "networkidle",
  });

  await expect(
    page.getByTestId("consumer-table").locator("tr"),
    "it renders one row of consumers"
  ).toHaveCount(1);

  await expect(
    page.getByTestId("consumer-table").getByRole("cell").nth(0).locator("div"),
    "it renders the text for the username"
  ).toHaveText("userA");

  await expect(
    page
      .getByTestId("consumer-table")
      .locator("tr")
      .getByRole("textbox")
      .first(),
    "it renders the text for the password"
  ).toHaveValue("password");

  await expect(
    page
      .getByTestId("consumer-table")
      .locator("tr")
      .getByRole("textbox")
      .nth(1),
    "it renders the text for the apikey"
  ).toHaveValue("123456789");

  await expect(
    page.getByTestId("consumer-groups").locator("div"),
    "it renders exactly two groups"
  ).toHaveCount(2);

  await expect(
    page.getByTestId("consumer-groups").locator("div").first(),
    "it renders the text of the first group"
  ).toHaveText("group1");

  await expect(
    page.getByTestId("consumer-tags").locator("div"),
    "it renders exactly one tag"
  ).toHaveCount(1);

  await expect(
    page.getByTestId("consumer-tags").locator("div"),
    "it renders the text of the first tag"
  ).toHaveText("tag1");
});

test("Consumer list will update the consumers when refetch button is clicked", async ({
  page,
}) => {
  await createFile({
    name: "consumer.yaml",
    namespace,
    type: "consumer",
    yaml: createRedisConsumerFile({
      username: "userOld",
      password: "passwordOld",
    }),
  });

  await page.goto(`/n/${namespace}/gateway/consumers`, {
    waitUntil: "networkidle",
  });

  await expect(
    page
      .getByTestId("consumer-table")
      .getByRole("cell", { name: "userOld", exact: true }),
    "it shows the (old) username"
  ).toBeVisible();

  await expect(
    page
      .getByTestId("consumer-table")
      .locator("tr")
      .getByRole("textbox")
      .first(),
    "it shows the (old) password"
  ).toHaveValue("passwordOld");

  await patchFile({
    payload: {
      data: encode(
        createRedisConsumerFile({
          username: "userNew",
          password: "passwordNew",
        })
      ),
    },
    urlParams: {
      baseUrl: process.env.PLAYWRIGHT_UI_BASE_URL,
      namespace,
      path: "/consumer.yaml",
    },
    headers,
  });

  await page.waitForTimeout(500);

  await page.getByLabel("Refetch consumers").click();

  await expect(
    page
      .getByTestId("consumer-table")
      .getByRole("cell", { name: "userNew", exact: true }),
    "it has updated the rendered username to the new value"
  ).toBeVisible();

  await expect(
    page
      .getByTestId("consumer-table")
      .locator("tr")
      .getByRole("textbox")
      .first(),
    "it has updated the rendered password to the new value"
  ).toHaveValue("passwordNew");
});
