import { DirektivPagesType } from "../..";

export default {
  direktiv_api: "page/v1",
  type: "page",
  blocks: [
    {
      type: "headline",
      level: "h1",
      label: "Welcome to Direktiv",
    },
    {
      type: "text",
      content:
        "This is a block that contains longer text. You might write some Terms and Conditions here or something similar",
    },
    {
      type: "columns",
      blocks: [
        {
          type: "column",
          blocks: [{ type: "text", content: "first column text" }],
        },
        {
          type: "column",
          blocks: [{ type: "text", content: "second column text" }],
        },
      ],
    },
  ],
} satisfies DirektivPagesType;
