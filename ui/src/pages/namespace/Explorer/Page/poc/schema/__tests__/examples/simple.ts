import { DirektivPagesType } from "../..";

export default {
  direktiv_api: "pages/v1",
  blocks: [
    {
      type: "headline",
      label: "Welcome to Direktiv",
      description: "This is a headline block inside a Direktiv page",
    },
    {
      type: "text",
      label:
        "This is a block that contains longer text. You might write some Terms and Conditions here or something similar",
    },
    {
      type: "two-columns",
      leftBlocks: [
        {
          type: "text",
          label: "Some text goes here",
        },
      ],
      rightBlocks: [
        {
          type: "text",
          label: "Some text goes here",
        },
      ],
    },
  ],
} satisfies DirektivPagesType;
