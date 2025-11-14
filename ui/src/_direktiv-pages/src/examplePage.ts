import { DirektivPagesType } from "~/pages/namespace/Explorer/Page/poc/schema";

export const page: DirektivPagesType = {
  direktiv_api: "page/v1",
  type: "page",
  blocks: [
    {
      type: "headline",
      level: "h1",
      label: "Direktiv Pages Demo",
    },
    {
      type: "text",
      content:
        "This is a demo page for the Direktiv Pages local development server. In the production environment, the page configuration is returned from the gateway, and the page content depends on the user's configuration and the current URL. Since the dev server runs locally and is not embedded in the Direktiv gateway, this example page simulates the behavior of the real page.",
    },
  ],
};
