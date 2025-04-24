import { Card } from "~/design/Card";
import { DirektivPagesType } from "./schema";
import Editor from "~/design/Editor";
import { PageCompiler } from "./PageCompiler";
import { jsonToYaml } from "../../utils";
import { useTheme } from "~/util/store/theme";

const examplePage = {
  direktiv_api: "pages/v1",
  path: "/som/path",
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
          type: "query-provider",
          query: {
            id: "fetching-resources",
            endpoint: "/api/get/resources",
            queryParams: [
              {
                key: "query",
                value: "my-search-query",
              },
            ],
          },
          blocks: [
            {
              type: "dialog",
              trigger: {
                type: "button",
                label: "open dialog",
              },
              blocks: [
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
                      type: "query-provider",
                      query: {
                        id: "fetching-resources",
                        endpoint: "/api/get/resources",
                        queryParams: [
                          {
                            key: "query",
                            value: "my-search-query",
                          },
                        ],
                      },
                      blocks: [],
                    },
                  ],
                },

                {
                  type: "form",
                  trigger: {
                    type: "button",
                    label: "delte",
                  },
                  mutation: {
                    id: "my-delete",
                    endpoint: "/api/delete/",
                    method: "DELETE",
                  },
                  blocks: [],
                },
              ],
            },
          ],
        },
      ],
    },
  ],
} satisfies DirektivPagesType;

const PageEditor = () => {
  const theme = useTheme();
  return (
    <div className="grid grid-cols-2 gap-5 grow p-5">
      <Card className="p-4">
        <Editor
          value={jsonToYaml(examplePage)}
          theme={theme ?? undefined}
          options={{ readOnly: true }}
        />
      </Card>
      <Card className="p-4">
        <PageCompiler page={examplePage} />
      </Card>
    </div>
  );
};

export default PageEditor;
