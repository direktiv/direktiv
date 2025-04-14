import {
  DialogClose,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "~/design/Dialog";
import { Save, Settings } from "lucide-react";
import { SubmitHandler, useForm } from "react-hook-form";
import {
  TextContentSchema,
  TextContentSchemaType,
} from "~/pages/namespace/Explorer/Page/PageEditor/schema";

import Button from "~/design/Button";
import FormErrors from "~/components/FormErrors";
import Input from "~/design/Input";
import { useTranslation } from "react-i18next";
import { zodResolver } from "@hookform/resolvers/zod";

const FooterForm = ({
  footer,
  onEdit,
  close,
}: {
  footer: TextContentSchemaType;
  onEdit: (content: TextContentSchemaType) => void;
  close: () => void;
}) => {
  const { t } = useTranslation();

  const onSubmit: SubmitHandler<TextContentSchemaType> = ({ content }) => {
    const newFooter = {
      type: footer.type,
      content,
    };

    onEdit(newFooter);
    close();
  };

  const defaultFooterElement: TextContentSchemaType = {
    type: "Text",
    content: "This is a Footer...",
  };

  const oldFooterElement: TextContentSchemaType =
    footer?.type === "Text" ? footer : defaultFooterElement;

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<TextContentSchemaType>({
    resolver: zodResolver(TextContentSchema),
    defaultValues: {
      type: "Text",
      content: oldFooterElement.content,
    },
  });

  const formId = "edit-footer-element";

  return (
    <>
      <DialogHeader>
        <DialogTitle>
          <Settings />
          {t("pages.explorer.page.editor.form.modals.edit.footer.title")}
        </DialogTitle>
      </DialogHeader>

      <FormErrors errors={errors} className="mb-5" />
      <form id={formId} onSubmit={handleSubmit(onSubmit)}>
        <div className="my-3">
          <fieldset className="flex items-center gap-5">
            <label className="w-[90px] text-right text-[14px]" htmlFor="text">
              {t("pages.explorer.page.editor.form.modals.edit.footer.label")}
            </label>
            <Input
              id="text"
              placeholder={t(
                "pages.explorer.page.editor.form.modals.edit.footer.placeholder"
              )}
              {...register("content")}
            />
          </fieldset>
        </div>
        <DialogFooter>
          <DialogClose asChild>
            <Button variant="ghost">
              {t(
                "pages.explorer.page.editor.form.modals.edit.footer.cancelBtn"
              )}
            </Button>
          </DialogClose>
          <Button type="submit" variant="outline">
            <Save />
            {t("pages.explorer.page.editor.form.modals.edit.footer.saveBtn")}
          </Button>
        </DialogFooter>
      </form>
    </>
  );
};

export default FooterForm;
