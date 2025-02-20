import {
  DialogClose,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "~/design/Dialog";
import { Diamond, PlusCircle } from "lucide-react";
import {
  RoleFormSchema,
  RoleFormSchemaType,
} from "~/api/enterprise/roles/schema";
import { SubmitHandler, useForm } from "react-hook-form";

import Button from "~/design/Button";
import RoleForm from "./Form";
import { useCreateRole } from "~/api/enterprise/roles/mutation/create";
import { useTranslation } from "react-i18next";
import { zodResolver } from "@hookform/resolvers/zod";

type CreateRoleProps = {
  close: () => void;
  unallowedNames?: string[];
};

const CreateRole = ({ close, unallowedNames }: CreateRoleProps) => {
  const { t } = useTranslation();
  const { mutate: createRole, isPending } = useCreateRole({
    onSuccess: () => {
      close();
    },
  });

  const resolver = zodResolver(
    RoleFormSchema.refine(
      (x) => !(unallowedNames ?? []).some((n) => n === x.name),
      {
        path: ["name"],
        message: t("pages.permissions.roles.form.name.alreadyExist"),
      }
    )
  );

  const form = useForm<RoleFormSchemaType>({
    defaultValues: {
      name: "",
      description: "",
      oidcGroups: [],
      permissions: [],
    },
    resolver,
  });

  const {
    formState: { isDirty, isValid, isSubmitted },
  } = form;

  const onSubmit: SubmitHandler<RoleFormSchemaType> = (params) => {
    createRole(params);
  };

  // you can not submit if the form has not changed or if there are any errors and
  // you have already submitted the form (errors will first show up after submit)
  const disableSubmit = !isDirty || (isSubmitted && !isValid);

  const formId = `new-role`;

  return (
    <>
      <DialogHeader>
        <DialogTitle>
          <Diamond /> {t("pages.permissions.roles.create.title")}
        </DialogTitle>
      </DialogHeader>
      <RoleForm form={form} onSubmit={onSubmit} formId={formId} />
      <DialogFooter>
        <DialogClose asChild>
          <Button variant="ghost">
            {t("pages.permissions.roles.create.cancelBtn")}
          </Button>
        </DialogClose>
        <Button
          type="submit"
          disabled={disableSubmit}
          loading={isPending}
          form={formId}
        >
          {!isPending && <PlusCircle />}
          {t("pages.permissions.roles.create.saveBtn")}
        </Button>
      </DialogFooter>
    </>
  );
};

export default CreateRole;
