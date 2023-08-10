import {
  DialogClose,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "~/design/Dialog";
import Editor, { EditorLanguagesType } from "~/design/Editor";
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
import { Card } from "~/design/Card";
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

// mimeType should always be initialized in the form, to avoid the backend
// setting defaults that may not fit with the options in MimeTypeSelect
const fallbackMimeType: MimeTypeType = "text/plain";

const Edit = ({ item, onSuccess }: EditProps) => {
  const { t } = useTranslation();
  const theme = useTheme();

  const { data, isFetched } = useVarContent(item.name);

  const [body, setBody] = useState<string | undefined>();
  const [mimeType, setMimeType] = useState<MimeTypeType>(fallbackMimeType);
  const [isInitialized, setIsInitialized] = useState<boolean>(false);
  const [editorLanguage, setEditorLanguage] = useState<EditorLanguagesType>(
    mimeTypeToLanguageDict[fallbackMimeType]
  );

  const {
    handleSubmit,
    formState: { errors },
  } = useForm<VarFormSchemaType>({
    resolver: zodResolver(VarFormSchema),
    values: {
      name: item.name,
      content: body ?? "",
      mimeType,
    },
  });

  useEffect(() => {
    if (!isInitialized && isFetched) {
      setBody(data?.body);

      const contentType = data?.headers["content-type"];

      const safeParsedContentType = MimeTypeSchema.safeParse(contentType);
      if (!safeParsedContentType.success) {
        return console.error(
          `Unexpected content-type, defaulting to ${fallbackMimeType}`
        );
      }
      setMimeType(safeParsedContentType.data);
      setEditorLanguage(mimeTypeToLanguageDict[safeParsedContentType.data]);
      setIsInitialized(true);
    }
  }, [data, isFetched, isInitialized]);

  const { mutate: updateVarMutation } = useUpdateVar({
    onSuccess,
  });

  const onSubmit: SubmitHandler<VarFormSchemaType> = (data) => {
    updateVarMutation(data);
  };

  return (
    <DialogContent>
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

        <FormErrors errors={errors} className="mb-5" />

        <fieldset className="flex items-center gap-5">
          <label className="w-[150px] text-right" htmlFor="mimetype">
            {t("pages.settings.variables.edit.mimeType.label")}
          </label>
          <MimeTypeSelect
            id="mimetype"
            loading={!isFetched}
            mimeType={mimeType}
            onChange={setMimeType}
          />
        </fieldset>

        <Card
          className="grow p-4 pl-0"
          background="weight-1"
          data-testid="variable-editor-card"
        >
          <div className="h-[500px]">
            {isFetched && body && (
              <Editor
                value={body}
                onChange={(newData) => {
                  setBody(newData);
                }}
                onMount={(editor) => editor.focus()}
                theme={theme ?? undefined}
                data-testid="variable-editor"
                language={editorLanguage}
              />
            )}
          </div>
        </Card>

        <DialogFooter>
          <DialogClose asChild>
            <Button variant="ghost" data-testid="var-edit-cancel">
              {t("components.button.label.cancel")}
            </Button>
          </DialogClose>
          <Button type="submit" data-testid="var-edit-submit">
            {t("components.button.label.save")}
          </Button>
        </DialogFooter>
      </form>
    </DialogContent>
  );
};

export default Edit;
