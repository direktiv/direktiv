import {
  DialogClose,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "~/design/Dialog";
import Editor, { EditorLanguagesType } from "~/design/Editor";
import MimeTypeSelect, {
  MimeTypeType,
  mimeTypeToLanguageDict,
} from "~/pages/namespace/Settings/Variables/MimeTypeSelect";
import { SubmitHandler, useForm } from "react-hook-form";
import {
  WorkflowVariableFormSchema,
  WorkflowVariableFormSchemaType,
} from "~/api/tree/schema";

import Button from "~/design/Button";
import { Card } from "~/design/Card";
import FormErrors from "~/componentsNext/FormErrors";
import Input from "~/design/Input";
import { PlusCircle } from "lucide-react";
import { useSetWorkflowVariable } from "~/api/tree/mutate/setVariable";
import { useState } from "react";
import { useTheme } from "~/util/store/theme";
import { useTranslation } from "react-i18next";
import { zodResolver } from "@hookform/resolvers/zod";

type CreateProps = { onSuccess: () => void; path: string };

const defaultMimeType: MimeTypeType = "application/json";

const Create = ({ onSuccess, path }: CreateProps) => {
  const { t } = useTranslation();
  const theme = useTheme();

  const [name, setName] = useState<string | undefined>();
  const [body, setBody] = useState<string | undefined>();
  const [mimeType, setMimeType] = useState<MimeTypeType>(defaultMimeType);
  const [editorLanguage, setEditorLanguage] = useState<EditorLanguagesType>(
    mimeTypeToLanguageDict[defaultMimeType]
  );

  const {
    handleSubmit,
    formState: { errors },
  } = useForm<WorkflowVariableFormSchemaType>({
    resolver: zodResolver(WorkflowVariableFormSchema),
    // mimeType should always be initialized to avoid backend defaulting to
    // "text/plain, charset=utf-8", which does not fit the options in
    // MimeTypeSelect
    values: {
      name: name ?? "",
      path,
      content: body ?? "",
      mimeType: mimeType ?? defaultMimeType,
    },
  });

  const onMimetypeChange = (value: MimeTypeType) => {
    setMimeType(value);
    setEditorLanguage(mimeTypeToLanguageDict[value]);
  };

  const { mutate: createVarMutation } = useSetWorkflowVariable({
    onSuccess,
  });

  const onSubmit: SubmitHandler<WorkflowVariableFormSchemaType> = (data) => {
    createVarMutation(data);
  };

  return (
    <DialogContent>
      <form
        id="create-variable"
        onSubmit={handleSubmit(onSubmit)}
        className="flex flex-col space-y-5"
      >
        <DialogHeader>
          <DialogHeader>
            <DialogTitle>
              <PlusCircle />
              {t(
                "pages.explorer.tree.workflow.settings.variables.create.title"
              )}
            </DialogTitle>
          </DialogHeader>
        </DialogHeader>

        <FormErrors errors={errors} className="mb-5" />

        <fieldset className="flex items-center gap-5">
          <label className="w-[150px] text-right" htmlFor="name">
            {t("pages.settings.variables.create.name.label")}
          </label>
          <Input
            id="name"
            data-testid="new-variable-name"
            placeholder={t("pages.settings.variables.create.name.placeholder")}
            onChange={(event) => setName(event.target.value)}
          />
        </fieldset>

        <fieldset className="flex items-center gap-5">
          <label className="w-[150px] text-right" htmlFor="mimetype">
            {t("pages.settings.variables.edit.mimeType.label")}
          </label>
          <MimeTypeSelect
            id="mimetype"
            mimeType={mimeType}
            onChange={onMimetypeChange}
          />
        </fieldset>

        <Card
          className="grow p-4 pl-0"
          background="weight-1"
          data-testid="variable-create-card"
        >
          <div className="h-[500px]">
            <Editor
              value={body}
              onChange={(newData) => {
                setBody(newData);
              }}
              theme={theme ?? undefined}
              data-testid="variable-editor"
              language={editorLanguage}
            />
          </div>
        </Card>

        <DialogFooter>
          <DialogClose asChild>
            <Button variant="ghost">
              {t("components.button.label.cancel")}
            </Button>
          </DialogClose>
          <Button data-testid="variable-create-submit" type="submit">
            {t("components.button.label.create")}
          </Button>
        </DialogFooter>
      </form>
    </DialogContent>
  );
};

export default Create;
