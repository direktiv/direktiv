import { Controller, SubmitHandler, useForm } from "react-hook-form";
import {
  DialogClose,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "~/design/Dialog";
import Editor, { EditorLanguagesType } from "~/design/Editor";
import {
  VarFormCreateSchema,
  VarFormCreateSchemaType,
  VarFormUpdateSchemaType,
} from "~/api/variables/schema";
import { decode, encode } from "js-base64";
import {
  getLanguageFromMimeType,
  isMimeTypeEditable,
  mimeTypeToLanguageDict,
} from "../utils";

import Button from "~/design/Button";
import { Card } from "~/design/Card";
import FileUpload from "../../components/FileUpload";
import FormErrors from "~/components/FormErrors";
import Input from "~/design/Input";
import MimeTypeSelect from "../MimeTypeSelect/";
import { PlusCircle } from "lucide-react";
import { useState } from "react";
import { useTheme } from "~/util/store/theme";
import { useTranslation } from "react-i18next";
import { zodResolver } from "@hookform/resolvers/zod";

const defaultMimeType = "application/json";

type VariableForm = {
  defaultValues: VarFormCreateSchemaType & VarFormUpdateSchemaType;
  onMutate: (data: VarFormCreateSchemaType & VarFormUpdateSchemaType) => void;
};

export const VariableForm = ({ onMutate, defaultValues }: VariableForm) => {
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
  } = useForm<VarFormCreateSchemaType & VarFormUpdateSchemaType>({
    // TODO: the resolver must be dynamic
    resolver: zodResolver(VarFormCreateSchema),
    defaultValues,
  });

  const onMimeTypeChange = (value: string) => {
    setValue("mimeType", value);
    const editorLanguage = getLanguageFromMimeType(value);
    if (editorLanguage) {
      setEditorLanguage(editorLanguage);
    }
  };

  const onSubmit: SubmitHandler<VarFormCreateSchemaType> = (data) => {
    onMutate(data);
  };

  // TODO: add header and footer slot
  // TODO: clean up translation keys
  return (
    <form onSubmit={handleSubmit(onSubmit)} className="flex flex-col space-y-5">
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
        <Controller
          control={control}
          name="mimeType"
          render={({ field }) => (
            <MimeTypeSelect
              id="mimetype"
              mimeType={field.value}
              onChange={onMimeTypeChange}
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
        data-testid="variable-create-card"
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
                {t("pages.settings.variables.create.noPreview")}
              </div>
            </div>
          )}
        </div>
      </Card>
      <DialogFooter>
        <DialogClose asChild>
          <Button variant="ghost">{t("components.button.label.cancel")}</Button>
        </DialogClose>
        <Button data-testid="variable-create-submit" type="submit">
          {t("components.button.label.create")}
        </Button>
      </DialogFooter>
    </form>
  );
};

export default VariableForm;
