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
import { NodeSchemaType } from "~/api/tree/schema/node";
import { mimeTypeToEditorSyntax } from "~/design/Editor/utils";
import { useNodeContent } from "~/api/tree/query/node";
import { useTheme } from "~/util/store/theme";
import { useTranslation } from "react-i18next";

const NoPreview = ({ mimeType }: { mimeType?: string }) => {
  const { t } = useTranslation();

  if (!mimeType) return null; // prevent layout shift

  return (
    <div className="flex grow flex-col items-center justify-center gap-3 p-10">
      <div>{t("pages.explorer.tree.fileViewer.notSupported")}</div>
      <code className="text-sm text-primary-500">{mimeType}</code>
    </div>
  );
};

const imagesSrc = (mimeType: string, source: string) =>
  `data:${mimeType};base64,${source}`;

const FileViewer = ({ node }: { node: NodeSchemaType }) => {
  const { t } = useTranslation();
  const theme = useTheme();
  const { data } = useNodeContent({ path: node.path });

  const fileContent = atob(data?.revision?.source ?? "");
  const mimeType = data?.node.mimeType;

  const supportedLanguage = mimeTypeToEditorSyntax(mimeType);
  const supportedImage = mimeType?.startsWith("image/");

  const showEditor = supportedLanguage !== undefined;
  const showImage = !showEditor && supportedImage;

  const noPreview = !showEditor && !showImage;

  return (
    <>
      <DialogHeader>
        <DialogTitle>
          <File /> {t("pages.explorer.tree.fileViewer.title")} {node.path}
        </DialogTitle>
      </DialogHeader>
      <Card className="grow p-4" background="weight-1">
        <div className="flex h-[700px]">
          {showImage && (
            <img
              src={imagesSrc(mimeType ?? "", data?.revision?.source ?? "")}
              className="w-full object-contain"
            />
          )}

          {showEditor && (
            <Editor
              language={supportedLanguage}
              value={fileContent}
              options={{
                readOnly: true,
              }}
              theme={theme ?? undefined}
            />
          )}

          {noPreview && <NoPreview mimeType={mimeType} />}
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
