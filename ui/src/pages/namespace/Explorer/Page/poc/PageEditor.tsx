import { ComponentProps, useState } from "react";
import { Card } from "~/design/Card";
import { DirektivPagesType } from "./schema";
import Editor from "~/design/Editor";
import { PageCompiler } from "./PageCompiler";
import { Separator } from "~/design/Separator";
import { Switch } from "~/design/Switch";
import { jsonToYaml } from "../../utils";
import { useTheme } from "~/util/store/theme";

const examplePage = {
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
          type: "query-provider",
          queries: [
            {
              id: "fetching-resources",
              endpoint: "/ns/demo/companies",
              queryParams: [
                {
                  key: "query",
                  value: "my-search-query",
                },
              ],
            },
          ],
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
                      queries: [
                        {
                          id: "fetching-resources-2",
                          endpoint: "/api/get/resources",
                          queryParams: [
                            {
                              key: "query",
                              value: "my-search-query",
                            },
                          ],
                        },
                      ],
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

type Mode = ComponentProps<typeof PageCompiler>["mode"];

const PageEditor = () => {
  const theme = useTheme();
  const [mode, setMode] = useState<Mode>("preview");
  return (
    <div className="grid grid-cols-2 gap-5 grow p-5">
      <Card className="p-4">
        <Editor
          value={jsonToYaml(examplePage)}
          theme={theme ?? undefined}
          options={{ readOnly: true }}
        />
      </Card>
      <Card className="p-4 flex flex-col gap-4">
        <div className="flex items-center space-x-2">
          <Switch
            id="mode"
            checked={mode === "preview"}
            onCheckedChange={(value) => {
              setMode(value ? "preview" : "live");
            }}
          />

          <label htmlFor="mode">Preview</label>
        </div>
        <Separator />
        <PageCompiler page={examplePage} mode={mode} />
      </Card>
    </div>
  );
};

export default PageEditor;
