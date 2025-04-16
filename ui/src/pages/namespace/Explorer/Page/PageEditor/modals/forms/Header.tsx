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

const HeaderForm = ({
  header,
  onEdit,
  close,
}: {
  header: TextContentSchemaType;
  onEdit: (content: TextContentSchemaType) => void;
  close: () => void;
}) => {
  const { t } = useTranslation();

  const onSubmit: SubmitHandler<TextContentSchemaType> = ({ content }) => {
    const newHeader = {
      type: header.type,
      content,
    };

    onEdit(newHeader);
    close();
  };

  const defaultHeaderElement: TextContentSchemaType = {
    type: "Text",
    content: "This is a Header...",
  };

  const oldHeaderElement =
    header?.type === "Text" ? header : defaultHeaderElement;

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<TextContentSchemaType>({
    resolver: zodResolver(TextContentSchema),
    defaultValues: {
      type: "Text",
      content: oldHeaderElement.content,
    },
  });

  const formId = "edit-header-element";

  return (
    <>
      <DialogHeader>
        <DialogTitle>
          <Settings />
          {t("pages.explorer.page.editor.form.modals.edit.header.title")}
        </DialogTitle>
      </DialogHeader>

      <FormErrors errors={errors} className="mb-5" />
      <form id={formId} onSubmit={handleSubmit(onSubmit)}>
        <div className="my-3">
          <fieldset className="flex items-center gap-5">
            <label className="w-[90px] text-right text-[14px]" htmlFor="text">
              {t("pages.explorer.page.editor.form.modals.edit.header.label")}
            </label>
            <Input
              id="text"
              placeholder={t(
                "pages.explorer.page.editor.form.modals.edit.header.placeholder"
              )}
              {...register("content")}
            />
          </fieldset>
        </div>
        <DialogFooter>
          <DialogClose asChild>
            <Button variant="ghost">
              {t(
                "pages.explorer.page.editor.form.modals.edit.header.cancelBtn"
              )}
            </Button>
          </DialogClose>
          <Button type="submit" variant="outline">
            <Save />
            {t("pages.explorer.page.editor.form.modals.edit.header.saveBtn")}
          </Button>
        </DialogFooter>
      </form>
    </>
  );
};

export default HeaderForm;
