import { ComponentProps, useState } from "react";
import { DirektivPagesSchema, DirektivPagesType } from "./schema";
import { jsonToYaml, yamlToJsonOrNull } from "../../utils";

import { Card } from "~/design/Card";
import Editor from "~/design/Editor";
import { PageCompiler } from "./PageCompiler";
import { Switch } from "~/design/Switch";
import { twMergeClsx } from "~/util/helpers";
import { useTheme } from "~/util/store/theme";

const examplePage = {
  direktiv_api: "pages/v1",
  blocks: [
    {
      type: "headline",
      label: "Welcome to Direktiv",
      size: "h3",
    },
    {
      type: "text",
      label:
        "This is a block that contains longer text. You might write some Terms and Conditions here or something similar",
    },
    {
      type: "card",
      blocks: [
        {
          type: "query-provider",
          queries: [
            {
              id: "company-list",
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
              type: "two-columns",
              leftBlocks: [
                {
                  type: "text",
                  label:
                    "I can access text from a query: {{query.company-list.data.5.addresses.0.streetName}}. I can also handle some edge caes like {{query.company-list.data.5.addresses}} {{query.company-list.i.made.this.up}}",
                },
                {
                  type: "loop",
                  id: "company",
                  variable: "query.company-list.data.name",
                  blocks: [
                    {
                      type: "text",
                      label: "I am a loop, but I don't work yet",
                    },
                  ],
                },
              ],
              rightBlocks: [
                {
                  type: "text",
                  label: "I am the right column",
                },
                {
                  type: "dialog",
                  trigger: {
                    type: "button",
                    label: "open dialog",
                  },
                  blocks: [
                    {
                      type: "headline",
                      label: "Hello",
                      size: "h3",
                    },
                    {
                      type: "text",
                      label:
                        "This modal will only fetch data when opened. Slow down your network to see a loading spinner (this query will fail intentionally)",
                    },
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
                    {
                      type: "text",
                      label: "This component is not implemented yet",
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
    },
  ],
} satisfies DirektivPagesType;

type Mode = ComponentProps<typeof PageCompiler>["mode"];

const PageEditor = () => {
  const theme = useTheme();
  const [mode, setMode] = useState<Mode>("live");
  const [page, setPage] = useState(examplePage);
  const [validate, setValidate] = useState(true);
  const [showEditor, setShowEditor] = useState(false);

  return (
    <div
      className={twMergeClsx(
        "grid gap-5 grow p-5 relative",
        showEditor && "grid-cols-2"
      )}
    >
      <div className="right-5 -top-12 absolute gap-5 flex text-sm">
        <div className="flex gap-2 items-center">
          <Switch
            id="mode"
            checked={mode === "preview"}
            onCheckedChange={(value) => {
              setMode(value ? "preview" : "live");
            }}
          />
          <label htmlFor="mode">Preview</label>
        </div>
        <div className="flex gap-2 items-center">
          <Switch
            id="show-editor"
            checked={showEditor}
            onCheckedChange={(value) => {
              setShowEditor(value);
            }}
          />
          <label htmlFor="show-editor">Show Editor</label>
        </div>
        <div className="flex gap-2 items-center">
          <Switch
            disabled={!showEditor}
            id="validate"
            checked={validate}
            onCheckedChange={(value) => {
              setValidate(value);
            }}
          />
          <label htmlFor="validate">Validate</label>
        </div>
      </div>
      {showEditor && (
        <Card className="p-4">
          <Editor
            value={jsonToYaml(examplePage)}
            theme={theme ?? undefined}
            onChange={(newValue) => {
              if (newValue) {
                const newValueJson = yamlToJsonOrNull(newValue);
                if (
                  validate &&
                  !DirektivPagesSchema.safeParse(newValueJson).success
                ) {
                  return;
                }
                setPage(newValueJson);
              }
            }}
          />
        </Card>
      )}
      <Card className="p-4 flex flex-col gap-4">
        <PageCompiler page={page} mode={mode} />
      </Card>
    </div>
  );
};

export default PageEditor;
