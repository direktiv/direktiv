import {
  DialogClose,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "~/design/Dialog";

import Button from "~/design/Button";
import { Card } from "~/design/Card";
import Editor from "~/design/Editor";
import { File } from "lucide-react";
import { NodeSchemaType } from "~/api/tree/schema";
import { useNodeContent } from "~/api/tree/query/node";
import { useTheme } from "~/util/store/theme";
import { useTranslation } from "react-i18next";

const FileViewer = ({ node }: { node: NodeSchemaType }) => {
  const { t } = useTranslation();
  const theme = useTheme();
  const { data } = useNodeContent({ path: node.path });
  const fileContent = atob(data?.revision?.source ?? "");

  return (
    <>
      <DialogHeader>
        <DialogTitle>
          <File /> {t("pages.explorer.tree.fileViewer.title")} {node.path}
        </DialogTitle>
      </DialogHeader>
      <Card className="grow p-4 pl-0" background="weight-1">
        <div className="h-[700px]">
          <Editor
            language="plaintext"
            value={fileContent}
            options={{
              readOnly: true,
            }}
            theme={theme ?? undefined}
          />
        </div>
      </Card>
      <DialogFooter>
        <DialogClose asChild>
          <Button>{t("pages.explorer.tree.fileViewer.closeBtn")}</Button>
        </DialogClose>
      </DialogFooter>
    </>
  );
};

export default FileViewer;
