import { ComponentProps, useState } from "react";
import { DirektivPagesSchema, DirektivPagesType } from "./schema";
import { jsonToYaml, yamlToJsonOrNull } from "../../utils";

import { Card } from "~/design/Card";
import Editor from "~/design/Editor";
import { PageCompiler } from "./PageCompiler";
import { Switch } from "~/design/Switch";
import { twMergeClsx } from "~/util/helpers";
import { useTheme } from "~/util/store/theme";

const examplePage: DirektivPagesType = {
  direktiv_api: "pages/v1",
  blocks: [
    {
      type: "columns",
      blocks: [
        {
          type: "column",
          blocks: [{ type: "text", content: "column 1 text" }],
        },
        {
          type: "column",
          blocks: [{ type: "text", content: "column 2 text" }],
        },
      ],
    },
    {
      type: "card",
      blocks: [
        {
          type: "card",
          blocks: [{ type: "text", content: "text block in 2 cards" }],
        },
      ],
    },
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
          type: "headline",
          level: "h3",
          label: "Found {{query.company-list.total}} companies",
        },
        {
          type: "loop",
          id: "company",
          data: "query.company-list.data",
          blocks: [
            {
              type: "card",
              blocks: [
                {
                  type: "text",
                  content:
                    "Company {{loop.company.id}} of {{query.company-list.total}}: {{loop.company.name}}",
                },
                {
                  type: "dialog",
                  trigger: {
                    type: "button",
                    label: "show address",
                  },
                  blocks: [
                    {
                      type: "text",
                      content: "{{loop.company.addresses.0.street}}",
                    },
                    {
                      type: "text",
                      content: "{{loop.company.addresses.0.city}}",
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
  const [mode, setMode] = useState<Mode>("edit");
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
            checked={mode === "edit"}
            onCheckedChange={(value) => {
              setMode(value ? "edit" : "live");
            }}
          />
          <label htmlFor="mode">Editor</label>
        </div>
        <div className="flex gap-2 items-center">
          <Switch
            id="show-editor"
            checked={showEditor}
            onCheckedChange={(value) => {
              setShowEditor(value);
            }}
          />
          <label htmlFor="show-editor">Show Code</label>
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
        <PageCompiler mode={mode} page={page} setPage={setPage} />
      </Card>
    </div>
  );
};

export default PageEditor;
