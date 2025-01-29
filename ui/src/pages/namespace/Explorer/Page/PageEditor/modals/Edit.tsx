import {
  DialogClose,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "~/design/Dialog";
import {
  LayoutSchemaType,
  PageElementContentSchemaType,
} from "~/pages/namespace/Explorer/Page/PageEditor/schema";
import { SubmitHandler, useForm } from "react-hook-form";

import Button from "~/design/Button";
import FormErrors from "~/components/FormErrors";
import Input from "~/design/Input";
import { Save } from "lucide-react";
import { useTranslation } from "react-i18next";

const EditModal = ({
  layout,
  pageElementID,
  close,
  success,
}: {
  layout: LayoutSchemaType;
  pageElementID: number;
  close: () => void;
  success: (newLayout: LayoutSchemaType) => void;
}) => {
  const { t } = useTranslation();

  let isPending = false;

  const onSubmit: SubmitHandler<PageElementContentSchemaType> = ({
    content,
  }) => {
    onEdit(content);
  };

  const onEdit = (content: string) => {
    const newElement = {
      name: oldElement?.name,
      hidden: oldElement?.hidden,
      preview: content,
      content,
    };

    isPending = true;
    const newLayout = [...layout];

    newLayout.splice(pageElementID, 1, newElement);

    success(newLayout);
    isPending = false;
    close();
  };

  const oldElement = layout ? layout[pageElementID] : { content: "nothing" };
  const oldContent = layout ? layout[pageElementID]?.content : "nothing";

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<PageElementContentSchemaType>({
    defaultValues: {
      content: oldContent,
    },
  });

  const formId = "edit-page-element";

  return (
    <>
      <DialogHeader>
        <DialogTitle>
          <Save /> Edit this
        </DialogTitle>
      </DialogHeader>

      <div className="my-3">
        <FormErrors errors={errors} className="mb-5" />
        <form id={formId} onSubmit={handleSubmit(onSubmit)}>
          <fieldset className="flex items-center gap-5">
            <label className="w-[90px] text-right text-[14px]" htmlFor="name">
              {t("pages.explorer.tree.newPage.nameLabel")}
            </label>
            <Input id="name" placeholder="page-name" {...register("content")} />
          </fieldset>

          <DialogFooter>
            <DialogClose asChild>
              <Button variant="ghost">
                {t("pages.explorer.tree.delete.cancelBtn")}
              </Button>
            </DialogClose>
            <Button type="submit" variant="outline" loading={isPending}>
              {!isPending && <Save />}
              Save
            </Button>
          </DialogFooter>
        </form>
      </div>
    </>
  );
};

export default EditModal;
