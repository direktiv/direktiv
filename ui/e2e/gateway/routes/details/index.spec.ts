import { createNamespace, deleteNamespace } from "e2e/utils/namespace";
import {
  createRouteFile,
  findRouteWithApiRequest,
  routeWithAnError,
} from "../utils";
import { expect, test } from "@playwright/test";

import { createFile } from "e2e/utils/files";
import { faker } from "@faker-js/faker";

let namespace = "";

test.beforeEach(async () => {
  namespace = await createNamespace();
});

test.afterEach(async () => {
  await deleteNamespace(namespace);
  namespace = "";
});

test("Route details page shows all important information about the route", async ({
  page,
}) => {
  const path = faker.lorem.word();
  const fileName = "my-route.yaml";
  await createFile({
    name: fileName,
    namespace,
    type: "endpoint",
    yaml: createRouteFile({
      path,
      targetType: "instant-response",
      targetConfigurationStatus: "202",
    }),
  });

  await expect
    .poll(
      async () =>
        await findRouteWithApiRequest({
          namespace,
          match: (route) => route.file_path === "/my-route.yaml",
        }),
      "the route was created and is available"
    )
    .toBeTruthy();

  await page.goto(`/n/${namespace}/gateway/routes`);

  page
    .getByTestId("route-table")
    .locator("div")
    .filter({ hasText: `/my-route.yaml` })
    .click();

  await expect(
    page,
    "it is possible to navigate to the route detail page"
  ).toHaveURL(`n/${namespace}/gateway/routes/my-route.yaml`);

  await expect(
    page
      .getByTestId("route-details-header")
      .getByRole("heading", { name: "/my-route.yaml" }),
    "it renders the text for the file path"
  ).toBeVisible();

  await page.getByTestId("route-details-header").getByText("9 methods").hover();

  const methods = [
    "connect",
    "get",
    "head",
    "options",
    "patch",
    "post",
    "put",
    "trace",
  ];

  for (const method of methods) {
    await expect(
      page.getByTestId("route-details-header").getByText(method)
    ).toBeVisible();
  }

  await expect(
    page
      .getByTestId("route-details-header")
      .locator("a")
      .filter({ hasText: "1 plugin" }),
    "it renders the correct number for plugins"
  ).toBeVisible();

  // hover over something else to make the overlay disappear
  page.getByTestId("route-details-header").getByText("yes").hover();

  await page
    .getByTestId("route-details-header")
    .locator("a")
    .filter({ hasText: "1 plugin" })
    .hover();

  await expect(
    page.getByTestId("route-details-header").getByText("1 target plugin"),
    "it shows the plugin details on hover"
  ).toBeVisible();

  await expect(
    page.getByTestId("route-details-header").getByText("yes"),
    "it renders the correct label for 'allow anonymous'"
  ).toBeVisible();

  await expect(
    page.getByTestId("route-details-header").getByRole("textbox"),
    "it renders the text for the path"
  ).toHaveValue(
    `${process.env.PLAYWRIGHT_UI_BASE_URL}/ns/${namespace}/${path}`
  );

  await expect(
    page.getByText("received 0 log entries"),
    "It does not have any log entries yet"
  ).toBeVisible();

  page
    .getByTestId("route-details-header")
    .getByRole("link", { name: "Edit Route" })
    .click();

  await expect(
    page,
    "when the edit route link is clicked, page should navigate to the route editor page"
  ).toHaveURL(`/n/${namespace}/explorer/endpoint/${fileName}`);
});

test("Route details page shows warning if the route was not configured correctly", async ({
  page,
}) => {
  const fileName = "my-route.yaml";

  await createFile({
    name: fileName,
    namespace,
    type: "endpoint",
    yaml: routeWithAnError,
  });

  await expect
    .poll(
      async () =>
        await findRouteWithApiRequest({
          namespace,
          match: (route) => route.file_path === "/my-route.yaml",
        }),
      "the route was created and is available"
    )
    .toBeTruthy();

  await page.goto(`/n/${namespace}/gateway/routes/${fileName}`);

  await page
    .getByTestId("route-details-header")
    .locator("a")
    .filter({ hasText: "1 error" })
    .hover();

  await expect(
    page.getByTestId("route-details-header").getByText("no valid http method")
  ).toBeVisible();

  await expect(
    page.getByTestId("route-details-header").getByText("no methods set"),
    "it renders the note that no methods are set"
  ).toBeVisible();
});
