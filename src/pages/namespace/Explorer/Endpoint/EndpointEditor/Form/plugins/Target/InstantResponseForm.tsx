import { Controller, useForm } from "react-hook-form";
import { FC, FormEvent } from "react";
import FormErrors, { errorsType } from "~/componentsNext/FormErrors";
import {
  InstantResponseFormSchema,
  InstantResponseFormSchemaType,
} from "../../../schema/plugins/target/InstantResponse";

import Button from "~/design/Button";
import { Card } from "~/design/Card";
import { DialogFooter } from "~/design/Dialog";
import Editor from "~/design/Editor";
import Input from "~/design/Input";
import { useTheme } from "~/util/store/theme";
import { zodResolver } from "@hookform/resolvers/zod";

type FormProps = {
  defaultConfig?: Partial<InstantResponseFormSchemaType["configuration"]>;
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
    control,
  } = useForm<InstantResponseFormSchemaType>({
    resolver: zodResolver(InstantResponseFormSchema),
    defaultValues: {
      type: "instant-response",
      configuration: {
        ...defaultConfig,
      },
    },
  });

  const submitForm = (e: FormEvent<HTMLFormElement>) => {
    e.stopPropagation(); // prevent the parent form from submitting
    handleSubmit(onSubmit)(e);
  };

  const theme = useTheme();

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
          <label className="w-[150px] overflow-hidden text-right text-sm">
            status code
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
            content type
          </label>
          <Input
            {...register("configuration.content_type")}
            placeholder="application/json"
          />
        </fieldset>
        <fieldset className="flex items-center gap-5">
          <label className="w-[150px] overflow-hidden text-right text-sm">
            status message
          </label>
          <Card className="h-[200px] w-full p-5" background="weight-1" noShadow>
            <Controller
              control={control}
              name="configuration.status_message"
              render={({ field }) => (
                <Editor
                  theme={theme ?? undefined}
                  language="plaintext"
                  value={field.value}
                  onChange={field.onChange}
                />
              )}
            />
          </Card>
        </fieldset>
      </div>
      <DialogFooter>
        <Button type="submit">Save</Button>
      </DialogFooter>
    </form>
  );
};
