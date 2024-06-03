import { createNamespace, deleteNamespace } from "../../utils/namespace";
import { expect, test } from "@playwright/test";

import { createFile } from "e2e/utils/files";
import { createInstance } from "../utils/index";
import { faker } from "@faker-js/faker";
import { simpleWorkflow as simpleWorkflowContent } from "../utils/workflows";

let namespace = "";

test.beforeEach(async () => {
  namespace = await createNamespace();
});

test.afterEach(async () => {
  await deleteNamespace(namespace);
  namespace = "";
});

test("it renders an error when the api response returns an error", async ({
  page,
}) => {
  /* prepare data */
  const simpleWorkflowName = faker.system.commonFileName("yaml");

  await createFile({
    name: simpleWorkflowName,
    namespace,
    type: "workflow",
    yaml: simpleWorkflowContent,
  });

  const instanceId = (
    await createInstance({
      namespace,
      path: simpleWorkflowName,
    })
  ).data.id;

  /* register mock error response */
  await page.route(
    `/api/v2/namespaces/${namespace}/logs?instance=${instanceId}`,
    async (route) => {
      if (route.request().method() === "GET") {
        const json = {
          error: { code: 422, message: "oh no!" },
        };
        await route.fulfill({ status: 422, json });
      } else route.continue();
    }
  );

  /* perform test */
  await page.goto(`/n/${namespace}/instances/${instanceId}`);

  await expect(
    page.getByText("The API returned an unexpected error: oh no!")
  ).toBeVisible();
});
