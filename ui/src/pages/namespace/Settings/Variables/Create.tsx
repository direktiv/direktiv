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
} from "./MimeTypeSelect";
import { SubmitHandler, useForm } from "react-hook-form";
import {
  VarFormSchema,
  VarFormSchemaType,
} from "~/api/variables_obsolete/schema";

import Button from "~/design/Button";
import { Card } from "~/design/Card";
import FormErrors from "~/components/FormErrors";
import Input from "~/design/Input";
import { PlusCircle } from "lucide-react";
import { useState } from "react";
import { useTheme } from "~/util/store/theme";
import { useTranslation } from "react-i18next";
import { useUpdateVar } from "~/api/variables_obsolete/mutate/updateVariable";
import { zodResolver } from "@hookform/resolvers/zod";

type CreateProps = { onSuccess: () => void };

const defaultMimeType: TextMimeTypeType = "application/json";

const Create = ({ onSuccess }: CreateProps) => {
  const { t } = useTranslation();
  const theme = useTheme();

  const [name, setName] = useState("");
  const [body, setBody] = useState<string | File>("");
  const [mimeType, setMimeType] = useState<MimeTypeType>(defaultMimeType);
  const [editorLanguage, setEditorLanguage] = useState<EditorLanguagesType>(
    mimeTypeToLanguageDict[defaultMimeType]
  );

  const {
    handleSubmit,
    setValue,
    formState: { errors },
  } = useForm<VarFormSchemaType>({
    resolver: zodResolver(VarFormSchema),
    // mimeType should always be initialized to avoid backend defaulting to
    // "text/plain, charset=utf-8", which does not fit the options in
    // MimeTypeSelect
    values: {
      name,
      content: body,
      mimeType,
    },
  });

  const onMimeTypeChange = (value: MimeTypeType) => {
    setMimeType(value);
    const editorLanguage = getLanguageFromMimeType(value);
    if (editorLanguage) {
      setEditorLanguage(editorLanguage);
    }
  };

  const { mutate: createVarMutation } = useUpdateVar({
    onSuccess,
  });

  const onSubmit: SubmitHandler<VarFormSchemaType> = (data) => {
    createVarMutation(data);
  };

  const onFilepickerChange = async (
    event: React.ChangeEvent<HTMLInputElement>
  ) => {
    const file = event.target.files?.[0];
    if (!file) return;

    const fileContent = await file.text();
    const mimeType = file?.type ?? defaultMimeType;
    const parsedMimetype = EditorMimeTypeSchema.safeParse(mimeType);

    setValue("mimeType", mimeType);
    onMimeTypeChange(mimeType);
    if (parsedMimetype.success) {
      setBody(fileContent);
    } else {
      setBody(file);
    }
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
              {t("pages.settings.variables.create.title")}
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
            onChange={onMimeTypeChange}
          />
        </fieldset>

        <fieldset className="flex items-center gap-5">
          <label className="w-[150px] text-right" htmlFor="file-upload">
            {t("pages.settings.variables.create.file.label")}
          </label>
          <Input id="file-upload" type="file" onChange={onFilepickerChange} />
        </fieldset>

        <Card
          className="grow p-4 pl-0"
          background="weight-1"
          data-testid="variable-create-card"
        >
          <div className="flex h-[400px]">
            {typeof body === "string" ? (
              <Editor
                value={body}
                onChange={(newData) => {
                  if (newData) {
                    setBody(newData);
                  }
                }}
                theme={theme ?? undefined}
                data-testid="variable-editor"
                language={editorLanguage}
              />
            ) : (
              <div className="flex grow p-10 text-center">
                <div className="flex items-center justify-center text-sm">
                  {t("pages.settings.variables.create.noPreview")}
                </div>
              </div>
            )}
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
