import { FC, useState } from "react";
import { Trans, useTranslation } from "react-i18next";
import { decode, encode } from "js-base64";
import { jsonToYaml, yamlToJsonOrNull } from "../../utils";
import {
  useSetUnsavedChanges,
  useUnsavedChanges,
} from "../../Workflow/store/unsavedChangesContext";

import Button from "~/design/Button";
import { CodeEditor } from "../../Workflow/Edit/CodeEditor";
import { FileSchemaType } from "~/api/files/schema";
import { OpenapiSpecificationFormSchema } from "./schema";
import { Save } from "lucide-react";
import { UnsavedChangesHint } from "~/components/NavigationBlocker";
import { useToast } from "~/design/Toast";
import { useUpdateFile } from "~/api/files/mutate/updateFile";

type OpenapiSpecificationEditorProps = {
  data: NonNullable<FileSchemaType>;
};

const OpenapiSpecificationEditor: FC<OpenapiSpecificationEditorProps> = ({
  data,
}) => {
  const { t } = useTranslation();
  const { toast: showToast } = useToast();

  const decodedFileContentFromServer = decode(data.data);
  const hasUnsavedChanges = useUnsavedChanges();
  const setHasUnsavedChanges = useSetUnsavedChanges();

  const [editorContent, setEditorContent] = useState(
    decodedFileContentFromServer
  );

  const { mutate: updateFile, isPending } = useUpdateFile({
    onError: (errorMessage) => {
      showToast({
        variant: "error",
        title: t("pages.explorer.tree.openapiSpecification.saveError"),
        description:
          errorMessage ||
          t("pages.explorer.tree.openapiSpecification.unknownError"),
      });
    },
    onSuccess: () => {
      setHasUnsavedChanges(false);

      showToast({
        variant: "success",
        title: t("pages.explorer.tree.openapiSpecification.saveSuccessTitle"),
        description: t("pages.explorer.tree.openapiSpecification.saveSuccess"),
      });
    },
  });

  const handleFormSubmit = (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    saveContent(editorContent);
  };

  const saveContent = (content: string | undefined) => {
    const parsedContent = yamlToJsonOrNull(content ?? "");
    const result = OpenapiSpecificationFormSchema.safeParse(parsedContent);
    if (!result.success) {
      showToast({
        variant: "error",
        title: t("pages.explorer.tree.openapiSpecification.errorTitle"),
        description: (
          <Trans i18nKey="pages.explorer.tree.openapiSpecification.errorDescription" />
        ),
      });
      return;
    }
    try {
      updateFile({
        path: data.path,
        payload: { data: encode(jsonToYaml(parsedContent ?? {})) },
      });
    } catch (err) {
      showToast({
        variant: "error",
        title: t("pages.explorer.tree.openapiSpecification.saveFailed"),
        description: String(err),
      });
    }
  };

  const handleEditorChange = (value: string) => {
    setEditorContent(value);
    setHasUnsavedChanges(true);
  };

  return (
    <form
      onSubmit={handleFormSubmit}
      className="relative size-full flex-col gap-4 p-5"
    >
      <div className="flex size-full flex-col gap-4">
        <div className="grid size-full grow">
          <div className="flex size-full grow flex-col gap-5">
            <CodeEditor
              value={editorContent}
              onValueChange={handleEditorChange}
              onSave={saveContent}
              updatedAt={data.updatedAt}
            />
            <div className="flex flex-col justify-end gap-4 sm:flex-row sm:items-center">
              {hasUnsavedChanges && <UnsavedChangesHint />}
              <Button
                variant={hasUnsavedChanges ? "primary" : "outline"}
                disabled={isPending}
                type="submit"
              >
                <Save />
                {t("pages.explorer.consumer.editor.saveBtn")}
              </Button>
            </div>
          </div>
        </div>
      </div>
    </form>
  );
};

export default OpenapiSpecificationEditor;
