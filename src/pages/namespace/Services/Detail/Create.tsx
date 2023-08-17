import {
  DialogClose,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "~/design/Dialog";
import { Diamond, PlusCircle } from "lucide-react";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "~/design/Select";
import {
  ServiceRevisionFormSchema,
  ServiceRevisionFormSchemaType,
} from "~/api/services/schema";
import { SubmitHandler, useForm } from "react-hook-form";

import Button from "~/design/Button";
import FormErrors from "~/componentsNext/FormErrors";
import Input from "~/design/Input";
import { Slider } from "~/design/Slider";
import { useCreateServiceRevision } from "~/api/services/mutate/createRevision";
import { useServiceDetails } from "~/api/services/query/details";
import { useTranslation } from "react-i18next";
import { zodResolver } from "@hookform/resolvers/zod";

const availableSizes = [0, 1, 2] as const;

const CreateRevision = ({
  service,
  defaultValues,
  close,
}: {
  service: string;
  close: () => void;
  defaultValues?: ServiceRevisionFormSchemaType;
}) => {
  const { t } = useTranslation();

  const { data } = useServiceDetails({ service });
  const { mutate: createServiceRevision, isLoading } = useCreateServiceRevision(
    {
      onSuccess: () => {
        close();
      },
    }
  );

  const {
    register,
    handleSubmit,
    watch,
    getValues,
    setValue,
    formState: { errors, isValid, isSubmitted },
  } = useForm<ServiceRevisionFormSchemaType>({
    defaultValues: defaultValues ?? {
      minscale: 0,
      size: 1,
    },
    resolver: zodResolver(ServiceRevisionFormSchema),
  });

  const onSubmit: SubmitHandler<ServiceRevisionFormSchemaType> = ({
    cmd,
    image,
    minscale,
    size,
  }) => {
    createServiceRevision({
      service,
      payload: {
        cmd,
        image,
        minscale,
        size,
      },
    });
  };

  // you can not submit if the form has not changed or if there are any errors and
  // you have already submitted the form (errors will first show up after submit)
  const disableSubmit = isSubmitted && !isValid;

  const formId = `new-service-revision`;

  const maxScale = data?.config.maxscale;
  if (maxScale === undefined) return null;

  return (
    <>
      <DialogHeader>
        <DialogTitle>
          <Diamond /> {t("pages.services.revision.create.title")}
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
            <label className="w-[90px] text-right text-[14px]" htmlFor="image">
              {t("pages.services.revision.create.imageLabel")}
            </label>
            <Input
              id="image"
              placeholder={t("pages.services.revision.create.imagePlaceholder")}
              {...register("image")}
            />
          </fieldset>
          <fieldset className="flex items-center gap-5">
            <label className="w-[90px] text-right text-[14px]" htmlFor="scale">
              {t("pages.services.revision.create.scaleLabel")}
            </label>
            <div className="flex w-full gap-5">
              <Input
                className="w-12"
                readOnly
                value={watch("minscale")}
                disabled
              />
              <Slider
                id="scale"
                step={1}
                min={0}
                max={maxScale}
                value={[watch("minscale") ?? 0]}
                onValueChange={(e) => {
                  const newValue = e[0];
                  newValue !== undefined && setValue("minscale", newValue);
                }}
              />
            </div>
          </fieldset>
          <fieldset className="flex items-center gap-5">
            <label className="w-[90px] text-right text-[14px]" htmlFor="size">
              {t("pages.services.revision.create.sizeLabel")}
            </label>
            <Select
              value={`${getValues("size")}`}
              onValueChange={(value) => setValue("size", parseInt(value))}
            >
              <SelectTrigger variant="outline" className="w-full" id="size">
                <SelectValue
                  placeholder={t(
                    "pages.services.revision.create.sizePlaceholder"
                  )}
                />
              </SelectTrigger>
              <SelectContent>
                {availableSizes.map((size) => (
                  <SelectItem key={size} value={`${size}`}>
                    {t(`pages.services.revision.create.sizeValues.${size}`)}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </fieldset>
          <fieldset className="flex items-center gap-5">
            <label className="w-[90px] text-right text-[14px]" htmlFor="cmd">
              {t("pages.services.revision.create.cmdLabel")}
            </label>
            <Input
              id="cmd"
              placeholder={t("pages.services.revision.create.cmdPlaceholder")}
              {...register("cmd")}
            />
          </fieldset>
        </form>
      </div>
      <DialogFooter>
        <DialogClose asChild>
          <Button variant="ghost">
            {t("pages.services.revision.create.createBtn")}
          </Button>
        </DialogClose>
        <Button
          type="submit"
          disabled={disableSubmit}
          loading={isLoading}
          form={formId}
        >
          {!isLoading && <PlusCircle />}
          {t("pages.services.revision.create.createBtn")}
        </Button>
      </DialogFooter>
    </>
  );
};

export default CreateRevision;
