import {
  DialogClose,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "~/design/Dialog";
import { Folder, PlusCircle } from "lucide-react";
import { SubmitHandler, useForm } from "react-hook-form";

import Button from "~/design/Button";
import { FileNameSchema } from "~/api/files/schema";
import FormErrors from "~/components/FormErrors";
import Input from "~/design/Input";
import { useCreateFile } from "~/api/files/mutate/createFile";
import { useNavigate } from "@tanstack/react-router";
import { useTranslation } from "react-i18next";
import { z } from "zod";
import { zodResolver } from "@hookform/resolvers/zod";

type FormInput = {
  name: string;
};

const NewDirectory = ({
  path,
  close,
  unallowedNames,
}: {
  path?: string;
  close: () => void;
  unallowedNames?: string[];
}) => {
  const { t } = useTranslation();
  const navigate = useNavigate();

  const resolver = zodResolver(
    z.object({
      name: FileNameSchema.and(
        z
          .string()
          .refine((name) => !(unallowedNames ?? []).some((n) => n === name), {
            message: t("pages.explorer.tree.newDirectory.nameAlreadyExists"),
          })
      ),
    })
  );

  const {
    register,
    handleSubmit,
    formState: { isDirty, errors, isValid, isSubmitted },
  } = useForm<FormInput>({
    resolver,
  });

  const { mutate: createDirectory, isPending } = useCreateFile({
    onSuccess: (data) => {
      navigate({
        to: "/n/$namespace/explorer/tree/$",
        from: "/n/$namespace",
        params: { _splat: data.data.path },
      });
      close();
    },
  });

  const onSubmit: SubmitHandler<FormInput> = ({ name }) => {
    createDirectory({ path, payload: { name, type: "directory" } });
  };

  // you can not submit if the form has not changed or if there are any errors and
  // you have already submitted the form (errors will first show up after submit)
  const disableSubmit = !isDirty || (isSubmitted && !isValid);

  const formId = `new-dir-${path}`;
  return (
    <>
      <DialogHeader>
        <DialogTitle>
          <Folder /> {t("pages.explorer.tree.newDirectory.title")}
        </DialogTitle>
      </DialogHeader>

      <div className="my-3">
        <FormErrors errors={errors} className="mb-5" />
        <form id={formId} onSubmit={handleSubmit(onSubmit)}>
          <fieldset className="flex items-center gap-5">
            <label className="w-[90px] text-right text-[14px]" htmlFor="name">
              {t("pages.explorer.tree.newDirectory.nameLabel")}
            </label>
            <Input
              id="name"
              placeholder={t(
                "pages.explorer.tree.newDirectory.folderPlaceholder"
              )}
              {...register("name")}
            />
          </fieldset>
        </form>
      </div>
      <DialogFooter>
        <DialogClose asChild>
          <Button variant="ghost">
            {t("pages.explorer.tree.newDirectory.cancelBtn")}
          </Button>
        </DialogClose>
        <Button
          type="submit"
          disabled={disableSubmit}
          loading={isPending}
          form={formId}
        >
          {!isPending && <PlusCircle />}
          {t("pages.explorer.tree.newDirectory.createBtn")}
        </Button>
      </DialogFooter>
    </>
  );
};

export default NewDirectory;
