import { Controller, SubmitHandler, useForm } from "react-hook-form";
import Editor, { EditorLanguagesType } from "~/design/Editor";
import {
  VarFormCreateEditSchema,
  VarFormCreateEditSchemaType,
} from "~/api/variables/schema";
import { decode, encode } from "js-base64";
import {
  getLanguageFromMimeType,
  isMimeTypeEditable,
  mimeTypeToLanguageDict,
} from "./utils";

import { Card } from "~/design/Card";
import { DialogHeader } from "~/design/Dialog";
import FileUpload from "./FileUpload";
import FormErrors from "~/components/FormErrors";
import Input from "~/design/Input";
import MimeTypeSelect from "./MimeTypeSelect";
import { useState } from "react";
import { useTheme } from "~/util/store/theme";
import { useTranslation } from "react-i18next";
import { zodResolver } from "@hookform/resolvers/zod";

const defaultMimeType = "application/json";

type VariableForm = {
  defaultValues: VarFormCreateEditSchemaType;
  dialogTitle: JSX.Element;
  dialogFooter: JSX.Element;
  onMutate: (data: VarFormCreateEditSchemaType) => void;
};

export const VariableForm = ({
  defaultValues,
  dialogTitle,
  dialogFooter,
  onMutate,
}: VariableForm) => {
  const { t } = useTranslation();
  const theme = useTheme();

  const [isEditable, setIsEditable] = useState(
    isMimeTypeEditable(defaultValues.mimeType)
  );

  const [editorLanguage, setEditorLanguage] = useState<EditorLanguagesType>(
    mimeTypeToLanguageDict[
      isMimeTypeEditable(defaultValues.mimeType)
        ? defaultValues.mimeType
        : defaultMimeType
    ]
  );

  const {
    register,
    handleSubmit,
    control,
    setValue,
    formState: { errors },
  } = useForm<VarFormCreateEditSchemaType>({
    resolver: zodResolver(VarFormCreateEditSchema),
    defaultValues,
  });

  const onMimeTypeChange = (value: string) => {
    setValue("mimeType", value);
    const editorLanguage = getLanguageFromMimeType(value);
    if (editorLanguage) {
      setEditorLanguage(editorLanguage);
    }
  };

  const onSubmit: SubmitHandler<VarFormCreateEditSchemaType> = (data) => {
    onMutate(data);
  };

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="flex flex-col space-y-5">
      <DialogHeader>
        <DialogHeader>{dialogTitle}</DialogHeader>
      </DialogHeader>
      <FormErrors errors={errors} className="mb-5" />
      <fieldset className="flex items-center gap-5">
        <label className="w-[150px] text-right" htmlFor="name">
          {t("components.variableForm.name.label")}
        </label>
        <Input
          id="name"
          {...register("name")}
          data-testid="variable-name"
          placeholder={t("components.variableForm.name.placeholder")}
        />
      </fieldset>
      <fieldset className="flex items-center gap-5">
        <label className="w-[150px] text-right" htmlFor="mimetype">
          {t("components.variableForm.mimeType.label")}
        </label>
        <Controller
          control={control}
          name="mimeType"
          render={({ field }) => (
            <MimeTypeSelect
              id="mimetype"
              mimeType={field.value}
              onChange={(newMimeType) => {
                if (newMimeType) {
                  onMimeTypeChange(newMimeType);
                }
              }}
            />
          )}
        />
      </fieldset>
      <FileUpload
        onChange={({ base64String, mimeType }) => {
          const isEditable = isMimeTypeEditable(mimeType);
          setIsEditable(isEditable);
          setValue("data", base64String);
          onMimeTypeChange(mimeType);
        }}
      />
      <Card
        className="grow p-4 pl-0"
        background="weight-1"
        data-testid="variable-editor-card"
      >
        <div className="flex h-[400px]">
          {isEditable ? (
            <Controller
              control={control}
              name="data"
              render={({ field }) => (
                <Editor
                  data-testid="variable-editor"
                  theme={theme ?? undefined}
                  language={editorLanguage}
                  value={decode(field.value)}
                  onChange={(newData) => {
                    if (newData) {
                      field.onChange(encode(newData));
                    }
                  }}
                />
              )}
            />
          ) : (
            <div className="flex grow p-10 text-center">
              <div className="flex items-center justify-center text-sm">
                {t("components.variableForm.noPreview")}
              </div>
            </div>
          )}
        </div>
      </Card>
      {dialogFooter}
    </form>
  );
};

export default VariableForm;
