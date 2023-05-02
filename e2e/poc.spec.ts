import { expect, test } from "@playwright/test";

test("it renders the pages", async ({ page }) => {
  // mocking endpoints
  await page.route("http://localhost:3000/api/namespaces", async (route) => {
    const json = {
      pageInfo: {
        order: [],
        filter: [],
        limit: 0,
        offset: 0,
        total: 2,
      },
      results: [
        {
          createdAt: "2023-04-28T06:26:38.163628Z",
          updatedAt: "2023-04-28T06:26:38.163629Z",
          name: "one",
          oid: "",
        },
        {
          createdAt: "2023-04-28T13:50:19.561625Z",
          updatedAt: "2023-04-28T13:50:19.561626Z",
          name: "two",
          oid: "",
        },
      ],
    };
    await route.fulfill({ json });
  });

  // visit page, expect to be redirected to first namespace
  await page.goto("http://localhost:3000/");
  await expect(page).toHaveURL("http://localhost:3000/one/explorer");
  await expect(page).toHaveTitle("direktiv.io");
  await expect(page.getByRole("link", { name: "one" })).toBeVisible();

  // click on monitoring, expect to be redirected
  await page.getByRole("link", { name: "Monitoring" }).first().click();
  await expect(page).toHaveURL("http://localhost:3000/one/monitoring");
  await expect(page.getByRole("main").getByText("Monitoring")).toBeVisible();
});
