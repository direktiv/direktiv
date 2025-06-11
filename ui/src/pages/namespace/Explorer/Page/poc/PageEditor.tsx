import { ComponentProps, useState } from "react";
import { DirektivPagesSchema, DirektivPagesType } from "./schema";
import { jsonToYaml, yamlToJsonOrNull } from "../../utils";

import Button from "~/design/Button";
import { Card } from "~/design/Card";
import Editor from "~/design/Editor";
import { PageCompiler } from "./PageCompiler";
import { Save } from "lucide-react";
import { Switch } from "~/design/Switch";
import { twMergeClsx } from "~/util/helpers";
import { useTheme } from "~/util/store/theme";
import { useTranslation } from "react-i18next";

type Mode = ComponentProps<typeof PageCompiler>["mode"];

type PageEditorProps = {
  isPending: boolean;
  page: DirektivPagesType;
  onSave: (page: DirektivPagesType) => void;
};

const PageEditor = ({ isPending, page: pageProp, onSave }: PageEditorProps) => {
  const theme = useTheme();
  const [mode, setMode] = useState<Mode>("edit");
  const [page, setPage] = useState(pageProp);
  const [validate, setValidate] = useState(true);
  const [showCode, setShowCode] = useState(false);
  const { t } = useTranslation();

  return (
    <div className="relative flex grow flex-col space-y-4 p-5">
      <div
        className={twMergeClsx(
          "relative grid grow gap-5",
          showCode && "grid-cols-2"
        )}
      >
        {showCode && (
          <Card className="p-4">
            <Editor
              value={jsonToYaml(page)}
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
        <Card className="flex flex-col gap-4 p-4">
          <PageCompiler
            mode={mode}
            page={page}
            setPage={(page) => setPage(page)}
          />
        </Card>
      </div>
      <div className="flex flex-col justify-between gap-4 sm:flex-row sm:items-center">
        <div className="flex gap-5 text-sm">
          <div className="flex items-center gap-2">
            <Switch
              id="mode"
              checked={mode === "edit"}
              onCheckedChange={(value) => {
                setMode(value ? "edit" : "live");
              }}
            />
            <label htmlFor="mode">Editor</label>
          </div>
          <div className="flex items-center gap-2">
            <Switch
              id="show-code"
              checked={showCode}
              onCheckedChange={(value) => {
                setShowCode(value);
              }}
            />
            <label htmlFor="show-code">Show Code</label>
          </div>
          <div className="flex items-center gap-2">
            <Switch
              disabled={!showCode}
              id="validate"
              checked={validate}
              onCheckedChange={(value) => {
                setValidate(value);
              }}
            />
            <label htmlFor="validate">Validate</label>
          </div>
        </div>
        <Button
          variant="outline"
          type="button"
          disabled={isPending}
          onClick={() => onSave(page)}
          data-testid="page-editor-btn-save"
        >
          <Save />
          {t("pages.explorer.workflow.editor.saveBtn")}
        </Button>
      </div>
    </div>
  );
};

export default PageEditor;
