import { Page, expect, test } from "@playwright/test";
import {
    createNamespace,
    deleteNamespace,
} from "../../utils/namespace";

import {
    createWorkflow,
} from "../../utils/node";
import { faker } from "@faker-js/faker";

let namespace = "";
let workflow = "";
const defaultDescription = 'A simple \'no-op\' state that returns \'Hello world!\''

test.beforeEach(async ({ page }) => {
    namespace = await createNamespace();
    workflow = await createWorkflow(namespace, faker.git.shortSha() + '.yaml');
    test.setTimeout(120000)
});

test.afterEach(async () => {
    await deleteNamespace(namespace);
    namespace = "";
});

const navigateToRevisions = async (page: Page) => {
    await page.goto(`${namespace}/explorer/workflow/revisions/${workflow}`)
}

const navigateToWorkflowEditor = async (page: Page) => {
    await page.goto(`${namespace}/explorer/workflow/active/${workflow}`)
}

const testEditAndSaveWorkflow = async (page: Page, updatedText?: string): Promise<[string, string]> => {
    const description = page.getByText(updatedText ? updatedText : defaultDescription);
    await description.click()
    // type random text in that textarea which is for description
    const testText = faker.random.alphaNumeric(9);
    await page.type("textarea", testText);

    const textArea = page.getByRole("textbox");

    //now click on Save
    const saveButton = page.getByTestId("workflow-editor-btn-save");
    await saveButton.click();

    // save button should be disabled/enabled while/after the api call
    await expect(saveButton, 'save button should be disabled during the api call').toBeDisabled();
    await expect(saveButton, 'save button should be enabled after the api call').toBeEnabled();

    // after saving is completed screen should have those new changed text before/after the page reload
    await expect(page.getByText(testText), 'after saving, screen should have the updated text').toBeVisible();
    await page.reload({ waitUntil: 'load' });
    await expect(page.getByText(testText), 'after reloading, screen should have the updated text').toBeVisible();

    // check the text at the bottom left
    await expect(page.getByTestId("workflow-txt-updated"), "text should be Updated a few seconds").toHaveText("Updated a few seconds");
    const updatedWorkflow = await textArea.inputValue();
    return [updatedWorkflow, testText];
}

const actionMakeRevision = async (page: Page) => {
    const revisionTrigger = page.getByTestId("workflow-edit-trg-revision");
    await revisionTrigger.click();
    const makeRevisionButton = page.getByTestId("workflow-editor-btn-make-revision");
    await makeRevisionButton.click();
}

const actionRevertRevision = async (page: Page) => {
    const revisionTrigger = page.getByTestId("workflow-edit-trg-revision");
    await revisionTrigger.click();
    const revertRevisionButton = page.getByTestId("workflow-editor-btn-revert-revision");
    await revertRevisionButton.click();
}

const actionWaitForSuccessToast = async (page: Page) => {
    const successToast = page.getByTestId("toast-success");
    await expect(successToast, 'success toast should appear after revert action button click').toBeVisible();
    await page.getByTestId("toast-close").click();
    await expect(successToast, 'success toast should disappear after click toast-close').toBeHidden();
}

const testMakeRevision = async (page: Page) => {
    await actionMakeRevision(page);
    // check the success toast action button
    const toastAction = page.getByTestId("make-revision-toast-success-action");
    await expect(toastAction, 'success toast should appear after make-revision button click').toBeVisible();

    await toastAction.click();
    await expect(toastAction, 'success toast should disappear after toast action click').toBeHidden();
    await expect(page, "page url should have revision param").toHaveURL(/revision=/);
    const revisionId = page.url().split('revision=').pop()!;
    await expect(page.getByText(revisionId), "revisionId should be in the revision list").toBeVisible();
    //go back to the workflow editor
    await page.getByTestId('workflow-tabs-trg-activeRevision').click()
}

test("it is possible to navigate to the revisions tab", async ({ page }) => {
    await page.goto("/");
    await expect(
        page.getByTestId("breadcrumb-namespace"),
        "it renders the breadcrumb for a namespace"
    ).toBeVisible();
    // at this point, any namespace may be loaded.
    // let's navigate to the test's namespace via breadcrumbs.
    await page.getByTestId("dropdown-trg-namespace").click();
    await page
        .getByRole("option", {
            name: namespace,
        })
        .click();
    await expect(page, "the namespace is reflected in the url").toHaveURL(
        `/${namespace}/explorer/tree`
    );
    await expect(
        page.getByTestId("breadcrumb-namespace"),
        "the namespace is reflected in the breadcrumbs"
    ).toHaveText(namespace);

    await page.getByTestId(`explorer-item-link-${workflow}`).click();
    const revisionsTab = page.getByTestId('workflow-tabs-trg-revisions');
    await expect(revisionsTab, 'screen should have activeRevision tab').toBeVisible();
    await revisionsTab.click();
    await expect(page, "the workflow is reflected in the url").toHaveURL(
        `${namespace}/explorer/workflow/revisions/${workflow}`
    );
});

test("latest is the only revision by default", async ({ page }) => {
    await navigateToRevisions(page);
    const revisions = page.getByTestId(/workflow-revisions-link-item-/);
    await expect(revisions, "revisions should have the name latest").toHaveText("latest")
    await expect(revisions, "number of revisions should be one").toHaveCount(1);
});

test("it is possible to revert to the previous the workflow", async ({
    page,
}) => {
    await navigateToWorkflowEditor(page);

    const [firstUpdatedWorkflow, firstUpdatedText] = await testEditAndSaveWorkflow(page);
    await actionMakeRevision(page);

    const [secondUpdateWorkflow, secondUpdateText] = await testEditAndSaveWorkflow(page, firstUpdatedText);
    await actionRevertRevision(page);

    // wait till the revert api to be completed and handle the success toast
    await actionWaitForSuccessToast(page);

    const textArea = page.getByRole("textbox");
    const workflowValue = await textArea.inputValue();
    expect(workflowValue, "after revert, it should be the same as the first updated workflow").toBe(firstUpdatedWorkflow);
});

test("it is possible to delete the revision", async ({
    page,
}) => {
    await navigateToWorkflowEditor(page);
    const [firstUpdatedWorkflow, firstUpdatedText] = await testEditAndSaveWorkflow(page);
    await actionMakeRevision(page);
    await actionWaitForSuccessToast(page);
    await navigateToRevisions(page);

    const firstRevision = await page.getByTestId(/workflow-revisions-link-item/).nth(1).innerText();
    const firstItemMenuTrg = page.getByTestId(`workflow-revisions-item-menu-trg-${firstRevision}`);
    await firstItemMenuTrg.click();

    await expect(
        page.getByTestId(`workflow-revisions-item-menu-content-${firstRevision}`),
        "after click menu trigger, menu content should appear"
    ).toBeVisible();

    //click on the delete button to show the Delete Dialog
    const deleteTrg = page.getByTestId(`workflow-revisions-trg-delete-dlg-${firstRevision}`);
    await deleteTrg.click();

    const deleteDialog = page.getByTestId("dialog-delete-revision");
    await expect(deleteDialog, "after click delete menu, it should show the delete confirm dialog").toBeVisible();
    const submitButton = page.getByTestId("dialog-delete-revision-btn-submit");
    await submitButton.click();

    await actionWaitForSuccessToast(page);

    //after delete success, confirm that page only has one revision that is latest
    const revisions = page.getByTestId(/workflow-revisions-link-item-/);
    await expect(revisions, "revisions should have the name latest").toHaveText("latest")
    await expect(revisions, "number of revisions should be one").toHaveCount(1);

});