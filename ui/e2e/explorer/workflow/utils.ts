import { Page, expect } from "@playwright/test";

export const waitForSuccessToast = async (page: Page) => {
  const successToast = page.getByTestId("toast-success");
  await expect(successToast, "a success toast appears").toBeVisible();
  await page.getByTestId("toast-close").click();
  await expect(
    successToast,
    "success toast disappears after clicking toast-close"
  ).toBeHidden();
};

export const testDiacriticsWorkflow = `// A workflow for testing characters like îèüñÆ.
const flow: FlowDefinition = {
  type: "default",
  timeout: "PT30S",
  state: "stateValidateInput",
};

function stateValidateInput(input) {
  if (typeof input === "object" && typeof input.name === "string") {
    return transition(stateSayHello, input)
  }
  throw new Error("invalid input")
}

function stateSayHello(input: { name: string}) {
  return finish({"result": \`Hello \${input.name}\`})
}
`;

export const workflowThatCreatesVariable = `direktiv_api: workflow/v1
states:
- id: store-workflow-var
  type: setter
  variables:
  - key: workflow
    scope: workflow
    # don't set a mime type on purpuse
    # mimeType: application/octet-stream
    value: This is my workflow variable value
`;
