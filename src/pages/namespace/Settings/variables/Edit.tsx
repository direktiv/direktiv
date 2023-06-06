import {
  DialogClose,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "~/design/Dialog";
import MimeTypeSelect, {
  MimeTypeSchema,
  MimeTypeType,
  mimeTypeToLanguageDict,
} from "./MimeTypeSelect";
import { SubmitHandler, useForm } from "react-hook-form";
import { Trans, useTranslation } from "react-i18next";
import {
  VarFormSchema,
  VarFormSchemaType,
  VarSchemaType,
} from "~/api/variables/schema";
import { useEffect, useState } from "react";

import Button from "~/design/Button";
import Editor from "~/design/Editor";
import FormErrors from "~/componentsNext/FormErrors";
import { Trash } from "lucide-react";
import { useTheme } from "~/util/store/theme";
import { useUpdateVar } from "~/api/variables/mutate/updateVariable";
import { useVarContent } from "~/api/variables/query/useVariableContent";
import { zodResolver } from "@hookform/resolvers/zod";

type EditProps = {
  item: VarSchemaType;
  onSuccess: () => void;
};

// mimeType defaults to text/plain to avoid backend defaulting to
// "text/plain, charset=utf-8", which does not fit the options in
// MimeTypeSelect
const defaultMimeType: MimeTypeType = "text/plain";

const Edit = ({ item, onSuccess }: EditProps) => {
  const { t } = useTranslation();
  const theme = useTheme();

  const varContent = useVarContent(item.name);

  const [body, setBody] = useState<string | undefined>();
  const [mimeType, setMimeType] = useState<MimeTypeType>(defaultMimeType);
  const [isInitialized, setIsInitialized] = useState<boolean>(false);
  const [editorLanguage, setEditorLanguage] = useState<string>("");

  const {
    handleSubmit,
    formState: { errors },
  } = useForm<VarFormSchemaType>({
    resolver: zodResolver(VarFormSchema),
    values: {
      name: item.name,
      content: body ?? "",
      mimeType: mimeType,
    },
  });

  useEffect(() => {
    if (!isInitialized && varContent.isFetched) {
      setBody(varContent.data?.body);

      const contentType = varContent.data?.headers["content-type"];

      const safeParsedContentType = MimeTypeSchema.safeParse(contentType);
      if (!safeParsedContentType.success) {
        return console.error(
          `Unexpected content-type, defaulting to ${defaultMimeType}`
        );
      }
      setMimeType(safeParsedContentType.data);

      setIsInitialized(true);
    }
  }, [varContent, isInitialized]);

  useEffect(() => {
    setEditorLanguage(mimeTypeToLanguageDict[mimeType]);
  }, [mimeType]);

  const { mutate: updateVarMutation } = useUpdateVar({
    onSuccess,
  });

  const onSubmit: SubmitHandler<VarFormSchemaType> = (data) => {
    updateVarMutation(data);
  };

  return (
    <DialogContent>
      {varContent.isFetched && (
        <form
          id="edit-variable"
          onSubmit={handleSubmit(onSubmit)}
          className="flex flex-col space-y-5"
        >
          <DialogHeader>
            <DialogTitle>
              <Trash />
              <Trans
                i18nKey="pages.settings.variables.edit.title"
                values={{ name: item.name }}
              />
            </DialogTitle>
          </DialogHeader>

          <FormErrors errors={errors} />

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

          <fieldset className="flex items-center gap-5">
            <label
              className="w-[150px] text-right text-[15px]"
              htmlFor="template"
            >
              {t("pages.settings.variables.edit.mimeType")}
            </label>
            <MimeTypeSelect mimeType={mimeType} onChange={setMimeType} />
          </fieldset>

          <DialogFooter>
            <DialogClose asChild>
              <Button variant="ghost">
                {t("components.button.label.cancel")}
              </Button>
            </DialogClose>
            <Button
              type="submit"
              data-testid="var-edit-submit"
              variant="destructive"
            >
              {t("components.button.label.save")}
            </Button>
          </DialogFooter>
        </form>
      )}
    </DialogContent>
  );
};

export default Edit;
