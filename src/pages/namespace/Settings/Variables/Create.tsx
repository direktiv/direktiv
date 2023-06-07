import {
  DialogClose,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "~/design/Dialog";
import { SubmitHandler, useForm } from "react-hook-form";
import { VarFormSchema, VarFormSchemaType } from "~/api/variables/schema";

import Button from "~/design/Button";
import Editor from "~/design/Editor";
import FormErrors from "~/componentsNext/FormErrors";
import Input from "~/design/Input";
import MimeTypeSelect from "./MimeTypeSelect";
import { PlusCircle } from "lucide-react";
import { useState } from "react";
import { useTheme } from "~/util/store/theme";
import { useTranslation } from "react-i18next";
import { useUpdateVar } from "~/api/variables/mutate/updateVariable";
import { zodResolver } from "@hookform/resolvers/zod";

type CreateProps = { onSuccess: () => void };

const Create = ({ onSuccess }: CreateProps) => {
  const { t } = useTranslation();
  const theme = useTheme();

  const [name, setName] = useState<string | undefined>();
  const [body, setBody] = useState<string | undefined>();
  const [mimeType, setMimeType] = useState<string | undefined>();

  const {
    handleSubmit,
    formState: { errors },
  } = useForm<VarFormSchemaType>({
    resolver: zodResolver(VarFormSchema),
    // mimeType defaults to text/plain to avoid backend defaulting to
    // "text/plain, charset=utf-8", which does not fit the options in
    // MimeTypeSelect
    values: {
      name: name ?? "",
      content: body ?? "",
      mimeType: mimeType ?? "text/plain",
    },
  });

  const { mutate: createVarMutation } = useUpdateVar({
    onSuccess,
  });

  const onSubmit: SubmitHandler<VarFormSchemaType> = (data) => {
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
              {t("pages.settings.variables.create.title")}
            </DialogTitle>
          </DialogHeader>
        </DialogHeader>

        <FormErrors errors={errors} />

        <fieldset className="flex items-center gap-5">
          <label className="w-[150px] text-right text-[15px]" htmlFor="name">
            {t("pages.settings.variables.create.name")}
          </label>
          <Input
            data-testid="new-variable-name"
            placeholder="variable-name"
            onChange={(event) => setName(event.target.value)}
          />
        </fieldset>

        <div className="h-[500px]">
          <Editor
            value={body}
            onChange={(newData) => {
              setBody(newData);
            }}
            theme={theme ?? undefined}
            data-testid="variable-editor"
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
            data-testid="variable-create-submit"
            type="submit"
            variant="primary"
          >
            {t("components.button.label.create")}
          </Button>
        </DialogFooter>
      </form>
    </DialogContent>
  );
};

export default Create;
