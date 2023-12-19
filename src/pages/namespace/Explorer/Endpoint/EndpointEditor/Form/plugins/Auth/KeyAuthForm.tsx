import { FC, FormEvent } from "react";
import FormErrors, { errorsType } from "~/componentsNext/FormErrors";
import {
  KeyAuthFormSchema,
  KeyAuthFormSchemaType,
} from "../../../schema/plugins/auth/keyAuth";

import Button from "~/design/Button";
import { Checkbox } from "~/design/Checkbox";
import { DialogFooter } from "~/design/Dialog";
import Input from "~/design/Input";
import { treatEmptyStringAsUndefined } from "../utils";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";

type OptionalConfig = Partial<KeyAuthFormSchemaType["configuration"]>;

const predfinedConfig: OptionalConfig = {
  add_groups_header: true,
  add_tags_header: true,
  add_username_header: true,
};

type FormProps = {
  defaultConfig?: OptionalConfig;
  onSubmit: (data: KeyAuthFormSchemaType) => void;
};

export const KeyAuthForm: FC<FormProps> = ({ defaultConfig, onSubmit }) => {
  const {
    handleSubmit,
    setValue,
    getValues,
    register,
    formState: { errors },
  } = useForm<KeyAuthFormSchemaType>({
    resolver: zodResolver(KeyAuthFormSchema),
    defaultValues: {
      type: "key-auth",
      configuration: {
        ...predfinedConfig,
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
            add username header
          </label>
          <Checkbox
            defaultChecked={getValues("configuration.add_username_header")}
            onCheckedChange={(value) => {
              if (typeof value === "boolean") {
                setValue("configuration.add_username_header", value);
              }
            }}
          />
        </fieldset>
        <fieldset className="flex items-center gap-5">
          <label className="w-[170px] overflow-hidden text-right text-sm">
            add tags header
          </label>
          <Checkbox
            defaultChecked={getValues("configuration.add_tags_header")}
            onCheckedChange={(value) => {
              if (typeof value === "boolean") {
                setValue("configuration.add_tags_header", value);
              }
            }}
          />
        </fieldset>
        <fieldset className="flex items-center gap-5">
          <label className="w-[170px] overflow-hidden text-right text-sm">
            add groups header
          </label>
          <Checkbox
            defaultChecked={getValues("configuration.add_groups_header")}
            onCheckedChange={(value) => {
              if (typeof value === "boolean") {
                setValue("configuration.add_groups_header", value);
              }
            }}
          />
        </fieldset>
        <fieldset className="flex items-center gap-5">
          <label className="w-[170px] overflow-hidden text-right text-sm">
            key name (optional)
          </label>
          <div>
            <Input
              {...register("configuration.key_name", {
                setValueAs: treatEmptyStringAsUndefined,
              })}
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
