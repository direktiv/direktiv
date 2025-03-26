import {
  DialogClose,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "~/design/Dialog";
import {
  LayoutSchemaType,
  TextContentSchema,
  TextContentSchemaType,
} from "~/pages/namespace/Explorer/Page/PageEditor/schema";
import { Save, Settings } from "lucide-react";
import { SubmitHandler, useForm } from "react-hook-form";

import Button from "~/design/Button";
import FormErrors from "~/components/FormErrors";
import Input from "~/design/Input";
import { useTranslation } from "react-i18next";
import { zodResolver } from "@hookform/resolvers/zod";

const TextForm = ({
  layout,
  pageElementID,
  onEdit,
}: {
  layout: LayoutSchemaType;
  pageElementID: number;
  onEdit: (content: TextContentSchemaType) => void;
}) => {
  const { t } = useTranslation();

  const onSubmit: SubmitHandler<TextContentSchemaType> = ({ content }) => {
    onEdit({ content });
  };

  const oldContent = layout ? layout[pageElementID]?.content : "nothing";

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<TextContentSchemaType>({
    resolver: zodResolver(TextContentSchema),
    defaultValues: {
      content: oldContent,
    },
  });

  const formId = "edit-text-element";

  return (
    <>
      <DialogHeader>
        <DialogTitle>
          <Settings />
          {t("pages.explorer.page.editor.form.modals.edit.text.title")}
        </DialogTitle>
      </DialogHeader>

      <FormErrors errors={errors} className="mb-5" />
      <form id={formId} onSubmit={handleSubmit(onSubmit)}>
        <div className="my-3">
          <fieldset className="flex items-center gap-5">
            <label className="w-[90px] text-right text-[14px]" htmlFor="text">
              {t("pages.explorer.page.editor.form.modals.edit.text.label")}
            </label>
            <Input
              id="text"
              placeholder={t(
                "pages.explorer.page.editor.form.modals.edit.text.placeholder"
              )}
              {...register("content")}
            />
          </fieldset>
        </div>
        <DialogFooter>
          <DialogClose asChild>
            <Button variant="ghost">
              {t("pages.explorer.page.editor.form.modals.edit.text.cancelBtn")}
            </Button>
          </DialogClose>
          <Button type="submit" variant="outline">
            <Save />
            {t("pages.explorer.page.editor.form.modals.edit.text.saveBtn")}
          </Button>
        </DialogFooter>
      </form>
    </>
  );
};

export default TextForm;
