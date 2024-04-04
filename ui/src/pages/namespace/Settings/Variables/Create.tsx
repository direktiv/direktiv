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
  getLanguageFromMimeType,
  mimeTypeToLanguageDict,
} from "./MimeTypeSelect";
import { SubmitHandler, useForm } from "react-hook-form";
import { VarFormSchema, VarFormSchemaType } from "~/api/variables/schema";
import { decode, encode } from "js-base64";

import Button from "~/design/Button";
import { Card } from "~/design/Card";
import FileUpload from "../components/FileUpload";
import FormErrors from "~/components/FormErrors";
import Input from "~/design/Input";
import { PlusCircle } from "lucide-react";
import { useCreateVar } from "~/api/variables/mutate/createVariable";
import { useState } from "react";
import { useTheme } from "~/util/store/theme";
import { useTranslation } from "react-i18next";
import { zodResolver } from "@hookform/resolvers/zod";

type CreateProps = { onSuccess: () => void };

const defaultMimeType = "application/json";

const Create = ({ onSuccess }: CreateProps) => {
  const { t } = useTranslation();
  const theme = useTheme();

  const [editorText, setEditorText] = useState("");
  const [base64String, setBase64String] = useState("");
  const [isEditable, setIsEditable] = useState(true);
  const [mimeType, setMimeType] = useState(defaultMimeType);
  const [editorLanguage, setEditorLanguage] = useState<EditorLanguagesType>(
    mimeTypeToLanguageDict[defaultMimeType]
  );

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<VarFormSchemaType>({
    resolver: zodResolver(VarFormSchema),
    values: {
      name: "",
      data: base64String,
      mimeType,
    },
  });

  const onMimeTypeChange = (value: string) => {
    setMimeType(value);
    const editorLanguage = getLanguageFromMimeType(value);
    if (editorLanguage) {
      setEditorLanguage(editorLanguage);
    }
  };

  const { mutate: createVar } = useCreateVar({
    onSuccess,
  });

  const onSubmit: SubmitHandler<VarFormSchemaType> = (data) => {
    createVar(data);
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
            {...register("name")}
            data-testid="new-variable-name"
            placeholder={t("pages.settings.variables.create.name.placeholder")}
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
        <FileUpload
          onChange={({ base64String, mimeType }) => {
            const parsedMimetype = EditorMimeTypeSchema.safeParse(mimeType);
            const isEditable = parsedMimetype.success;
            setIsEditable(isEditable);
            if (isEditable) {
              setEditorText(decode(base64String));
            }
            onMimeTypeChange(mimeType);
            setBase64String(base64String);
          }}
        />

        <Card
          className="grow p-4 pl-0"
          background="weight-1"
          data-testid="variable-create-card"
        >
          <div className="flex h-[400px]">
            {isEditable ? (
              <Editor
                value={editorText}
                onChange={(newData) => {
                  if (newData) {
                    setEditorText(newData);
                    setBase64String(encode(newData));
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
