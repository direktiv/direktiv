import { ComponentProps, useState } from "react";
import { DirektivPagesSchema, DirektivPagesType } from "../schema";
import { jsonToYaml, yamlToJsonOrNull } from "../../../utils";

import Button from "~/design/Button";
import ButtonBar from "./ButtonBar";
import { Card } from "~/design/Card";
import Editor from "~/design/Editor";
import { PageCompiler } from "../PageCompiler";
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
      <Card className="grow p-5">
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
        {/* <PageCompiler
          mode={mode}
          page={page}
          setPage={(page) => setPage(page)}
        /> */}
      </Card>
      <div className="flex flex-col justify-end gap-4 sm:flex-row sm:items-center">
        <ButtonBar />
        <Button
          variant="outline"
          type="button"
          disabled={isPending}
          onClick={() => onSave(page)}
        >
          <Save />
          {t("direktivPage.blockEditor.generic.saveButton")}
        </Button>
      </div>
    </div>
  );
};

export default PageEditor;
