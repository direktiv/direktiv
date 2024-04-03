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

export const jsonSchemaFormWorkflow = `description: A workflow with a complex json schema form'
states:
- id: input
  type: validate
  schema:
    title: some test
    type: object
    required:
    - firstName
    - lastName
    properties:
      firstName:
        type: string
        title: First name
      lastName:
        type: string
        title: Last name
      select:
        title: role
        type: string
        enum: 
          - admin
          - guest
      array:
        title: A list of strings
        type: array
        items:
          type: string
      age:
        type: integer
        title: Age
      file:
        type: string
        title: file upload
        format: data-url`;

export const jsonSchemaWithRequiredEnum = `description: A workflow with a complex json schema form'
states:
- id: input
  type: validate
  schema:
    title: some test
    type: object
    required:
    - firstName
    - lastName
    - select
    properties:
      firstName:
        type: string
        title: First name
      lastName:
        type: string
        title: Last name
      select:
        title: role
        type: string
        enum: 
          - admin
          - guest
      `;

export const testDiacriticsWorkflow = `direktiv_api: workflow/v1
description: A workflow for testing characters like îèüñÆ.
states:
- id: validate-input
  type: validate
  schema:
    type: object
    required:
    - name
    properties:
      name:
        type: string
        description: Name to greet
        title: Name
  transition: sayhello

- id: sayhello
  type: noop
  transform:
    result: 'Hello jq(.name)'
`;
