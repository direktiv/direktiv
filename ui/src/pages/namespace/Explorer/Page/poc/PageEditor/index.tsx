import {
  NavigationBlocker,
  UnsavedChangesHint,
} from "~/components/NavigationBlocker";
import { useMemo, useState } from "react";

import Button from "~/design/Button";
import { Card } from "~/design/Card";
import { DirektivPagesType } from "../schema";
import Editor from "~/design/Editor";
import EditorModeSwitcher from "./EditorModeSwitcher";
import { PageCompiler } from "../PageCompiler";
import { PageCompilerMode } from "../PageCompiler/context/pageCompilerContext";
import { Save } from "lucide-react";
import { jsonToYaml } from "../../../utils";
import { useTheme } from "~/util/store/theme";
import { useTranslation } from "react-i18next";

type PageEditorProps = {
  isPending: boolean;
  page: DirektivPagesType;
  onSave: (page: DirektivPagesType) => void;
};

export type PageEditorMode = PageCompilerMode | "code";

const PageEditor = ({
  isPending,
  page: chachedPage,
  onSave,
}: PageEditorProps) => {
  const { t } = useTranslation();
  const theme = useTheme();
  const [page, setPage] = useState(chachedPage);
  const [mode, setMode] = useState<PageEditorMode>("edit");

  const isDirty = useMemo(
    () => JSON.stringify(page) !== JSON.stringify(chachedPage),
    [page, chachedPage]
  );

  const disableSaveBtn = isPending || !isDirty;

  return (
    <div className="relative flex grow flex-col space-y-4 p-5">
      {isDirty && <NavigationBlocker />}
      <Card className="flex grow">
        {mode === "code" ? (
          <Editor
            value={jsonToYaml(page)}
            options={{ readOnly: true }}
            theme={theme ?? undefined}
            className="p-5"
          />
        ) : (
          <PageCompiler mode={mode} page={page} setPage={setPage} />
        )}
      </Card>
      <div className="flex flex-col justify-end gap-4 sm:flex-row sm:items-center">
        {isDirty && <UnsavedChangesHint />}
        <EditorModeSwitcher value={mode} onChange={setMode} />
        <Button
          variant="primary"
          type="button"
          disabled={disableSaveBtn}
          onClick={() => {
            onSave(page);
          }}
        >
          <Save />
          {t("direktivPage.blockEditor.generic.saveButton")}
        </Button>
      </div>
    </div>
  );
};

export default PageEditor;
