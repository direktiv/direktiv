import {
  InstantResponseFormSchema,
  InstantResponseFormSchemaType,
} from "../../../schema/plugins/target/InstantResponse";
import { SubmitHandler, useForm } from "react-hook-form";

import Button from "~/design/Button";
import { DialogFooter } from "~/design/Dialog";
import { FC } from "react";
import FormErrors from "~/componentsNext/FormErrors";
import Input from "~/design/Input";
import { Textarea } from "~/design/TextArea";
import { zodResolver } from "@hookform/resolvers/zod";

type FormProps = {
  defaultConfig?: InstantResponseFormSchemaType["configuration"];
  onSubmit: (data: InstantResponseFormSchemaType) => void;
};

export const InstantResponseForm: FC<FormProps> = ({
  defaultConfig,
  onSubmit,
}) => {
  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<InstantResponseFormSchemaType>({
    resolver: zodResolver(InstantResponseFormSchema),
    defaultValues: {
      type: "instant-response",
      configuration: {
        ...defaultConfig,
      },
    },
  });

  return (
    <form
      onSubmit={(e) => {
        e.stopPropagation();
        handleSubmit(onSubmit)(e);
      }}
    >
      {errors?.configuration && (
        <FormErrors errors={errors?.configuration} className="mb-5" />
      )}

      <div className="my-3 flex flex-col gap-y-5">
        <fieldset className="flex items-center gap-5">
          <label className="w-[150px] overflow-hidden text-right text-sm">
            content_type
          </label>
          <Input
            {...register("configuration.content_type")}
            placeholder="application/json"
          />
        </fieldset>
        <fieldset className="flex items-center gap-5">
          <label className="w-[150px] overflow-hidden text-right text-sm">
            status_code
          </label>
          <Input
            {...register("configuration.status_code", {
              valueAsNumber: true,
            })}
            type="number"
            placeholder="200"
          />
        </fieldset>
        <fieldset className="flex items-center gap-5">
          <label className="w-[150px] overflow-hidden text-right text-sm">
            status_message
          </label>
          <Textarea {...register("configuration.status_message")} />
        </fieldset>
      </div>
      <DialogFooter>
        <Button type="submit" variant="primary">
          Save
        </Button>
      </DialogFooter>
    </form>
  );
};
