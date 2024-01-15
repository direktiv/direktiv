import {
  DialogClose,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "~/design/Dialog";
import Editor, { EditorLanguagesType } from "~/design/Editor";
import MimeTypeSelect, {
  EditorMimeTypeSchema,
  MimeTypeType,
  TextMimeTypeType,
  getLanguageFromMimeType,
  mimeTypeToLanguageDict,
} from "~/pages/namespace/Settings/Variables/MimeTypeSelect";
import { SubmitHandler, useForm } from "react-hook-form";
import { Trans, useTranslation } from "react-i18next";
import {
  WorkflowVariableFormSchema,
  WorkflowVariableFormSchemaType,
  WorkflowVariableSchemaType,
} from "~/api/tree/schema/workflowVariable";
import { useEffect, useState } from "react";

import Alert from "~/design/Alert";
import { Braces } from "lucide-react";
import Button from "~/design/Button";
import { Card } from "~/design/Card";
import FormErrors from "~/components/FormErrors";
import Input from "~/design/Input";
import { useSetWorkflowVariable } from "~/api/tree/mutate/setVariable";
import { useTheme } from "~/util/store/theme";
import { useWorkflowVariableContent } from "~/api/tree/query/variableContent";
import { zodResolver } from "@hookform/resolvers/zod";

type EditProps = {
  item: WorkflowVariableSchemaType;
  path: string;
  onSuccess: () => void;
};

// mimeType should always be initialized in the form, to avoid the backend
// setting defaults that may not fit with the options in MimeTypeSelect
const fallbackMimeType: TextMimeTypeType = "text/plain";

const Edit = ({ item, onSuccess, path }: EditProps) => {
  const { t } = useTranslation();
  const theme = useTheme();

  const {
    data,
    isSuccess: isInitialized,
    isError,
  } = useWorkflowVariableContent(item.name, path);

  const [body, setBody] = useState<string | File>("");
  const [mimeType, setMimeType] = useState<MimeTypeType>(fallbackMimeType);

  /**
   * when the initial loaded content is from a non text mime type
   * we can't edit or save it, because we have applied res.text()
   * to the response body, which means we can't save it back to its
   * original format. Saving would not make much sense anyway, since
   * nothing would be changed.
   */
  const [saveable, setSaveable] = useState(true);

  const [editorLanguage, setEditorLanguage] = useState<EditorLanguagesType>(
    mimeTypeToLanguageDict[fallbackMimeType]
  );

  const {
    handleSubmit,
    setValue,
    formState: { errors },
  } = useForm<WorkflowVariableFormSchemaType>({
    resolver: zodResolver(WorkflowVariableFormSchema),
    values: {
      name: item.name,
      path,
      content: body ?? "",
      mimeType,
    },
  });

  const onMimeTypeChange = (value: MimeTypeType) => {
    setMimeType(value);
    setSaveable(true);
    const editorLanguage = getLanguageFromMimeType(value);
    if (editorLanguage) {
      setEditorLanguage(editorLanguage);
    }
  };

  useEffect(() => {
    if (isInitialized) {
      const contentType = data.headers["content-type"];
      const safeParsedContentType = EditorMimeTypeSchema.safeParse(contentType);
      setValue("mimeType", contentType ?? "");
      onMimeTypeChange(contentType ?? "");
      if (safeParsedContentType.success) {
        setBody(data.body);
      } else {
        setSaveable(false);
      }
    }
  }, [data, isInitialized, setValue]);

  const { mutate: setVariable } = useSetWorkflowVariable({
    onSuccess,
  });

  const onSubmit: SubmitHandler<WorkflowVariableFormSchemaType> = (data) => {
    setVariable(data);
  };

  const onFilepickerChange = async (
    event: React.ChangeEvent<HTMLInputElement>
  ) => {
    const file = event.target.files?.[0];
    if (!file) return;

    const fileContent = await file.text();
    const mimeType = file?.type ?? fallbackMimeType;

    const parsedMimetype = EditorMimeTypeSchema.safeParse(mimeType);

    setValue("mimeType", mimeType, { shouldDirty: true });
    onMimeTypeChange(mimeType);

    if (parsedMimetype.success) {
      setBody(fileContent);
    } else {
      setBody(file);
    }
  };

  if (!isInitialized) return null;

  const showEditor = saveable && typeof body === "string";

  return (
    <DialogContent>
      {isError ? (
        <>
          <Alert variant="error">
            {t("pages.settings.variables.edit.fetchError")}
          </Alert>
          <DialogFooter>
            <DialogClose asChild>
              <Button variant="ghost" data-testid="var-edit-cancel">
                {t("components.button.label.cancel")}
              </Button>
            </DialogClose>
          </DialogFooter>
        </>
      ) : (
        <form
          id="edit-variable"
          onSubmit={handleSubmit(onSubmit)}
          className="flex flex-col space-y-5"
          data-testid="wf-form-edit-variable"
        >
          <DialogHeader>
            <DialogTitle>
              <Braces />
              <Trans
                i18nKey="pages.settings.variables.edit.title"
                values={{ name: item.name }}
              />
            </DialogTitle>
          </DialogHeader>

          <FormErrors errors={errors} className="mb-5" />

          <fieldset className="flex items-center gap-5">
            <label className="w-[150px] text-right" htmlFor="mimetype">
              {t("pages.settings.variables.edit.mimeType.label")}
            </label>
            <MimeTypeSelect
              id="mimetype"
              loading={!isInitialized}
              mimeType={mimeType}
              onChange={setMimeType}
            />
          </fieldset>

          <fieldset className="flex items-center gap-5">
            <label className="w-[150px] text-right" htmlFor="file-upload">
              {t("pages.settings.variables.edit.file.label")}
            </label>
            <Input id="file-upload" type="file" onChange={onFilepickerChange} />
          </fieldset>

          <Card
            className="grow p-4 pl-0"
            background="weight-1"
            data-testid="variable-editor-card"
          >
            <div className="flex h-[400px]">
              {showEditor ? (
                <Editor
                  value={body}
                  onChange={(newData) => {
                    if (newData) {
                      setBody(newData);
                    }
                  }}
                  onMount={(editor) => editor.focus()}
                  theme={theme ?? undefined}
                  data-testid="variable-editor"
                  language={editorLanguage}
                />
              ) : (
                <div className="flex grow p-10 text-center">
                  <div className="flex items-center justify-center text-sm">
                    {t("pages.settings.variables.edit.noPreview")}
                  </div>
                </div>
              )}
            </div>
          </Card>

          <DialogFooter>
            <DialogClose asChild>
              <Button variant="ghost" data-testid="var-edit-cancel">
                {t("components.button.label.cancel")}
              </Button>
            </DialogClose>
            <Button type="submit" disabled={!saveable}>
              {t("components.button.label.save")}
            </Button>
          </DialogFooter>
        </form>
      )}
    </DialogContent>
  );
};

export default Edit;
