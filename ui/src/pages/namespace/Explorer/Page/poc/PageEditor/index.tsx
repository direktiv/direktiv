import Button from "~/design/Button";
import { Card } from "~/design/Card";
import { DirektivPagesType } from "../schema";
import Editor from "~/design/Editor";
import EditorModeSwitcher from "./EditorModeSwitcher";
import { PageCompiler } from "../PageCompiler";
import { PageCompilerMode } from "../PageCompiler/context/pageCompilerContext";
import { Save } from "lucide-react";
import { jsonToYaml } from "../../../utils";
import { useState } from "react";
import { useTheme } from "~/util/store/theme";
import { useTranslation } from "react-i18next";

type PageEditorProps = {
  isPending: boolean;
  page: DirektivPagesType;
  onSave: (page: DirektivPagesType) => void;
};

export type PageEditorMode = PageCompilerMode | "code";

const PageEditor = ({ isPending, page: pageProp, onSave }: PageEditorProps) => {
  const theme = useTheme();
  const [page, setPage] = useState(pageProp);
  const [mode, setMode] = useState<PageEditorMode>("edit");

  const { t } = useTranslation();

  return (
    <div className="relative flex grow flex-col space-y-4 p-5">
      <Card className="grow p-5">
        {mode === "code" ? (
          <Editor
            value={jsonToYaml(page)}
            options={{ readOnly: true }}
            theme={theme ?? undefined}
          />
        ) : (
          <PageCompiler
            mode={mode}
            page={page}
            setPage={(page) => setPage(page)}
          />
        )}
      </Card>
      <div className="flex flex-col justify-end gap-4 sm:flex-row sm:items-center">
        <EditorModeSwitcher value={mode} onChange={setMode} />
        <Button
          variant="outline"
          type="button"
          disabled={isPending}
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
