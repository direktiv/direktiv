import { DirektivPagesType } from "../..";

export default {
  direktiv_api: "pages/v1",
  blocks: [
    {
      type: "headline",
      size: "h1",
      label: "Welcome to Direktiv",
    },
    {
      type: "text",
      label:
        "This is a block that contains longer text. You might write some Terms and Conditions here or something similar",
    },
    {
      type: "columns",
      columns: [
        [
          {
            type: "text",
            label: "Some text goes here",
          },
        ],
        [
          {
            type: "text",
            label: "Some text goes here",
          },
        ],
      ],
    },
  ],
} satisfies DirektivPagesType;
