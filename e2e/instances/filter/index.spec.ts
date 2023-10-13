import { createNamespace, deleteNamespace } from "../../utils/namespace";
import {
  parentWorkflow as parentWorkflowContent,
  simpleWorkflow as simpleWorkflowContent,
  workflowThatFails as workflowThatFailsContent,
} from "./utils";

import { createWorkflow } from "~/api/tree/mutate/createWorkflow";
import { faker } from "@faker-js/faker";
import { runWorkflow } from "~/api/tree/mutate/runWorkflow";
import { test } from "@playwright/test";

type Instance = Awaited<ReturnType<typeof runWorkflow>>;

let namespace = "";
const simpleWorkflow = faker.system.commonFileName("yaml");
const workflowThatFails = faker.system.commonFileName("yaml");
const totalCount = 35;

test.beforeAll(async () => {
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

  const parentWorkflow = faker.system.commonFileName("yaml");

  await createWorkflow({
    payload: parentWorkflowContent({
      childName: simpleWorkflow,
      children: totalCount - 1,
    }),
    urlParams: {
      baseUrl: process.env.VITE_DEV_API_DOMAIN,
      namespace,
      name: parentWorkflow,
    },
  });
});

test.afterAll(async () => {
  await deleteNamespace(namespace);
  namespace = "";
});

test("it displays all instances without filter.", async ({ page }) => {
  await page.goto(`${namespace}/instances/`);
});
