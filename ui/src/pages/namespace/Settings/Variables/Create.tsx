import {
  DialogClose,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "~/design/Dialog";
import Editor, { EditorLanguagesType } from "~/design/Editor";
import { Loader2, PlusCircle } from "lucide-react";
import MimeTypeSelect, {
  EditorMimeTypeSchema,
  getLanguageFromMimeType,
  mimeTypeToLanguageDict,
} from "./MimeTypeSelect";
import { SubmitHandler, useForm } from "react-hook-form";
import { VarFormSchema, VarFormSchemaType } from "~/api/variables/schema";

import Button from "~/design/Button";
import { Card } from "~/design/Card";
import FormErrors from "~/components/FormErrors";
import Input from "~/design/Input";
import { InputWithButton } from "~/design/InputWithButton";
import { encode } from "js-base64";
import { useCreateVar } from "~/api/variables/mutate/createVariable";
import { useState } from "react";
import { useTheme } from "~/util/store/theme";
import { useTranslation } from "react-i18next";
import { zodResolver } from "@hookform/resolvers/zod";

type CreateProps = { onSuccess: () => void };

const defaultMimeType = "application/json";

const parseDataUrl = (dataUrl: string) => {
  const splitUrl = dataUrl.split(";");
  if (!splitUrl || !splitUrl[0] || !splitUrl[1]) return null;

  const mimeType = splitUrl[0].split(":")[1];
  const data = splitUrl[1].split(",")[1];

  if (!mimeType || !data) return null;
  return {
    mimeType,
    data,
  };
};

const Create = ({ onSuccess }: CreateProps) => {
  const { t } = useTranslation();
  const theme = useTheme();

  const [name, setName] = useState("");
  const [editorText, setEditorText] = useState("");
  const [base64String, setBase64String] = useState("");
  const [isEditable, setIsEditable] = useState(true);
  const [isUploading, setIsUploading] = useState(false);
  const [mimeType, setMimeType] = useState(defaultMimeType);
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

  const onFilepickerChange = async (
    event: React.ChangeEvent<HTMLInputElement>
  ) => {
    const file = event.target.files?.[0];
    if (!file) return;
    const fileReader = new FileReader();
    fileReader.onload = function (e) {
      const fileContent = e.target?.result;
      if (typeof fileContent === "string") {
        const parsedDataUrl = parseDataUrl(fileContent);
        if (parsedDataUrl) {
          setBase64String(parsedDataUrl.data);
          const mimeType = parsedDataUrl.mimeType ?? defaultMimeType;
          const parsedMimetype = EditorMimeTypeSchema.safeParse(mimeType);
          setIsEditable(parsedMimetype.success);
          setValue("mimeType", mimeType);
          onMimeTypeChange(mimeType);
        }
      }
      setIsUploading(false);
    };

    fileReader.onerror = function () {
      setIsUploading(false);
    };

    setIsUploading(true);
    fileReader.readAsDataURL(file);
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

          <InputWithButton>
            <Input id="file-upload" type="file" onChange={onFilepickerChange} />
            {isUploading && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
          </InputWithButton>
        </fieldset>
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
