import { FC, FormEvent } from "react";
import FormErrors, { errorsType } from "~/componentsNext/FormErrors";
import {
  GithubWebhookAuthFormSchema,
  GithubWebhookAuthFormSchemaType,
} from "../../../schema/plugins/auth/githubWebhookAuth";

import Button from "~/design/Button";
import { DialogFooter } from "~/design/Dialog";
import Input from "~/design/Input";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";

type OptionalConfig = Partial<GithubWebhookAuthFormSchemaType["configuration"]>;

type FormProps = {
  defaultConfig?: OptionalConfig;
  onSubmit: (data: GithubWebhookAuthFormSchemaType) => void;
};

export const GithubWebhookAuthForm: FC<FormProps> = ({
  defaultConfig,
  onSubmit,
}) => {
  const {
    handleSubmit,
    register,
    formState: { errors },
  } = useForm<GithubWebhookAuthFormSchemaType>({
    resolver: zodResolver(GithubWebhookAuthFormSchema),
    defaultValues: {
      type: "github-webhook-auth",
      configuration: {
        ...defaultConfig,
      },
    },
  });

  const submitForm = (e: FormEvent<HTMLFormElement>) => {
    e.stopPropagation(); // prevent the parent form from submitting
    handleSubmit(onSubmit)(e);
  };

  return (
    <form onSubmit={submitForm}>
      {errors?.configuration && (
        <FormErrors
          errors={errors?.configuration as errorsType}
          className="mb-5"
        />
      )}

      <div className="my-3 flex flex-col gap-y-5">
        <fieldset className="flex items-center gap-5">
          <label className="w-[170px] overflow-hidden text-right text-sm">
            secret
          </label>
          <div>
            <Input
              {...register("configuration.secret")}
              placeholder="name of key"
            />
          </div>
        </fieldset>
      </div>
      <DialogFooter>
        <Button type="submit">Save</Button>
      </DialogFooter>
    </form>
  );
};
