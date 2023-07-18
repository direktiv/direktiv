import { createNamespace, deleteNamespace } from "../../utils/namespace";
import { expect, test } from "@playwright/test";
import {
  simpleWorkflow as simpleWorkflowContent,
  workflowThatFails as workflowThatFailsContent,
} from "./utils";

import { createWorkflow } from "~/api/tree/mutate/createWorkflow";
import { faker } from "@faker-js/faker";
import { runWorkflow } from "~/api/tree/mutate/runWorkflow";

let namespace = "";
const simpleWorkflow = faker.system.commonFileName("yaml");
const workflowThatFails = faker.system.commonFileName("yaml");

test.beforeEach(async () => {
  namespace = await createNamespace();
  // place some workflows in the namespace that we can use to create instances
  await createWorkflow({
    payload: simpleWorkflowContent,
    urlParams: {
      baseUrl: process.env.VITE_DEV_API_DOMAIN,
      namespace,
      name: simpleWorkflow,
    },
  });

  await createWorkflow({
    payload: workflowThatFailsContent,
    urlParams: {
      baseUrl: process.env.VITE_DEV_API_DOMAIN,
      namespace,
      name: workflowThatFails,
    },
  });
});

test.afterEach(async () => {
  await deleteNamespace(namespace);
  namespace = "";
});

const createBasicInstance = async () => {
  await runWorkflow({
    urlParams: {
      baseUrl: process.env.VITE_DEV_API_DOMAIN,
      namespace,
      path: simpleWorkflow,
    },
  });
};

const createFailedInstance = async () =>
  await runWorkflow({
    urlParams: {
      baseUrl: process.env.VITE_DEV_API_DOMAIN,
      namespace,
      path: workflowThatFails,
    },
  });

test("this is an example test", async ({ page }) => {
  await createFailedInstance();
  await createBasicInstance();
  await createBasicInstance();
  // { waitUntil: "networkidle" } might not be necessary, I just added
  // it so that running this example will show a nicely loaded page
  await page.goto(`${namespace}/instances/`, { waitUntil: "networkidle" });
});

test("there is no result", async ({ page }) => {
  // await createFailedInstance();
  // await createBasicInstance();
  // await createBasicInstance();
  // { waitUntil: "networkidle" } might not be necessary, I just added
  // it so that running this example will show a nicely loaded page
  await page.goto(`${namespace}/instances/`, { waitUntil: "networkidle" });
  await expect(
    page.getByTestId("instance-no-result"),
    "no result message should be visible"
  ).toBeVisible();
  await expect(
    page.getByTestId("instance-list-pagination"),
    "there is no pagination when there is no result"
  ).not.toBeVisible();
});

test("test through the instance list screen without pagination", async ({
  page,
}) => {
  const failedInstance = await createFailedInstance();
  const successInstance1 = await createBasicInstance();
  const successInstance2 = await createBasicInstance();
  // { waitUntil: "networkidle" } might not be necessary, I just added
  // it so that running this example will show a nicely loaded page
  await page.goto(`${namespace}/instances/`, { waitUntil: "networkidle" });
  const failedItem = page.getByTestId(
    `instance-item-${failedInstance.instance}`
  );
  await expect(failedItem, "failed Item should have the id").toContainText(
    failedInstance.instance
  );
});
