import {
  DialogClose,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "~/design/Dialog";
import { Diamond, PlusCircle } from "lucide-react";
import {
  GroupFormSchema,
  GroupFormSchemaType,
} from "~/api/enterprise/groups/schema";
import { SubmitHandler, useForm } from "react-hook-form";

import Button from "~/design/Button";
import FormErrors from "~/componentsNext/FormErrors";
import Input from "~/design/Input";
import { useCreateGroup } from "~/api/enterprise/groups/mutation/create";
import { useTranslation } from "react-i18next";
import { zodResolver } from "@hookform/resolvers/zod";

const CreateGroup = ({
  close,
  unallowedNames,
}: {
  close: () => void;
  unallowedNames?: string[];
}) => {
  const { t } = useTranslation();
  const { mutate: createGroup, isLoading } = useCreateGroup({
    onSuccess: () => {
      close();
    },
  });

  const {
    register,
    handleSubmit,
    formState: { isDirty, errors, isValid, isSubmitted },
  } = useForm<GroupFormSchemaType>({
    defaultValues: {
      group: "",
      description: "",
      permissions: [],
    },
    resolver: zodResolver(
      GroupFormSchema.refine(
        (x) => !(unallowedNames ?? []).some((n) => n === x.group),
        {
          path: ["group"],
          message: t("pages.permissions.groups.create.group.alreadyExist"),
        }
      )
    ),
  });

  const onSubmit: SubmitHandler<GroupFormSchemaType> = (params) => {
    createGroup(params);
  };

  // you can not submit if the form has not changed or if there are any errors and
  // you have already submitted the form (errors will first show up after submit)
  const disableSubmit = !isDirty || (isSubmitted && !isValid);

  const formId = `new-group`;

  return (
    <>
      <DialogHeader>
        <DialogTitle>
          <Diamond /> {t("pages.permissions.groups.create.title")}
        </DialogTitle>
      </DialogHeader>

      <div className="my-3">
        <FormErrors errors={errors} className="mb-5" />
        <form
          id={formId}
          onSubmit={handleSubmit(onSubmit)}
          className="flex flex-col space-y-5"
        >
          <fieldset className="flex items-center gap-5">
            <label className="w-[90px] text-right text-[14px]" htmlFor="group">
              {t("pages.permissions.groups.create.group.label")}
            </label>
            <Input
              id="group"
              placeholder={t(
                "pages.permissions.groups.create.group.placeholder"
              )}
              {...register("group")}
            />
          </fieldset>
          <fieldset className="flex items-center gap-5">
            <label
              className="w-[90px] text-right text-[14px]"
              htmlFor="description"
            >
              {t("pages.permissions.groups.create.description.label")}
            </label>
            <Input
              id="description"
              placeholder={t(
                "pages.permissions.groups.create.description.placeholder"
              )}
              {...register("description")}
            />
          </fieldset>
          <fieldset className="flex items-center gap-5">
            <label
              className="w-[90px] text-right text-[14px]"
              htmlFor="permissions"
            >
              {t("pages.permissions.groups.create.permissions")}
            </label>
            <div>Permissions will go here</div>
          </fieldset>
        </form>
      </div>
      <DialogFooter>
        <DialogClose asChild>
          <Button variant="ghost">
            {t("pages.services.create.createBtn")}
          </Button>
        </DialogClose>
        <Button
          type="submit"
          disabled={disableSubmit}
          loading={isLoading}
          form={formId}
        >
          {!isLoading && <PlusCircle />}
          {t("pages.permissions.groups.create.createBtn")}
        </Button>
      </DialogFooter>
    </>
  );
};

export default CreateGroup;
