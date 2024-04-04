import { Controller, SubmitHandler, useForm } from "react-hook-form";
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
import {
  VarFormCreateSchema,
  VarFormCreateSchemaType,
  VarFormUpdateSchemaType,
} from "~/api/variables/schema";
import { decode, encode } from "js-base64";

import Button from "~/design/Button";
import { Card } from "~/design/Card";
import FileUpload from "../components/FileUpload";
import FormErrors from "~/components/FormErrors";
import Input from "~/design/Input";
import { PlusCircle } from "lucide-react";
import { useCreateVar } from "~/api/variables/mutate/create";
import { useState } from "react";
import { useTheme } from "~/util/store/theme";
import { useTranslation } from "react-i18next";
import { zodResolver } from "@hookform/resolvers/zod";

type CreateProps = { onSuccess: () => void };

const defaultMimeType = "application/json";

const Create = ({ onSuccess }: CreateProps) => {
  const { mutate: createVar } = useCreateVar({
    onSuccess,
  });

  return (
    <DialogContent>
      <Generic
        onMutate={createVar}
        defaultValues={{
          name: "",
          data: "",
          mimeType: defaultMimeType,
        }}
      />
    </DialogContent>
  );
};

type GenericProps = {
  defaultValues: VarFormCreateSchemaType & VarFormUpdateSchemaType;
  onMutate: (data: VarFormCreateSchemaType & VarFormUpdateSchemaType) => void;
};

export const Generic = ({ onMutate, defaultValues }: GenericProps) => {
  const { t } = useTranslation();
  const theme = useTheme();

  const [isEditable, setIsEditable] = useState(true);

  const [editorLanguage, setEditorLanguage] = useState<EditorLanguagesType>(
    mimeTypeToLanguageDict[defaultMimeType]
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

  // TODO: move this component
  // TODO: add header and footer slot s
  return (
    <form
      // TODO: return ref (toid="edit-variable" and "create-variable" shoudl set)
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
          const parsedMimetype = EditorMimeTypeSchema.safeParse(mimeType);
          const isEditable = parsedMimetype.success;
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

export default Create;
